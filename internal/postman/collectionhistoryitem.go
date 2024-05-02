package postman

import (
	"fmt"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type CollectionHistoryItems []CollectionHistoryItem

type CollectionHistoryItem struct {
	Number int
	Item   Item

	Status string
	Body   []byte

	Env    *Env
	Params []Param

	ExecutedAt time.Time
}

func NewCollectionHistoryItem(number int, item Item, status string, body []byte, env *Env, params []Param) CollectionHistoryItem {
	return CollectionHistoryItem{
		Number: number,
		Item:   item,

		Status: status,
		Body:   body,

		Env:    env,
		Params: params,

		ExecutedAt: time.Now(),
	}
}

// GetSuggestDescription builds the item description for the prompt suggestions.
func (i CollectionHistoryItem) GetSuggestDescription() string {
	envName := "No environment"
	if i.Env != nil {
		envName = i.Env.GetName()
	}
	return fmt.Sprintf("%s (@%s)", i.ExecutedAt.Format("2006-01-02 15:04:05"), envName)
}

// GetSuggestText builds the item text for the prompt suggestions.
func (i CollectionHistoryItem) GetSuggestText() string {
	return fmt.Sprintf("%s#%d", i.Item.GetLabel(), i.Number)
}

// SortByExecutedAt sorts collection history items by {executedAt} field.
func (r CollectionHistoryItems) SortByExecutedAt() CollectionHistoryItems {
	return slicesutil.SortT[CollectionHistoryItem](r, func(i, j CollectionHistoryItem) bool {
		return i.ExecutedAt.After(j.ExecutedAt)
	})
}

// FindByLabel finds collection history item that matches with {label}.
func (r CollectionHistoryItems) FindByLabel(label string) *CollectionHistoryItem {
	return slicesutil.FindT[CollectionHistoryItem](r, func(i CollectionHistoryItem) bool {
		return i.GetSuggestText() == label
	})
}
