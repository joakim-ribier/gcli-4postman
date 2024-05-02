package promptactions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

type PromptSelectEnv struct {
	c      *internal.Context
	logger logger.Logger
}

func NewPromptSelectEnv(c *internal.Context) internal.PromptAction {
	p := PromptSelectEnv{c: c}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptSelectEnv) GetName() string {
	return "PromptSelectEnv"
}

func (p PromptSelectEnv) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptSelectEnv) GetActionKeys() []string {
	return []string{"env", ":e"}
}

func (p PromptSelectEnv) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptSelectEnv) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString("Select the collection execution environment.")
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :e localhost", "Y", markdown)))
	return builder.String()
}

func (p PromptSelectEnv) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{}
}

func (p PromptSelectEnv) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		if len(in) > 1 {
			selectedEnv := postman.NewEnv()
			for _, env := range p.c.Envs {
				if env.GetName() == in[1] {
					selectedEnv = env
					break
				}
			}
			p.c.Env = &selectedEnv
			if selectedEnv.Name != "" {
				p.c.Print("INFO", "switch on {%s} env", selectedEnv.GetName())
				return nil
			}
		}
		p.c.Print("WARN", "select an environnment from the suggestions")
	}
	return nil
}

func (p PromptSelectEnv) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) {
		return []prompt.Suggest{}, nil
	}
	var suggests = []prompt.Suggest{
		{Text: "none", Description: "No environment"},
	}
	for _, env := range p.c.Envs {
		suggests = append(suggests, prompt.Suggest{Text: env.GetName(), Description: ""})
	}
	return suggests, nil
}

func (p PromptSelectEnv) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
