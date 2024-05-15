package internal

import (
	"slices"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

// PromptAction interface which defines how to implement a new prompt action
type PromptAction interface {
	// GetName returns the name of the action
	GetName() string

	// GetDescription returns the action's description (for "help" cmd)
	GetDescription(markdown bool) string

	// GetOptions returns the action's options can be used (for "help" cmd)
	GetOptions(markdown bool) []Option

	// GetActionKeys returns the command(s) (load, :l) which should used to use this prompt action
	GetActionKeys() []string

	// GetParamKeys returns the param(s) (-update-readme, -secure-mode) which should used to use this prompt action
	GetParamKeys() []ParamWithRole

	// GetPromptExecutor builds the action's executor used to execute commands
	GetPromptExecutor() PromptExecutor

	// PromptSuggest returns the action's suggests available
	PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error)

	// PromptExecutor executes action's command and can return a callback
	PromptExecutor(in []string) *PromptCallback

	// PromptCallback executes command after callback - often used to have user confirmation
	PromptCallback(in []string, actions []PromptAction, args ...any)
}

type Option struct {
	Value       string
	Description string
}

type ParamWithRole struct {
	Value string
	Roles []string
}

// HasRightToExecute checks if the prompt action {p} can be executed based on user input and current mode.
func HasRightToExecute(p PromptAction, in []string, role string) bool {
	if slices.Contains(p.GetActionKeys(), in[0]) {
		if len(in) == 1 || p.GetParamKeys() == nil {
			return true
		}
		return slicesutil.ExistT(p.GetParamKeys(), func(pwr ParamWithRole) bool {
			return slicesutil.Exist(in, pwr.Value) && (slicesutil.Exist(pwr.Roles, APP_MODE) || len(pwr.Roles) == 0)
		})
	}
	return false
}

// FindPromptActionExecutor finds the prompt action executor {T}.
func FindPromptActionExecutor[T PromptExecutor](actions []PromptAction) *T {
	if found := slicesutil.FindT[PromptAction](actions, func(pa PromptAction) bool {
		_, is := pa.GetPromptExecutor().(T)
		return is
	}); found != nil {
		p := *found
		t := p.GetPromptExecutor().(T)
		return &t
	}
	return nil
}

// AddCMDHistory historises the command {cmd} (writes data on the disk).
func AddCMDHistory(c Context, cmd string) {
	if cmd != "" {
		histories := slicesutil.AddOrReplaceT[CMDHistory](c.CMDsHistory, NewCMDHistory(cmd), func(c CMDHistory) bool {
			return c.CMD == cmd
		})
		ioutil.Write[CMDHistories](histories, c.GetCMDHistoryPath(), SECRET)
	}
}
