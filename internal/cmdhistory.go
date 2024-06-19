package internal

import (
	"time"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type CMDHistories []CMDHistory

type CMDHistory struct {
	CMD        string
	ExecutedAt time.Time
}

func NewCMDHistory(cmd string) CMDHistory {
	return CMDHistory{
		CMD:        cmd,
		ExecutedAt: time.Now(),
	}
}

// GetName returns the CMD history
func (s CMDHistories) GetName() []string {
	out := slicesutil.SortTByTime[CMDHistory](s, func(c1, c2 CMDHistory) (time.Time, time.Time) {
		return c1.ExecutedAt, c2.ExecutedAt
	})
	return slicesutil.TransformT[CMDHistory, string](out, func(c CMDHistory) (*string, error) {
		return &c.CMD, nil
	})
}

// SortByExecutedAt sorts the CMD by {ExecutedAt}
func (s CMDHistories) SortByExecutedAt() []CMDHistory {
	return slicesutil.SortTByTime[CMDHistory](s, func(c1, c2 CMDHistory) (time.Time, time.Time) {
		return c1.ExecutedAt, c2.ExecutedAt
	})
}
