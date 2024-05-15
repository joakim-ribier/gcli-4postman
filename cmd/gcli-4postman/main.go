package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/promptactions"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

const logo = `
       ______ __     ____            __ __   ____                __
      / ____// /    /  _/           / // /  / __ \ ____   _____ / /_ ____ ___   ____ _ ____
     / /    / /     / /   ______   / // /_ / /_/ // __ \ / ___// __// __ '__ \ / __ '// __ \
    / /___ / /___ _/ /   /_____/  /__  __// ____// /_/ /(__  )/ /_ / / / / / // /_/ // / / /
    \____//_____//___/              /_/  /_/     \____//____/ \__//_/ /_/ /_/ \__,_//_/ /_/
                                            https://github.com/joakim-ribier/gcli-4postman
`

const Title = "Prompt for Postman"

var LivePrefix string

var (
	context        *internal.Context
	actions        []internal.PromptAction
	promptCallback *internal.PromptCallback

	log logger.Logger
)

func main() {
	args := slicesutil.ToMap(os.Args[1:])
	if arg, ok := args[string("--home")]; ok {
		internal.GCLI_4POSTMAN_HOME = arg
	}
	if arg, ok := args[string("--mode")]; ok {
		internal.APP_MODE = arg
	}
	if arg, ok := args[string("--secret")]; ok {
		internal.SECRET = arg
	}
	if arg, ok := args[string("--log")]; ok {
		internal.FILE_LOG = arg
	}

	log = logger.NewLogger(stringsutil.NewStringS(internal.FILE_LOG).OrElse("gcli-4postman.log"))

	print("", logo)
	print("", "MODE=%s SECURE=%s", prettyprint.FormatTextWithColor(strings.ToUpper(internal.APP_MODE), "INFO", false), prettyprint.FormatTextWithColor(secureMode(), "INFO", false))
	print("", "$CLI-4Postman %s", prettyprint.FormatTextWithColor(internal.GCLI_4POSTMAN_HOME, "INFO", false))
	print("", " ")

	log.Info(logo, "GCLI_4POSTMAN_HOME", internal.GCLI_4POSTMAN_HOME, "mode", internal.APP_MODE, "secure", secureMode())

	if _, err := os.ReadDir(internal.GCLI_4POSTMAN_HOME); err != nil {
		log.Error(err, "$GCLI_4POSTMAN_HOME is not a folder", "GCLI_4POSTMAN_HOME", internal.GCLI_4POSTMAN_HOME)

		print("WARN", "$CLI-4Postman is not a folder, please select a correct one before continuing...\n")
		print("INFO", "declare the %s var in your home", prettyprint.FormatTextWithColor("$CLI-4Postman", "Y", false))
		print("INFO", "or directly by adding argument %s\n", prettyprint.FormatTextWithColor("./go-cli-4Postman --home {folder}", "Y", false))
		return
	}

	print("INFO", "Type %s for available commands...", prettyprint.FormatTextWithColor("help", "INFO", false))

	context = internal.NewContext(log, print)

	if v, err := ioutil.Load[internal.CMDHistories](context.GetCMDHistoryPath(), internal.SECRET); err == nil {
		context.CMDsHistory = v
	}

	actions = append(actions,
		promptactions.NewPromptLoadCollection(context),
		promptactions.NewPromptSelectEnv(context),
		promptactions.NewPromptExecuteRequest(context),
		promptactions.NewPromptDisplayCollection(context),
		promptactions.NewPromptPostman(context),
		promptactions.NewPromptSettings(context),
		promptactions.NewPromptExitApp(context),
	)

	actions = append(actions, promptactions.NewPromptHelp(actions))

	p := prompt.New(
		promptExecutor,
		promptCompleter,
		prompt.OptionTitle(Title),
		prompt.OptionLivePrefix(func() (string, bool) {
			return LivePrefix, true
		}),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.DarkGreen),
		prompt.OptionSelectedSuggestionBGColor(prompt.Black),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionDescriptionBGColor(prompt.DarkGray),
		prompt.OptionDescriptionTextColor(prompt.White),
		prompt.OptionSelectedDescriptionBGColor(prompt.White),
		prompt.OptionSelectedDescriptionTextColor(prompt.Black),
		prompt.OptionHistory(context.CMDsHistory.GetName()),
	)

	LivePrefix = promptRefreshPrefix(false)

	p.Run()
}

func promptCompleter(d prompt.Document) []prompt.Suggest {
	getArgs := func(input string) (string, string) {
		and := strings.Split(input, "&&")
		or := strings.Split(input, "||")
		arg0 := input
		arg1 := input
		if len(and) > 1 {
			arg0 = and[0]
			arg1 = and[1]
		}
		if len(or) > 1 {
			arg0 = or[0]
			arg1 = or[1]
		}
		return arg0, arg1
	}

	input := strings.ToLower(d.GetWordBeforeCursor())
	arg0, arg1 := getArgs(input)

	return slicesutil.FilterT[prompt.Suggest](promptSuggest(d), func(s prompt.Suggest) bool {
		text := strings.ToLower(s.Text)
		description := strings.ToLower(s.Description)
		if strings.Contains(input, "&&") {
			return strings.Contains(text, arg0) && strings.Contains(description, arg1)
		}
		if strings.Contains(input, "||") {
			return strings.Contains(text, arg0) || strings.Contains(description, arg1)
		}
		return strings.Contains(text, arg0)
	})
}

func promptRefreshPrefix(debug bool) string {
	if promptCallback != nil && !debug {
		return fmt.Sprintf(">> %s : ", promptCallback.Text)
	} else {
		LivePrefix := ""
		if context != nil && context.CollectionName != "" {
			LivePrefix = "$ " + context.CollectionName
		} else {
			LivePrefix = "$ {collection}"
		}

		if context != nil && context.GetEnvName() != "" {
			LivePrefix += " >> {" + context.GetEnvName() + "}"
		} else {
			LivePrefix += " >> {no-env}"
		}

		LivePrefix += " " + "# "

		return LivePrefix
	}
}

func promptSuggest(d prompt.Document) []prompt.Suggest {
	if promptCallback != nil {
		return promptCallback.GetSuggests()
	}

	tab := strings.Split(strings.TrimSpace(d.TextBeforeCursor()), " ")

	var suggests []prompt.Suggest
	for _, promptAction := range actions {
		if v, err := promptAction.PromptSuggest(tab, d); err != nil {
			log.Error(err, fmt.Sprintf("cannot get suggestions for %s action", promptAction.GetName()))
			print("ERROR", fmt.Sprint(err), nil)
			break
		} else {
			if len(v) > 0 {
				suggests = v
				break
			}
		}
	}

	if len(suggests) > 0 {
		return suggests
	} else {
		if len(tab) < 2 && len(slicesutil.FilterT(actions, func(pa internal.PromptAction) bool {
			return slicesutil.Exist(pa.GetActionKeys(), tab[0])
		})) == 0 {
			suggests := []prompt.Suggest{
				{Text: "load", Description: "[:l]oad a collection"},
				{Text: "postman", Description: "synchronize data from [:P]ostman account (API Key)"},
				{Text: "env", Description: "select the execution [:e]nvironnment"},
				{Text: "display", Description: "[:d]isplay the selected collection"},
				{Text: "http", Description: "execute an [:h]ttp API request"},
				{Text: "help", Description: "show help"},
				{Text: "settings", Description: "application's [:s]ettings"},
				{Text: "exit", Description: "[:q]uit the application (Bye)"},
			}

			return slicesutil.FilterT(suggests, func(s prompt.Suggest) bool {
				return internal.APP_MODE == "admin" || (s.Text != "postman" && s.Text != "settings")
			})

		} else {
			return []prompt.Suggest{}
		}
	}
}

func promptExecutor(in string) {
	in = strings.TrimSpace(in)
	tab := strings.Split(in, " ")

	if promptCallback != nil && promptCallback.IsAvailableAnswer(in) {
		promptCallback.Callback(tab, actions)
		promptCallback = nil
	} else {
		for _, promptAction := range actions {
			if callback := promptAction.PromptExecutor(slicesutil.FilterByNonEmpty(tab)); callback != nil {
				promptCallback = callback
			}
		}
	}

	LivePrefix = promptRefreshPrefix(false)
}

func print(level, text string, args ...any) {
	prefix, suffix := "", ""
	switch strings.ToLower(level) {
	case "error":
		if text != "" {
			prefix = ">>>> "
		}
	case "":
		prefix, suffix = "", ""
	case "warn":
		prefix = ">> "
	case "info":
		level = ""
		fallthrough
	default:
		prefix = "> "
	}

	if strings.ToLower(level) != "debug" && text != "" {
		if len(args) > 0 {
			prettyprint.Print(prettyprint.SPrintInColor(fmt.Sprintf(prefix+text+suffix, args...), level, false))
		} else {
			prettyprint.Print(prettyprint.SPrintInColor(prefix+text+suffix, level, false))
		}
	}
}

func secureMode() string {
	return genericsutil.When(internal.SECRET, stringsutil.IsEmpty, "DISABLE", "ENABLE")
}
