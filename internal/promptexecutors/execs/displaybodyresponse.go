package execs

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/antchfx/jsonquery"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type DisplayBodyResponseExec struct {
	logger logger.Logger
	output func(string)
}

func NewDisplayBodyResponseExec(logger logger.Logger, output func(string)) DisplayBodyResponseExec {
	return DisplayBodyResponseExec{
		logger: logger,
		output: output,
	}
}

// Display displays the body response of the collection item.
func (d DisplayBodyResponseExec) Display(in []string, itemResponse *postman.CollectionHistoryItem) {
	if itemResponse != nil {
		d.displayBodyResponse(in, *itemResponse)
	} else {
		d.logger.Error(errors.New("itemResponse nil"), "Cannot display body response!")
	}
}

// DisplayBodyResponse displays body of the response
func (d DisplayBodyResponseExec) displayBodyResponse(in []string, itemResponse postman.CollectionHistoryItem) {
	d.output(fmt.Sprintf(
		"METHOD=%s STATUS=%s\nURL=%s",
		prettyprint.FormatTextWithColor(itemResponse.Item.Request.Method, "G", false),
		prettyprint.FormatTextWithColor(itemResponse.Status, "G", false),
		prettyprint.FormatTextWithColor(itemResponse.Item.Request.Url.Get(itemResponse.Env, itemResponse.Params), "G", false),
	))
	d.output("BODY=")
	d.output(prettyprint.SPrintJson([]byte(itemResponse.Item.Request.Body.Get(itemResponse.Env, itemResponse.Params)), true))
	d.output("____")
	d.output(fmt.Sprintf(
		"EXECUTED_AT=%s SIZE=%s TIME_(ms)=%s\nRESPONSE=",
		prettyprint.FormatTextWithColor(itemResponse.ExecutedAt.Format("2006-01-02 15:04:05"), "G", false),
		prettyprint.FormatTextWithColor(strconv.FormatInt(itemResponse.ContentLength, 10), "G", false),
		prettyprint.FormatTextWithColor(strconv.FormatInt(itemResponse.TimeInMillis, 10), "G", false),
	))
	if len(itemResponse.Body) > 0 {
		if v := slicesutil.FindNextEl(in, "--search"); v != "" {
			doc, err := jsonquery.Parse(bytes.NewReader(itemResponse.Data))
			if err != nil {
				d.logger.Error(err, "`body` cannot be parsed")
			}
			nodes, err := jsonquery.QueryAll(doc, v)
			if err != nil {
				d.logger.Error(err, "`expr` cannot be parsed")
				d.output(prettyprint.SPrintInColor(err.Error(), "error", false))
			} else {
				for i, node := range nodes {
					navigator := jsonquery.CreateXPathNavigator(node)
					d.output(prettyprint.SPrintInColor(fmt.Sprintf("[PATH] %s", buildPathUntilRootParent(node, "")), "", false))
					d.output(prettyprint.SPrintJson([]byte(navigator.Value()), slicesutil.Exist(in, "--pretty")))
					if i+1 < len(nodes) {
						d.output("")
					}
				}
			}
		} else {
			maxLimit := genericsutil.OrElse(-1, func() bool { return slicesutil.Exist(in, "--full") }, internal.HTTP_BODY_SIZE_LIMIT)
			d.output(prettyprint.SPrintJson(itemResponse.GetBody(maxLimit), slicesutil.Exist(in, "--pretty")))
			if itemResponse.Trunc && maxLimit != -1 && itemResponse.Data != nil {
				d.output("\n...payload too big, the body has been truncated...")
				d.output(fmt.Sprintf("(add option %s to display the full body or export it with %s)",
					prettyprint.FormatTextWithColor("--full", "Y", false),
					prettyprint.FormatTextWithColor("--save {path}", "Y", false)))
			}
		}
	}
}

func buildPathUntilRootParent(parent *jsonquery.Node, path string) string {
	if parent == nil {
		return path
	}
	if path == "" {
		return buildPathUntilRootParent(parent.Parent, parent.Data)
	} else {
		return buildPathUntilRootParent(parent.Parent, parent.Data+"/"+path)
	}
}
