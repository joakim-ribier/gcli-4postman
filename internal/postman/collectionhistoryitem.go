package postman

import (
	"fmt"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type CollectionHistoryItemsLight []CollectionHistoryItemLight

type CollectionHistoryItemLight struct {
	Number     int
	Item       Item
	Env        *Env
	ExecutedAt time.Time
}

type CollectionHistoryItem struct {
	Number int
	Item   Item

	Status       string
	TimeInMillis int64

	Data          []byte
	ContentLength int64

	Env    *Env
	Params []Param

	ExecutedAt time.Time
}

func NewCollectionHistoryItem(number int, item Item, status string, timeInMillis int64, data []byte, contentLength int64, env *Env, params []Param) CollectionHistoryItem {
	return CollectionHistoryItem{
		Number: number,
		Item:   item,

		Status:       status,
		TimeInMillis: timeInMillis,

		Data:          data,
		ContentLength: contentLength,

		Env:    env,
		Params: params,

		ExecutedAt: time.Now(),
	}
}

// GetBody returns the full {Data} or truncated by the {max} limit.
func (c CollectionHistoryItem) GetData(max int) []byte {
	if len(c.Data) > max && max != -1 {
		return c.Data[:max]
	}
	return c.Data
}

// GetSize returns the size of {Data}.
func (c CollectionHistoryItem) GetSize() int {
	if c.Data == nil {
		return 0
	}
	return len(c.Data)
}

// ToLight transforms a collection history item to a light one
func (c CollectionHistoryItem) ToLight() CollectionHistoryItemLight {
	return CollectionHistoryItemLight{
		Number:     c.Number,
		Item:       c.Item,
		Env:        c.Env,
		ExecutedAt: c.ExecutedAt,
	}

}

// GetSuggestDescription builds the item description for the prompt suggestions.
func (c CollectionHistoryItemLight) GetSuggestDescription() string {
	envName := "No environment"
	if c.Env != nil {
		envName = c.Env.GetName()
	}
	return fmt.Sprintf("%s (@%s)", c.ExecutedAt.Format("2006-01-02 15:04:05"), envName)
}

// GetSuggestText builds the item text for the prompt suggestions.
func (c CollectionHistoryItemLight) GetSuggestText() string {
	return fmt.Sprintf("%s#%d", c.Item.GetLabel(), c.Number)
}

// FindByLabel finds collection history item that matches with {label}.
func (c CollectionHistoryItemsLight) FindByLabel(label string) *CollectionHistoryItemLight {
	return slicesutil.FindT[CollectionHistoryItemLight](c, func(i CollectionHistoryItemLight) bool {
		return i.GetSuggestText() == label
	})
}

// GetNameFile builds the history item filename.
func (c CollectionHistoryItemLight) BuildNameFile() string {
	return c.ExecutedAt.Format("2006-01-02_150405") + "_.json"
}

// SortByExecutedAt sorts collection history items by {executedAt} field.
func (c CollectionHistoryItemsLight) SortByExecutedAt() CollectionHistoryItemsLight {
	return slicesutil.SortT[CollectionHistoryItemLight](c, func(i, j CollectionHistoryItemLight) int {
		return i.ExecutedAt.Compare(j.ExecutedAt) * -1 // inverse order of `Compare` function
	})
}
