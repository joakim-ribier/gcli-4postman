package promptactions

import (
	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors"
)

type PromptHelp struct {
	actions []internal.PromptAction
}

func NewPromptHelp(actions []internal.PromptAction) internal.PromptAction {
	return PromptHelp{actions: actions}
}

func (p PromptHelp) GetName() string {
	return "PromptHelp"
}

func (p PromptHelp) GetPromptExecutor() internal.PromptExecutor {
	return promptexecutors.NewHelpExecutor(p.actions)
}

func (p PromptHelp) GetActionKeys() []string {
	return []string{"help"}
}

func (p PromptHelp) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptHelp) GetDescription(markdown bool) string {
	return ""
}

func (p PromptHelp) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{}
}

func (p PromptHelp) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	return []prompt.Suggest{}, nil
}

func (p PromptHelp) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		p.GetPromptExecutor().(promptexecutors.HelpExecutor).Display()
	}
	return nil
}

func (p PromptHelp) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
