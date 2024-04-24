package promptactions

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type PromptDisplayCollection struct {
	c *internal.Context
}

func NewPromptDisplayCollection(c *internal.Context) internal.PromptAction {
	return PromptDisplayCollection{
		c: c,
	}
}

func (p PromptDisplayCollection) GetName() string {
	return "PromptDisplayCollection"
}

func (p PromptDisplayCollection) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptDisplayCollection) GetActionKeys() []string {
	return []string{"display", ":d"}
}

func (p PromptDisplayCollection) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptDisplayCollection) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{
		{Value: "--search {pattern}", Description: "API requests full-text search"},
	}
}

func (p PromptDisplayCollection) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString("Display API requests of the current loaded collection.")
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :d --search users", "Y", markdown)))
	return builder.String()
}

func (p PromptDisplayCollection) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	return []prompt.Suggest{}, nil
}

func (p PromptDisplayCollection) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRight(p, in, internal.APP_MODE) {
		if p.c.Collection == nil {
			p.c.Print("WARN", "select a collection before to display it")
			return nil
		}
		value := slicesutil.FindNextEl(in, "--search")
		prettyprint.PrintCollection(p.c.Collection.SortByName(), value)
	}
	return nil
}

func (p PromptDisplayCollection) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
