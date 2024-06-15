package promptactions

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors/execs"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

var (
	httpMethodS = prompt.Suggest{Text: "-m", Description: "filter requests by method (GET, POST...)"}
	httpUrlS    = prompt.Suggest{Text: "-u", Description: "find a request to execute"}
	historyS    = prompt.Suggest{Text: "-history", Description: "find a previous request"}
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
	return promptexecutors.NewExecuteRequestExecutor(*p.c, p.logger)
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
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor(`# :h -u GET../users/findByName {{id}} "Joakim Ribier" {{x-organisation}} "GitHub" --pretty`, "Y", markdown)))
	builder.WriteString(fmt.Sprintf("\n_to not send the header parameter, add %s after the {{x-organisation}}_", prettyprint.FormatTextWithColor(`--delete`, "Y", markdown)))
	return builder.String()
}

func (p PromptExecuteRequest) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{
		{Value: httpMethodS.Text, Description: httpMethodS.Description},
		{Value: httpUrlS.Text, Description: httpUrlS.Description},
		{Value: historyS.Text, Description: fmt.Sprintf("%s\n%s", historyS.Description, prettyprint.FormatTextWithColor("# :h -history GET../users/findByName#1 --pretty", "Y", markdown))},
		{Value: "--search {pattern}", Description: fmt.Sprintf("find data in the response using %s awesome lib\nmore details on %s", prettyprint.FormatTextWithColor("tidwall/gjson", "Y", markdown), prettyprint.FormatTextWithColor("https://github.com/tidwall/gjson", "B", markdown))},
		{Value: "--pretty", Description: "display a beautiful HTTP json response"},
		{Value: "--full", Description: fmt.Sprintf("display the full response (not limited to %s characters)", prettyprint.FormatTextWithColor(strconv.Itoa(internal.HTTP_BODY_SIZE_LIMIT), "Y", markdown))},
		{Value: "--save {/path/file.json}", Description: "save the full body response in a file"},
		{Value: "--reset", Description: "reset the collection history requests"},
	}
}

func (p PromptExecuteRequest) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) || p.c.Collection == nil {
		return []prompt.Suggest{}, nil
	}

	if slices.Contains(in, historyS.Text) {
		return slicesutil.TransformT[postman.CollectionHistoryItemLight, prompt.Suggest](p.c.CollectionHistoryRequests.SortByExecutedAt(), func(f postman.CollectionHistoryItemLight) (*prompt.Suggest, error) {
			return &prompt.Suggest{Text: f.GetSuggestText(), Description: f.GetSuggestDescription()}, nil
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
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		if p.c.Collection == nil {
			p.c.Print("WARN", "select a collection from the suggestions")
			return nil
		}
		if slices.Contains(in, historyS.Text) && len(in) > 1 {
			if slices.Contains(in, "--reset") {
				p.GetPromptExecutor().(promptexecutors.ExecuteRequestExecutor).ResetHistory()
				p.c.CollectionHistoryRequests = postman.CollectionHistoryItemsLight{}
			} else {
				if len(in) > 2 {
					if historyItemLight := p.c.CollectionHistoryRequests.FindByLabel(in[2]); historyItemLight != nil {
						historyItemPath := p.c.GetCollectionHistoryPathFolder() + "/" + historyItemLight.BuildNameFile()
						historyItem, err := ioutil.Load[postman.CollectionHistoryItem](historyItemPath, internal.SECRET)
						if err != nil {
							p.c.Log.Error(err, "data cannot be loaded", "ressource", historyItemPath)
						} else {
							execs.NewDisplayBodyResponseExec(p.logger, prettyprint.Print).Display(in, &historyItem)
						}
					}
				}
			}
		} else {
			value := slicesutil.FindNextEl(in, httpUrlS.Text)
			if item := p.c.Collection.FindItemByLabel(value); item != nil {
				if response, err := p.GetPromptExecutor().(promptexecutors.ExecuteRequestExecutor).Call(in, *item); err != nil {
					p.c.Print("ERROR", stringsutil.NewStringS(err.Error()).ReplaceAll("%7B", "{").ReplaceAll("%7D", "}").S())
				} else {
					// refresh the context
					p.c.CollectionHistoryRequests = append(p.c.CollectionHistoryRequests, response.ToLight())

					internal.HistoriseCommand(*p.c, strings.Join(in, " "))
					p.GetPromptExecutor().(promptexecutors.ExecuteRequestExecutor).HistoriseNewCollectionItem(*response)

					execs.NewDisplayBodyResponseExec(p.logger, prettyprint.Print).Display(in, response)

					if path := slicesutil.FindNextEl(in, "--save"); path != "" {
						if err := iosutil.Write(response.Data, path); err != nil {
							p.c.Log.Error(err, "data cannot be writed")
							p.c.Print("ERROR", "the body's response cannot be saved...")
						}
					}
				}
			} else {
				p.c.Print("WARN", "request {%s} does not exist in the collection", value)
				return nil
			}
		}
	}
	return nil
}

func (p PromptExecuteRequest) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
