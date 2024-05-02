package execs

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
)

type DisplayCollectionExec struct{}

func NewDisplayCollectionExec() DisplayCollectionExec {
	return DisplayCollectionExec{}
}

// Display builds and displays collection using the {out} provided function.
func (u DisplayCollectionExec) Display(collection postman.Collection, filterBy string, out func(string)) {
	out(u.buildWriter(collection.SortByName(), filterBy).Render())
}

func (u DisplayCollectionExec) buildWriter(collection postman.Collection, filterBy string) list.Writer {
	writer := list.NewWriter()
	writer.Reset()
	writer.SetStyle(list.StyleConnectedLight)
	for _, item := range collection.Items {
		appendItem(writer, item, strings.ToLower(strings.TrimSpace(filterBy)), item.Contains(filterBy))
	}
	return writer
}

func appendItem(writer list.Writer, item postman.Item, filterBy string, iContainsPattern postman.ItemContainsPattern) list.Writer {
	if item.Items != nil {
		if iContainsPattern.Contains() {
			writer.AppendItem(strings.ToUpper(item.Name))
		}
		for _, subItem := range item.Items {
			writer.Indent()
			childContainsPattern := item.Contains(filterBy)
			childContainsPattern.Parent = childContainsPattern.Parent || iContainsPattern.Parent
			appendItem(writer, subItem, filterBy, childContainsPattern)
			writer.UnIndent()
		}
		return writer
	} else {
		if filterBy == "" || strings.Contains(strings.ToLower(item.GetLabel()), filterBy) || iContainsPattern.Parent {
			writer.AppendItem(prettyprint.FormatTextWithColor(strings.ToUpper(item.Request.Method), item.Request.Method, false) + " " + item.Name)
		}
		return writer
	}
}
