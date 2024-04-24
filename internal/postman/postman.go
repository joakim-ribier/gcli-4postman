package postman

import (
	"strings"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type PSTCollections struct {
	Collections []PSTCollection
}

type PSTWorkspaces struct {
	Workspaces []PSTWorkspace
}

type PSTWorkspace struct {
	Id   string
	Name string
}

type PSTCollection struct {
	Id   string
	Name string
}

type PSTEnvironments struct {
	Environments []PSTEnvironment
}

type PSTEnvironment struct {
	Id   string
	Name string
}

func (w PSTWorkspaces) Find(workspaceIdOrName string) *PSTWorkspace {
	return slicesutil.FindT(w.Workspaces, func(w PSTWorkspace) bool {
		return w.Id == workspaceIdOrName || strings.EqualFold(w.Name, workspaceIdOrName)
	})
}
