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
