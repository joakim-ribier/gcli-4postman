package promptexecutors

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

// Executor for help action.
type HelpExecutor struct {
	actions []internal.PromptAction
}

// NewHelpExecutor builds executor for help action.
func NewHelpExecutor(actions []internal.PromptAction) HelpExecutor {
	return HelpExecutor{
		actions: actions,
	}
}

func (h HelpExecutor) Generate() string {
	return h.help(true)
}

func (h HelpExecutor) Display() {
	fmt.Println()
	fmt.Println(h.help(false))
	fmt.Println()
}

func (h HelpExecutor) help(markdown bool) string {
	sb := strings.Builder{}
	sb.Reset()

	// CMD part
	sb.WriteString("## CMD\n\n")
	sb.WriteString(fmt.Sprintf("The %s is handled by a %s which tries to get the correct suggestions.",
		prettyprint.FormatTextWithColor("CLI", "Y", markdown),
		prettyprint.FormatTextWithColor("prompt completer", "Y", markdown)))
	sb.WriteString(fmt.Sprintf("\nTo optimize the list of suggestions from the prompt completer, it is possible to combine %s and %s operators.",
		prettyprint.FormatTextWithColor("&&", "Y", markdown),
		prettyprint.FormatTextWithColor("||", "Y", markdown)))
	sb.WriteString(fmt.Sprintf("\n* %s only matches with the left side of the suggestion %s.",
		prettyprint.FormatTextWithColor("{a single value}", "Y", markdown),
		prettyprint.FormatTextWithColor("{Suggest.Text}", "Y", markdown)))
	sb.WriteString(fmt.Sprintf("\n* %s matches the left value with the %s %s the right side with the %s.",
		prettyprint.FormatTextWithColor("{value}&&{value}", "Y", markdown),
		prettyprint.FormatTextWithColor("{Suggest.Text}", "Y", markdown),
		prettyprint.FormatTextWithColor("AND", "Y", markdown),
		prettyprint.FormatTextWithColor("{Suggest.Description}", "Y", markdown)))
	sb.WriteString(fmt.Sprintf("\n* %s matches the left value with the %s %s the right side with the %s.",
		prettyprint.FormatTextWithColor("{value}||{value}", "Y", markdown),
		prettyprint.FormatTextWithColor("{Suggest.Text}", "Y", markdown),
		prettyprint.FormatTextWithColor("OR", "Y", markdown),
		prettyprint.FormatTextWithColor("{Suggest.Description}", "Y", markdown)))
	sb.WriteString("\n\n")
	if markdown {
		actionsHelp := h.getActionsHelp(markdown).RenderMarkdown()
		read_lines := strings.Split(actionsHelp, "\n")
		action := ""
		for i, line := range read_lines {
			if i > 1 {
				columns := strings.Split(line, "|")
				if len(slicesutil.FilterByNonEmpty(columns)) > 2 {
					if columns[1] != action {
						sb.WriteString(line)
						action = columns[1]
					} else {
						columns[1] = ""
						columns[2] = ""
						sb.WriteString(strings.Join(columns[:], " | "))
					}
					sb.WriteString("\n")
				}
			} else {
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}
	} else {
		sb.WriteString(h.getActionsHelp(markdown).Render())
	}
	sb.WriteString("\n\n")

	moreDetailsW := table.NewWriter()
	moreDetailsW.Style().Options = table.OptionsNoBordersAndSeparators

	moreDetailsSB := strings.Builder{}
	moreDetailsSB.WriteString(fmt.Sprintf("is a 'XPath query' (%s) format to filter the response,\n", prettyprint.FormatTextWithColor("//*/status[text()='ERROR']", "Y", markdown)))
	moreDetailsSB.WriteString(fmt.Sprintf("go to %s to see all possibilities.", prettyprint.FormatTextWithColor("https://github.com/antchfx/jsonquery", "B", markdown)))
	moreDetailsW.AppendRows([]table.Row{
		{
			prettyprint.FormatTextWithColor("--search {pattern}", "Y", markdown),
			moreDetailsSB.String(),
		},
	})

	sb.WriteString(moreDetailsW.Render())

	return sb.String()
}

func (h HelpExecutor) getActionsHelp(markdown bool) table.Writer {
	getShortcutKey := func(action internal.PromptAction) string {
		if len(action.GetActionKeys()) > 1 {
			return action.GetActionKeys()[1]
		} else {
			return ""
		}
	}

	t := table.NewWriter()

	t.SetStyle(table.StyleDouble)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateFooter = true
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
		{Number: 3, AutoMerge: true},
		{Number: 4, AutoMerge: true},
	})

	t.AppendHeader(table.Row{"Command", "", "Option", "Description"})

	for _, action := range h.actions {
		t.AppendSeparator()

		t.AppendRows([]table.Row{
			{
				action.GetActionKeys()[0],
				getShortcutKey(action),
				"",
				action.GetDescription(markdown)},
		})

		for _, o := range action.GetOptions(markdown) {
			t.AppendRows([]table.Row{
				{
					action.GetActionKeys()[0],
					getShortcutKey(action),
					prettyprint.FormatTextWithColor(o.Value, "Y", markdown),
					"- " + o.Description},
			})
		}

		// add one blank line
		t.AppendRows([]table.Row{
			{
				action.GetActionKeys()[0],
				getShortcutKey(action),
				" ",
				" "},
		})
	}
	return t
}
