package execs

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/tidwall/gjson"
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
func (d DisplayBodyResponseExec) displayBodyResponse(in []string, historyItem postman.CollectionHistoryItem) {
	d.output(fmt.Sprintf(
		"METHOD=%s STATUS=%s\nURL=%s",
		prettyprint.FormatTextWithColor(historyItem.Item.Request.Method, "G", false),
		prettyprint.FormatTextWithColor(historyItem.Status, "G", false),
		prettyprint.FormatTextWithColor(historyItem.Item.Request.Url.Get(historyItem.Env, historyItem.Params), "G", false),
	))
	d.output("BODY=")
	d.output(prettyprint.SPrintJson([]byte(historyItem.Item.Request.Body.Get(historyItem.Env, historyItem.Params)), true))
	d.output("____")
	d.output(fmt.Sprintf(
		"EXECUTED_AT=%s SIZE=%s TIME_(ms)=%s\nRESPONSE=",
		prettyprint.FormatTextWithColor(historyItem.ExecutedAt.Format("2006-01-02 15:04:05"), "G", false),
		prettyprint.FormatTextWithColor(strconv.FormatInt(historyItem.ContentLength, 10), "G", false),
		prettyprint.FormatTextWithColor(strconv.FormatInt(historyItem.TimeInMillis, 10), "G", false),
	))
	if historyItem.GetSize() > 0 {
		if v := slicesutil.FindNextEl(in, "--search"); v != "" {
			data := gjson.GetBytes(historyItem.Data, v).String()
			d.output(prettyprint.SPrintJson(
				[]byte(stringsutil.OrElse(data, `"no result"`)),
				slicesutil.Exist(in, "--pretty")))
		} else {
			maxLimit := genericsutil.OrElse(
				-1, func() bool { return slicesutil.Exist(in, "--full") },
				internal.HTTP_BODY_SIZE_LIMIT)

			// display the response on output
			d.output(prettyprint.SPrintJson(historyItem.GetData(maxLimit), slicesutil.Exist(in, "--pretty")))

			truncated := len(historyItem.Data) > internal.HTTP_BODY_SIZE_LIMIT
			if truncated && !slicesutil.Exist(in, "--full") {
				d.output("\n...payload too big, the body has been truncated...")
				d.output(fmt.Sprintf("(add option %s to display the full body or export it with %s)",
					prettyprint.FormatTextWithColor("--full", "Y", false),
					prettyprint.FormatTextWithColor("--save {path}", "Y", false)))
			}
		}
	}
}
