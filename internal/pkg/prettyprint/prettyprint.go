package prettyprint

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/tidwall/pretty"
)

// Print prints text on the console.
func Print(value string) {
	for _, line := range strings.Split(value, "\n") {
		fmt.Printf("%s\n", line)
	}
}

// PrintJson prints and returns json value
func SPrintJson(data []byte, prettyJson bool) string {
	if prettyJson {
		data = pretty.Pretty(data)
	}
	return string(pretty.Color(data, nil)[:])
}

// PrintInColor colorizes and prints the 'str' value based on the 'pattern'
func SPrintInColor(str, level string, markdown bool) string {
	return FormatTextWithColor(str, level, markdown)
}

// FormatTextWithColor colorizes the 'str' value based on the 'pattern'
func FormatTextWithColor(str, level string, markdown bool) string {
	return formatText(str, level, markdown)
}

func formatText(str, level string, markdown bool) string {
	if markdown {
		return fmt.Sprintf("`%s`", str)
	}

	switch strings.ToLower(level) {
	case "delete":
		return text.FgRed.Sprintf(str)
	case "r", "error":
		return text.FgRed.Sprintf(str)
	case "g", "green", "get", "info":
		return text.FgGreen.Sprintf(str)
	case "y", "yellow", "warn", "post":
		return text.FgYellow.Sprintf(str)
	case "b", "blue", "put", "path":
		return text.FgBlue.Sprintf(str)
	case "patch":
		return text.FgMagenta.Sprintf(str)
	default:
		return str
	}
}
