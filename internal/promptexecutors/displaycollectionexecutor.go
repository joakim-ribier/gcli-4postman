package promptexecutors

import (
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors/execs"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

// Executor for display collection action.
type DisplayCollectionExecutor struct {
	c      internal.Context
	logger logger.Logger
}

// NewDisplayCollectionExecutor builds executor for display collection action.
func NewDisplayCollectionExecutor(c internal.Context, logger logger.Logger) DisplayCollectionExecutor {
	return DisplayCollectionExecutor{
		c:      c,
		logger: logger,
	}
}

// Display displays on the out the current selected collection.
func (dc DisplayCollectionExecutor) Display(filterBy string) {
	execs.NewDisplayCollectionExec().Display(*dc.c.Collection, filterBy, func(render string) {
		if len(dc.c.Collection.Items) > 0 {
			print("["+dc.c.Collection.Info.Name+"'s collection]\n", render, "")
		}
	})
}
