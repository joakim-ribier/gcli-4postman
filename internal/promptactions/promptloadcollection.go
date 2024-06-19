package promptactions

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type PromptLoadCollection struct {
	c      *internal.Context
	logger logger.Logger
}

func NewPromptLoadCollection(c *internal.Context) internal.PromptAction {
	p := PromptLoadCollection{c: c}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptLoadCollection) GetName() string {
	return "PromptLoadCollection"
}

func (p PromptLoadCollection) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptLoadCollection) GetActionKeys() []string {
	return []string{"load", ":l"}
}

func (p PromptLoadCollection) GetParamKeys() []internal.ParamWithRole {
	return nil
}

func (p PromptLoadCollection) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{}
}

func (p PromptLoadCollection) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Load a collection - %s - from the local disk.", prettyprint.FormatTextWithColor("Postman API HTTP requests format", "Y", markdown)))
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :l my-collection", "Y", markdown)))
	return builder.String()
}

func (p PromptLoadCollection) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) {
		return []prompt.Suggest{}, nil
	}

	var suggests []prompt.Suggest
	for workspace, collections := range p.getFolders() {
		for _, collection := range collections {
			suggests = append(suggests, prompt.Suggest{Text: fmt.Sprintf("%s/_%s", workspace, collection), Description: ""})
		}
	}
	return slicesutil.SortT(suggests, func(s1, s2 prompt.Suggest) (string, string) {
		return s1.Text, s2.Text
	}), nil
}

func (p PromptLoadCollection) getFolders() map[string][]string {
	var workspaceMapCollections = make(map[string][]string)
	folders, _ := os.ReadDir(internal.GCLI_4POSTMAN_HOME)
	for _, folder := range folders {
		if folder.IsDir() {
			files, err := os.ReadDir(internal.GCLI_4POSTMAN_HOME + "/" + folder.Name() + "/")
			if err != nil {
				p.logger.Error(err, "folder cannot be opened", "GCLI_4POSTMAN_HOME", internal.GCLI_4POSTMAN_HOME+"/"+folder.Name()+"/")
				break
			}
			var collections []string
			for _, file := range files {
				if strings.Contains(file.Name(), ".collection.json") {
					collections = append(collections, strings.ReplaceAll(file.Name(), ".collection.json", ""))
				}
			}
			if len(collections) > 0 {
				workspaceMapCollections[folder.Name()] = collections
			}
		}
	}
	return workspaceMapCollections
}

func (p PromptLoadCollection) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRightToExecute(p, in, internal.APP_MODE) {
		if len(in) > 1 {
			if tab := strings.Split(strings.ReplaceAll(in[1], "", ""), "/_"); len(tab) == 2 {
				if collections, is := p.getFolders()[tab[0]]; is && slicesutil.Exist(collections, tab[1]) {
					p.c.Clean()
					p.c.WorkspaceName = tab[0]
					p.c.CollectionName = tab[1]
					p.load()
					return nil
				}
			}
		}
		p.c.Print("WARN", "select a collection from the suggestions")
	}
	return nil
}

func (p PromptLoadCollection) load() {
	if p.c.WorkspaceName == "" || p.c.CollectionName == "" {
		p.c.Print("WARN", "unable to load collection %s...", p.c.CollectionName)
		return
	}

	c, err := ioutil.Load[postman.Collection](p.c.GetCollectionPath(), internal.SECRET)
	if err != nil {
		p.logger.Error(err, "file cannot be loaded", "resource", p.c.GetCollectionPath())
		p.c.Print("ERROR", "unable to load collection '%s/%s'", p.c.WorkspaceName, p.c.CollectionName)
		return
	} else {
		p.c.Print("INFO", "loads collection '%s' from workspace '%s'", p.c.CollectionName, p.c.WorkspaceName)
		p.c.Collection = &c
	}

	files, err := os.ReadDir(p.c.GetWorkspacePath())
	if err != nil {
		p.logger.Error(err, "folder cannot be read", "resource", p.c.GetWorkspacePath())
		p.c.Print("ERROR", "unable to load environment files in the workspace '%s'", p.c.WorkspaceName)
		p.c.Envs = []postman.Env{}
	} else {
		for _, file := range files {
			if !file.IsDir() && strings.Contains(strings.ToLower(file.Name()), ".env.json") {
				env, err := ioutil.Load[postman.Env](p.c.GetWorkspacePath()+"/"+file.Name(), internal.SECRET)
				if err != nil {
					p.logger.Error(err, "file cannot be loaded", "resource", p.c.GetWorkspacePath()+"/"+file.Name())
					p.c.Print("ERROR", "unable to load environment '%s'", file.Name())
				} else {
					p.c.Print("INFO", "loads environment '%s' from workspace '%s'", env.GetName(), p.c.WorkspaceName)
					p.c.Envs = append(p.c.Envs, env)
				}
			}
		}
	}

	files, err = os.ReadDir(p.c.GetCollectionHistoryPathFolder())
	if err != nil {
		p.logger.Error(err, "folder cannot be read", "resource", p.c.GetCollectionHistoryPathFolder())
		p.c.Print("WARN", "unable to load history items files for the collection '%s/%s'", p.c.WorkspaceName, p.c.CollectionName)
		p.c.CollectionHistoryRequests = postman.CollectionHistoryItemsLight{}
	} else {
		for _, file := range files {
			if !file.IsDir() {
				historyItem, err := ioutil.Load[postman.CollectionHistoryItemLight](p.c.GetCollectionHistoryPathFolder()+"/"+file.Name(), internal.SECRET)
				if err != nil {
					p.logger.Error(err, "file cannot be loaded", "resource", p.c.GetCollectionHistoryPathFolder()+"/"+file.Name())
					p.c.Print("ERROR", "unable to load history request '%s'", file.Name())
				} else {
					p.c.CollectionHistoryRequests = append(p.c.CollectionHistoryRequests, historyItem)
				}
			}
		}
		if len(p.c.CollectionHistoryRequests) > 0 {
			p.c.Print("INFO", "loads collection %s history requests (%d)", p.c.CollectionName, len(p.c.CollectionHistoryRequests))
		}
	}
}

func (p PromptLoadCollection) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	// -- not used
}
