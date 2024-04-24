package internal

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type PromptCallback struct {
	Text         string
	Callbacks    []PromptSuggestCallback
	PromptAction PromptAction
	Args         []any
}

func (p PromptCallback) Callback(in []string, actions []PromptAction) {
	p.PromptAction.PromptCallback(in, actions, p.Args...)
}

func (p PromptCallback) IsAvailableAnswer(in string) bool {
	return slicesutil.ExistT[prompt.Suggest](p.GetSuggests(), func(s prompt.Suggest) bool { return strings.EqualFold(s.Text, in) })
}

type PromptSuggestCallback struct {
	Text        string
	Description string
}

func (p PromptCallback) GetSuggests() []prompt.Suggest {
	return slicesutil.TransformT[PromptSuggestCallback, prompt.Suggest](p.Callbacks, func(psc PromptSuggestCallback) (*prompt.Suggest, error) {
		return &prompt.Suggest{Text: psc.Text, Description: psc.Description}, nil
	})
}

func NewPromptCallback(text string, callbacks []PromptSuggestCallback, promptAction PromptAction, args ...any) *PromptCallback {
	return &PromptCallback{
		Text:         text,
		Callbacks:    callbacks,
		PromptAction: promptAction,
		Args:         args,
	}
}

func NewPromptSuggestCallback(text string, desc string) PromptSuggestCallback {
	return PromptSuggestCallback{
		Text:        text,
		Description: desc,
	}
}
