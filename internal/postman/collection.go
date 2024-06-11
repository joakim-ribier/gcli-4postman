package postman

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type Info struct {
	Name string
}

type Collection struct {
	Info     Info
	Items    Items    `json:"item"`
	Metadata Metadata // compute date
}

type Metadata struct {
	Methods []string
}

type Items []Item
type Headers []Header

type Item struct {
	Name    string
	Request Request `json:"request,omitempty"`
	Items   Items   `json:"item,omitempty"`
}

type Request struct {
	Method string
	Header Headers `json:"header,omitempty"`
	Url    Url
	Auth   Auth
	Body   Body
}

type Header struct {
	Key   string
	Value string
}

type Url struct {
	Raw  string
	Path []string
}

type Auth struct {
	Type  string
	Basic []AuthBasicValue
}

type AuthBasicValue struct {
	Key   string
	Value string
}

type Body struct {
	Raw string
}

type Param struct {
	Key   string
	Value string
}

type ItemContainsPattern struct {
	Parent bool
	Child  bool
}

// GetMethods finds the items methods.
func (c Collection) GetMethods() []string {
	return findMethods(c.Items, []string{})
}

// FindByMethod finds the items that match with the {method}.
func (c Collection) FindByMethod(method string) Items {
	return findByMethod(c.Items, []Item{}, strings.ToLower(strings.TrimSpace(method)))
}

// FindItemByLabel finds the item that matches with the {label}.
func (c Collection) FindItemByLabel(label string) *Item {
	return findByLabel(c.Items, label)
}

// SortByName sorts collection items by the {name} field.
func (c Collection) SortByName() Collection {
	c.Items.SortByName()
	return c
}

// SortByName sorts items by the {name} field.
func (items Items) SortByName() Items {
	sortItemsBy(items, func(item Item) string {
		return item.Name
	})
	return items
}

// SortByLabel sorts items by the {label} computed field.
func (items Items) SortByLabel() Items {
	sortItemsBy(items, func(item Item) string {
		return item.GetLabel()
	})
	return items
}

// GetLabel builds the item label "{METHOD}.../{path[len(path)-3:]}" computed field.
func (i Item) GetLabel() string {
	return fmt.Sprintf("%s../%s", i.Request.Method, strings.ToLower(strings.ReplaceAll(i.Name, " ", "_")))
}

// Get builds the header params using the provided context (env and params).
func (h Headers) Get(env *Env, params []Param) map[string]string {
	var out = make(map[string]string, len(h))
	for _, header := range h {
		if v := replaceRawWithParams(header.Value, env, params); v != "--delete" {
			out[header.Key] = v
		}
	}
	return out
}

// Get builds the API {url} using the provided context (env and params).
func (u Url) Get(env *Env, params []Param) string {
	return replaceRawWithParams(u.Raw, env, params)
}

// GetLongPath builds the API path.
func (u Url) GetLongPath() string {
	return strings.Join(u.Path[:], "/")
}

// GetShortPath builds the short (3 max) API path.
func (u Url) GetShortPath() string {
	if len(u.Path) > 3 {
		return strings.Join(u.Path[len(u.Path)-3:], "/")
	} else {
		return strings.Join(u.Path, "/")
	}
}

// Get builds the API {body} using the provided context (env and params).
func (u Body) Get(env *Env, params []Param) string {
	return replaceRawWithParams(u.Raw, env, params)
}

// replaceRawWithParams finds and replaces params in {raw} by the values using the provided context (params then env).
func replaceRawWithParams(raw string, env *Env, params []Param) string {
	for _, param := range params {
		raw = strings.ReplaceAll(raw, param.Key, param.Value)
	}
	if env != nil {
		for _, param := range env.Params {
			key := "{{" + param.Key + "}}"
			raw = strings.ReplaceAll(raw, key, param.Value)
		}
	}
	return raw
}

// GetAuthBasicCredential returns the couple {username}/{password} for basic authentication type.
func (a Auth) GetAuthBasicCredential(env *Env, params []Param) (string, string) {
	if a.Type != "basic" {
		return "", ""
	}
	username := a.Basic[slices.IndexFunc(a.Basic, func(el AuthBasicValue) bool { return el.Key == "username" })].Value
	password := a.Basic[slices.IndexFunc(a.Basic, func(el AuthBasicValue) bool { return el.Key == "password" })].Value

	for _, param := range params {
		username = strings.ReplaceAll(username, param.Key, param.Value)
		password = strings.ReplaceAll(password, param.Key, param.Value)
	}
	if env != nil {
		for _, param := range env.Params {
			key := "{{" + param.Key + "}}"
			username = strings.ReplaceAll(username, key, param.Value)
			password = strings.ReplaceAll(password, key, param.Value)
		}
	}
	return username, password
}

func (i ItemContainsPattern) Contains() bool {
	return i.Parent || i.Child
}

func (a Auth) extractParams(extract func(string) []string) []string {
	return slicesutil.FlatTransformT[AuthBasicValue, string](a.Basic, func(abv AuthBasicValue) ([]string, error) {
		return extract(abv.Value), nil
	})
}

func (h Headers) extractParams(extract func(string) []string) []string {
	return slicesutil.FlatTransformT[Header, string](h, func(h Header) ([]string, error) {
		return extract(h.Value), nil
	})
}

// GetParams finds all params from {url} annd {body} fields.
func (i Item) GetParams() []string {
	r, _ := regexp.Compile(`{{[a-zA-Z_#\- ]*}}`)
	extract := func(in string) []string {
		return r.FindAllString(in, -1)
	}

	return slicesutil.NewSliceS(extract(i.Request.Body.Raw)).
		Append(extract(i.Request.Url.Raw)).
		Append(i.Request.Auth.extractParams(extract)).
		Append(i.Request.Header.extractParams(extract)).
		Distinct().
		Sort()
}

// IsRequest returns {true} if the item is an API request.
func (i Item) IsRequest() bool {
	return i.Request.Method != ""
}

func (i Item) Contains(pattern string) ItemContainsPattern {
	return contains(i, strings.ToLower(strings.TrimSpace(pattern)), ItemContainsPattern{false, false})
}

func contains(i Item, pattern string, iContainsPattern ItemContainsPattern) ItemContainsPattern {
	if i.Items != nil {
		if strings.Contains(strings.ToLower(i.Name), pattern) {
			iContainsPattern.Parent = true
		}
		for _, subItem := range i.Items {
			iContainsPattern = contains(subItem, pattern, iContainsPattern)
		}
	} else {
		iContainsPattern.Child = strings.Contains(strings.ToLower(i.GetLabel()), pattern) || iContainsPattern.Child
	}
	return iContainsPattern
}

// sortItemsBy sorts the items using the provided {f} function.
func sortItemsBy(items []Item, f func(Item) string) {
	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(f(items[i])) < strings.ToLower(f(items[j]))
	})
	for _, values := range items {
		sortItemsBy(values.Items, f)
	}
}

// recursive function that finds the item that matchs with the {label}.
func findByLabel(items Items, label string) *Item {
	for _, item := range items {
		if item.GetLabel() == label {
			return &item
		}
		if i := findByLabel(item.Items, label); i != nil {
			return i
		}
	}
	return nil
}

// recursive function that returns the methods using by the collection.
func findMethods(items []Item, methods []string) []string {
	for _, item := range items {
		if item.IsRequest() && !slices.Contains(methods, item.Request.Method) {
			methods = append(methods, item.Request.Method)
		}
		methods = findMethods(item.Items, methods)
	}
	return slicesutil.Sort(methods)
}

// recursive function that finds items that match with the {method}.
func findByMethod(items []Item, list Items, method string) []Item {
	for _, item := range items {
		if item.IsRequest() && (strings.EqualFold(item.Request.Method, method) || method == "") {
			list = append(list, item)
		}
		list = findByMethod(item.Items, list, method)
	}
	return list.SortByName()
}
