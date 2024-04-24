package postman

import (
	"fmt"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type CollectionHistoryItems []CollectionHistoryItem

type CollectionHistoryItem struct {
	Item *Item

	Status string
	Body   []byte

	Env    *Env
	Params []Param

	ExecutedAt time.Time
}

func NewCollectionHistoryItem(item *Item, status string, body []byte, env *Env, params []Param) CollectionHistoryItem {
	return CollectionHistoryItem{
		Item: item,

		Status: status,
		Body:   body,

		Env:    env,
		Params: params,

		ExecutedAt: time.Now(),
	}
}

// GetSuggestDescription builds a descriptionof the collection history items for the prompt.
func (i CollectionHistoryItem) GetSuggestDescription() string {
	envName := "No environment"
	if i.Env != nil {
		envName = i.Env.GetName()
	}
	return fmt.Sprintf("%s (@%s)", i.ExecutedAt.Format("2006-01-02 15:04:05"), envName)
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
		return i.Item.GetLabel() == label
	})
}
