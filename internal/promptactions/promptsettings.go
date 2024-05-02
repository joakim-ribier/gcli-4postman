package promptactions

import (
	"fmt"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

const (
	updateReadmeKeyParam = "-update-readme"
	secureModeKeyParam   = "-secure-mode"
	enableOptionParam    = "enable"
	disableOptionParam   = "disable"
	secretOptionParam    = "--secret"
)

type PromptSettings struct {
	updateReadmeSuggest prompt.Suggest
	secureModeSuggest   prompt.Suggest

	c      *internal.Context
	logger logger.Logger
}

func NewPromptSettings(c *internal.Context) internal.PromptAction {
	p := PromptSettings{
		updateReadmeSuggest: prompt.Suggest{Text: updateReadmeKeyParam, Description: "update the README from help documentation"},
		secureModeSuggest:   prompt.Suggest{Text: secureModeKeyParam, Description: "enable or disable the secure mode"},
		c:                   c,
	}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptSettings) GetName() string {
	return "PromptSettings"
}

func (p PromptSettings) GetPromptExecutor() internal.PromptExecutor {
	return promptexecutors.NewSettingsExecutor(*p.c, p.logger)
}

func (p PromptSettings) GetActionKeys() []string {
	return []string{"settings", ":s"}
}

func (p PromptSettings) GetParamKeys() []internal.ParamWithRole {
	return []internal.ParamWithRole{
		{Value: updateReadmeKeyParam, Roles: []string{"admin"}},
		{Value: secureModeKeyParam, Roles: []string{"admin"}},
	}
}

func (p PromptSettings) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Available settings (or actions) on %s", prettyprint.FormatTextWithColor("CLI-4Postman", "Y", markdown)))
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :s -secure-mode enable --secret {secret}", "Y", markdown)))
	return builder.String()
}

func (p PromptSettings) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{
		{Value: p.updateReadmeSuggest.Text, Description: fmt.Sprintf("%s %s", p.updateReadmeSuggest.Description, prettyprint.FormatTextWithColor("// --mode admin", "G", markdown))},
		{Value: fmt.Sprintf("%s %s", p.secureModeSuggest.Text, enableOptionParam), Description: fmt.Sprintf("enable secure mode by adding (or update) a new secret %s %s", prettyprint.FormatTextWithColor("--secret {secret}", "Y", markdown), prettyprint.FormatTextWithColor("// --mode admin", "G", markdown))},
		{Value: fmt.Sprintf("%s %s", p.secureModeSuggest.Text, disableOptionParam), Description: fmt.Sprintf("disable secure mode %s %s", prettyprint.FormatTextWithColor("!! NOT RECOMMENDED !!", "R", markdown), prettyprint.FormatTextWithColor("// --mode admin", "G", markdown))},
	}
}

func (p PromptSettings) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) || internal.APP_MODE != "admin" {
		return []prompt.Suggest{}, nil
	}

	if slices.Contains(in, p.secureModeSuggest.Text) {
		return p.getSecureModeSuggest(in)
	}

	return []prompt.Suggest{
		p.updateReadmeSuggest,
		p.secureModeSuggest}, nil
}

func (p PromptSettings) getSecureModeSuggest(in []string) ([]prompt.Suggest, error) {
	if v := slicesutil.FindNextEl(in, p.secureModeSuggest.Text); v != enableOptionParam && v != disableOptionParam {
		return []prompt.Suggest{
			{Text: enableOptionParam, Description: "enable secure mode by adding a new secret (--secret {secret})"},
			{Text: disableOptionParam, Description: "not recommended..."},
		}, nil
	}
	if slices.Contains(in, enableOptionParam) && !slices.Contains(in, secretOptionParam) {
		return []prompt.Suggest{
			{Text: secretOptionParam, Description: "add a new strong secret to encrypt the data"},
		}, nil
	}
	return []prompt.Suggest{}, nil
}

func (p PromptSettings) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		if slicesutil.Exist(in, p.updateReadmeSuggest.Text) {
			return internal.NewPromptCallback(
				"Update README (Yes / No)",
				[]internal.PromptSuggestCallback{
					internal.NewPromptSuggestCallback("Yes", "Update the README.md from HELP docs"),
					internal.NewPromptSuggestCallback("No", "...")},
				p, p.updateReadmeSuggest.Text)
		}
		if slicesutil.Exist(in, p.secureModeSuggest.Text) {
			if slicesutil.Exist(in, enableOptionParam) {
				if newSecret := slicesutil.FindNextEl(in, secretOptionParam); newSecret != "" {
					if r := p.GetPromptExecutor().(promptexecutors.SettingsExecutor).EnableSecureMode(newSecret); r {
						internal.SECRET = newSecret
					}
				} else {
					p.c.Print("WARN", "select a new {secret} to continue...")
				}
				return nil
			}
			if slicesutil.Exist(in, disableOptionParam) {
				if r := p.GetPromptExecutor().(promptexecutors.SettingsExecutor).DisableSecureMode(); r {
					internal.SECRET = ""
				}
				return nil
			}
		}
		p.c.Print("WARN", "select an available option to continue...")
	}
	return nil
}

func (p PromptSettings) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	if len(args) > 0 && args[0].(string) == p.updateReadmeSuggest.Text {
		if slicesutil.Exist(in, "Yes") {
			p.GetPromptExecutor().(promptexecutors.SettingsExecutor).UpdateReadme(
				promptexecutors.NewHelpExecutor(actions).Generate())
		}
	}
}
