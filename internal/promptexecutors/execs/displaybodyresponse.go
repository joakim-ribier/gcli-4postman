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
			result := gjson.GetBytes(itemResponse.Data, v)
			d.output(prettyprint.SPrintJson(
				[]byte(stringsutil.OrElse(result.String(), `"no result"`)),
				slicesutil.Exist(in, "--pretty")))
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
