package promptactions

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

type PromptExitApp struct {
	c      *internal.Context
	logger logger.Logger
}

func NewPromptExitApp(c *internal.Context) internal.PromptAction {
	p := PromptExitApp{c: c}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptExitApp) GetName() string {
	return "PromptExitApp"
}

func (p PromptExitApp) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptExitApp) GetActionKeys() []string {
	return []string{"exit", ":q"}
}

func (p PromptExitApp) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptExitApp) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString("Exit the application.")
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :q", "Y", markdown)))
	return builder.String()
}

func (p PromptExitApp) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{}
}

func (p PromptExitApp) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	return []prompt.Suggest{}, nil
}

func (p PromptExitApp) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		p.logger.Info("Application exit!")
		os.Exit(0)
	}
	return nil
}

func (p PromptExitApp) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
