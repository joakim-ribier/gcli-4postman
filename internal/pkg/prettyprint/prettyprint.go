package prettyprint

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/tidwall/pretty"
)

// Print prints text on the console.
func print(title string, content string, prefix string) {
	if title != "" {
		fmt.Printf("%s\n", title)
	}

	for _, line := range strings.Split(content, "\n") {
		fmt.Printf("%s%s\n", prefix, line)
	}
}

// PrintJson prints and returns json value
func PrintJson(data []byte, prettyJson bool) string {
	if prettyJson {
		data = pretty.Pretty(data)
	}
	str := string(pretty.Color(data, nil)[:])
	print("", str, "")
	return str
}

// PrintInColor colorizes and prints the 'str' value based on the 'pattern'
func PrintInColor(str, level string, markdown bool) {
	print("", FormatTextWithColor(str, level, markdown), "")
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
