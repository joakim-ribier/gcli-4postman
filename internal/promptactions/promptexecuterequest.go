package promptactions

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/httputil"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

var (
	httpMethodS   = prompt.Suggest{Text: "-m", Description: "filter requests by method (GET, POST, ...)"}
	httpUrlS      = prompt.Suggest{Text: "-u", Description: "find a request to execute"}
	historyS      = prompt.Suggest{Text: "-history", Description: "find a history request"}
	historyResetS = prompt.Suggest{Text: "-history --reset", Description: "reset the collection history requests"}
)

type PromptExecuteRequest struct {
	c      *internal.Context
	logger logger.Logger
}

func NewPromptExecuteRequest(c *internal.Context) internal.PromptAction {
	p := PromptExecuteRequest{c: c}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptExecuteRequest) GetName() string {
	return "PromptExecuteRequest"
}

func (p PromptExecuteRequest) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptExecuteRequest) GetActionKeys() []string {
	return []string{"http", ":h"}
}

func (p PromptExecuteRequest) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptExecuteRequest) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Execute a request from the collection - %s", prettyprint.FormatTextWithColor("!! BE CAREFUL TO THE ENVIRONMENT !!", "R", markdown)))
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :h -m GET -u GET:users --pretty", "Y", markdown)))
	return builder.String()
}

func (p PromptExecuteRequest) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{
		{Value: httpMethodS.Text, Description: httpMethodS.Description},
		{Value: httpUrlS.Text, Description: httpUrlS.Description},
		{Value: historyS.Text, Description: fmt.Sprintf("%s\n%s", historyS.Description, prettyprint.FormatTextWithColor("# :h -history GET:users --pretty", "Y", markdown))},
		{Value: historyResetS.Text, Description: historyResetS.Description},
		{Value: "--search {pattern}", Description: "XPath query to extract data from the response"},
		{Value: "--pretty", Description: "display a beautiful HTTP json response"},
	}
}

func (p PromptExecuteRequest) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) || p.c.Collection == nil {
		return []prompt.Suggest{}, nil
	}

	if slices.Contains(in, historyS.Text) {
		return slicesutil.TransformT[postman.CollectionHistoryItem, prompt.Suggest](p.c.CollectionHistoryRequests.SortByExecutedAt(), func(f postman.CollectionHistoryItem) (*prompt.Suggest, error) {
			return &prompt.Suggest{Text: f.Item.GetLabel(), Description: f.GetSuggestDescription()}, nil
		}), nil
	}

	if !slices.Contains(in, httpMethodS.Text) && !slices.Contains(in, httpUrlS.Text) {
		return []prompt.Suggest{httpMethodS, httpUrlS, historyS}, nil
	}

	if len(in) > 1 {
		selectedMethod := slicesutil.FindNextEl(in, httpMethodS.Text)

		currentSelectedOption := func() string {
			options := []string{httpMethodS.Text, httpUrlS.Text}
			currentOption := slicesutil.FindLastOccurrenceIn(in, options)

			currentText := in[len(in)-1]
			isAnOptionValue := slices.Contains(options, currentText)

			if (!isAnOptionValue && d.GetWordBeforeCursor() != "") ||
				(currentOption == currentText && currentText != d.GetWordBeforeCursorWithSpace()) {
				return strings.TrimSpace(currentOption)
			}

			return ""
		}

		switch currentSelectedOption() {
		case httpMethodS.Text:
			return slicesutil.TransformT[string, prompt.Suggest](p.c.Collection.GetMethods(), func(s string) (*prompt.Suggest, error) {
				return &prompt.Suggest{Text: s, Description: ""}, nil
			}), nil
		case httpUrlS.Text:
			items := p.c.Collection.FindByMethod(selectedMethod).SortByLabel()
			return slicesutil.TransformT[postman.Item, prompt.Suggest](items, func(i postman.Item) (*prompt.Suggest, error) {
				return &prompt.Suggest{Text: i.GetLabel(), Description: i.Request.Url.GetLongPath()}, nil
			}), nil
		}
	}

	if !slices.Contains(in, httpUrlS.Text) {
		suggests := []prompt.Suggest{httpUrlS}
		if slices.Contains(in, httpMethodS.Text) {
			return suggests, nil
		} else {
			return append(suggests, httpMethodS), nil
		}
	}

	item := p.c.Collection.FindItemByLabel(slicesutil.FindNextEl(in, httpUrlS.Text))
	if item == nil {
		return []prompt.Suggest{}, nil
	}

	return slicesutil.TransformT[string, prompt.Suggest](item.GetParams(), func(param string) (*prompt.Suggest, error) {
		if !slices.Contains(in, param) {
			return &prompt.Suggest{Text: param, Description: ""}, nil
		} else {
			return nil, errors.New("param already exists")
		}
	}), nil
}

func (p PromptExecuteRequest) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRight(p, in, internal.APP_MODE) {
		if slices.Contains(in, historyS.Text) && len(in) > 1 {
			if slices.Contains(in, "--reset") {
				p.resetHistory()
			} else {
				if itemResponse := p.c.CollectionHistoryRequests.FindByLabel(in[2]); itemResponse != nil {
					internal.DisplayBodyResponse(in, itemResponse, p.logger)
				}
			}
		} else {
			p.executeRequest(in)
		}
	}
	return nil
}

func (p PromptExecuteRequest) executeRequest(in []string) {
	if p.c.Collection == nil {
		p.c.Print("WARN", "select a collection from the suggestions")
		return
	}

	selectedUrl := slicesutil.FindNextEl(in, httpUrlS.Text)

	if item := p.c.Collection.FindItemByLabel(selectedUrl); item != nil {
		var params []postman.Param = slicesutil.TransformT[string, postman.Param](item.GetParams(), func(param string) (*postman.Param, error) {
			if value := slicesutil.FindNextEl(in, param); value != "" {
				return &postman.Param{Key: param, Value: value}, nil
			} else {
				return nil, nil
			}
		})

		response, err := httputil.Call(item, p.c.Env, params)
		if err != nil {
			p.logger.Error(err, "request cannot be called", "resource", item.GetLabel(), "url", item.Request.Url.Raw)
			p.c.Print("ERROR", stringsutil.NewStringS(err.Error()).ReplaceAll("%7B", "{").ReplaceAll("%7D", "}").S())
			return
		}

		internal.AddCMDHistory(*p.c, strings.Join(in, " "))
		itemResponse := postman.NewCollectionHistoryItem(item, response.Status, response.Body, p.c.Env, params)
		p.historise(itemResponse)
		internal.DisplayBodyResponse(in, &itemResponse, p.logger)
	} else {
		p.c.Print("WARN", "request {%s} does not exist in the collection", selectedUrl)
		return
	}
}

func (p PromptExecuteRequest) historise(r postman.CollectionHistoryItem) {
	p.c.AddCollectionHistoryRequest(r)
	p.writeCollectionHistorise()
}

func (p PromptExecuteRequest) writeCollectionHistorise() {
	if err := ioutil.Write[postman.CollectionHistoryItems](p.c.CollectionHistoryRequests, p.c.GetCollectionHistoryPath(), internal.SECRET); err != nil {
		p.logger.Error(err, "collection history cannot be written", "resource", p.c.GetCollectionHistoryPath())
		p.c.Print("ERROR", "unable to write collection history...")
	} else {
		p.logger.Info("write history collection", "resource", p.c.GetCollectionHistoryPath(), "size", len(p.c.CollectionHistoryRequests))
	}
}

func (p PromptExecuteRequest) resetHistory() {
	p.c.CollectionHistoryRequests = postman.CollectionHistoryItems{}
	p.writeCollectionHistorise()
}

func (p PromptExecuteRequest) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
