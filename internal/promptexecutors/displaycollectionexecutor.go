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
	output func(string)
}

// NewDisplayCollectionExecutor builds executor for display collection action.
func NewDisplayCollectionExecutor(c internal.Context, logger logger.Logger, output func(string)) DisplayCollectionExecutor {
	return DisplayCollectionExecutor{
		c:      c,
		logger: logger,
		output: output,
	}
}

// Display displays on the out the current selected collection.
func (dc DisplayCollectionExecutor) Display(filterBy string) {
	execs.NewDisplayCollectionExec(dc.output).Display(*dc.c.Collection, filterBy)
}
