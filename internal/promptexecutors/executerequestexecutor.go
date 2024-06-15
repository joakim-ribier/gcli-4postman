package promptexecutors

import (
	"os"

	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/httputil"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

// Executor for execute request action.
type ExecuteRequestExecutor struct {
	c      internal.Context
	logger logger.Logger
}

// NewSettingsExecutor builds executor for execute request action.
func NewExecuteRequestExecutor(c internal.Context, logger logger.Logger) ExecuteRequestExecutor {
	return ExecuteRequestExecutor{
		c:      c,
		logger: logger,
	}
}

// Call calls the API {item} request.
func (er ExecuteRequestExecutor) Call(in []string, item postman.Item) (*postman.CollectionHistoryItem, error) {
	var params []postman.Param = slicesutil.TransformT[string, postman.Param](item.GetParams(), func(param string) (*postman.Param, error) {
		if value := slicesutil.FindNextEl(in, param); value != "" {
			return &postman.Param{Key: param, Value: value}, nil
		} else {
			return nil, nil
		}
	})

	response, err := httputil.Call(item, er.c.Env, params)
	if err != nil {
		er.logger.Error(err, "request cannot be called", "resource", item.GetLabel(), "url", item.Request.Url.Raw)
		return nil, err
	}

	var itemResponse = postman.NewCollectionHistoryItem(
		len(er.c.CollectionHistoryRequests)+1, item,
		response.Status, response.TimeInMillis,
		response.Body, response.ContentLength,
		er.c.Env, params)

	return &itemResponse, nil
}

// ResetHistory resets the history of the current selected collection.
func (er ExecuteRequestExecutor) ResetHistory() {
	if err := os.RemoveAll(er.c.GetCollectionHistoryPathFolder()); err != nil {
		er.c.Print("ERROR", "unable to remove collection history %s", er.c.GetCollectionHistoryPathFolder())
	}
}

// HistoriseCollection writes on the disk the collection items response.
func (er ExecuteRequestExecutor) HistoriseNewCollectionItem(item postman.CollectionHistoryItem) bool {
	if _, err := os.Open(er.c.GetCollectionHistoryPathFolder()); err != nil {
		// assume that folder does not exist and try to create it
		if err = os.Mkdir(er.c.GetCollectionHistoryPathFolder(), os.ModePerm); err != nil {
			er.logger.Error(err, "collection history folder cannot be created", "resource", er.c.GetCollectionHistoryPathFolder())
			return false
		}
	}
	historyItemPath := er.c.GetCollectionHistoryPathFolder() + "/" + item.ToLight().BuildNameFile()
	if err := ioutil.Write[postman.CollectionHistoryItem](item, historyItemPath, internal.SECRET); err != nil {
		er.logger.Error(err, "collection history cannot be written", "resource", historyItemPath)
		return false
	}
	return true
}
