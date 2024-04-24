package prettyprint

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/tidwall/pretty"
)

// PrintCollection prints the wole collection on the console.
func PrintCollection(collection postman.Collection, pattern string) {
	if len(collection.Items) > 0 {
		writer := list.NewWriter()
		writer.Reset()
		writer.SetStyle(list.StyleConnectedLight)
		for _, item := range collection.Items {
			appendItem(writer, item, strings.ToLower(strings.TrimSpace(pattern)), item.Contains(pattern))
		}
		print("["+collection.Info.Name+"'s collection]", writer.Render(), "")
	}
}

/* recursive method to build collection writer */
func appendItem(writer list.Writer, item postman.Item, pattern string, iContainsPattern postman.ItemContainsPattern) list.Writer {
	if item.Items != nil {
		if iContainsPattern.Contains() {
			writer.AppendItem(strings.ToUpper(item.Name))
		}
		for _, subItem := range item.Items {
			writer.Indent()
			childContainsPattern := item.Contains(pattern)
			childContainsPattern.Parent = childContainsPattern.Parent || iContainsPattern.Parent
			appendItem(writer, subItem, pattern, childContainsPattern)
			writer.UnIndent()
		}
		return writer
	} else {
		if pattern == "" || strings.Contains(strings.ToLower(item.GetLabel()), pattern) || iContainsPattern.Parent {
			writer.AppendItem(FormatTextWithColor(strings.ToUpper(item.Request.Method), item.Request.Method, false) + " " + item.Name)
		}
		return writer
	}
}

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
