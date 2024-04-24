package promptactions

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/gosimple/slug"
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/pkg/prettyprint"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
	"github.com/joakim-ribier/go-utils/pkg/httpsutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

const (
	apiKeyParam    = "--apiKey"
	workspaceParam = "-workspace"
	syncParam      = "-sync"

	workspacesEndpoint  = "https://api.getpostman.com/workspaces"
	collectionsEndpoint = "https://api.getpostman.com/collections"
	envsEndpoint        = "https://api.getpostman.com/environments"
)

type PromptPostman struct {
	c      *internal.Context
	logger logger.Logger
}

func NewPromptPostman(c *internal.Context) internal.PromptAction {
	p := PromptPostman{c: c}
	p.logger = c.Log.Namespace(p.GetName())
	return p
}

func (p PromptPostman) GetName() string {
	return "PromptPostman"
}

func (p PromptPostman) GetPromptExecutor() internal.PromptExecutor {
	return nil
}

func (p PromptPostman) GetActionKeys() []string {
	return []string{"postman", ":p"}
}

func (p PromptPostman) GetParamKeys() []internal.ParamWithRole {
	return []internal.ParamWithRole{
		{Value: workspaceParam, Roles: []string{"admin"}},
		{Value: syncParam, Roles: []string{"admin"}},
	}
}

func (p PromptPostman) GetOptions(markdown bool) []internal.Option {
	return []internal.Option{
		{Value: apiKeyParam, Description: "API keys settings"},
		{Value: workspaceParam, Description: "display the remote workspaces linked to the {API_KEY}"},
		{Value: fmt.Sprintf("%s {workspace Id/Name}", syncParam), Description: "sync one of the workspaces locally"},
	}
}

func (p PromptPostman) GetDescription(markdown bool) string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("Connexion to a %s account to sync the workspaces on the local disk.", prettyprint.FormatTextWithColor("Postman", "Y", markdown)))
	builder.WriteString(fmt.Sprintf("\n%s", prettyprint.FormatTextWithColor("# :p --apiKey {KEY} -sync {workspace}", "Y", markdown)))
	return builder.String()
}

func (p PromptPostman) PromptSuggest(in []string, d prompt.Document) ([]prompt.Suggest, error) {
	if !slices.Contains(p.GetActionKeys(), in[0]) || internal.APP_MODE != "admin" {
		return []prompt.Suggest{}, nil
	}

	if !slicesutil.Exist(in, apiKeyParam) || slicesutil.FindNextEl(in, apiKeyParam) == "" {
		return []prompt.Suggest{
			{Text: apiKeyParam, Description: "Postman API_KEY"}}, nil
	}

	return []prompt.Suggest{
		{Text: workspaceParam, Description: "list remote workspaces"},
		{Text: syncParam, Description: "sync a specific workspace"}}, nil
}

func (p PromptPostman) PromptExecutor(in []string) *internal.PromptCallback {
	if internal.HasRight(p, in, internal.APP_MODE) {
		apiKey := slicesutil.FindNextEl(in, apiKeyParam)

		if apiKey == "" {
			p.c.Print("WARN", "set {API_KEY} to connect to the account...")
			return nil
		}

		if slicesutil.Exist(in, workspaceParam) {
			internal.AddCMDHistory(*p.c, strings.Join(in, " "))

			if bytes := p.call(workspacesEndpoint, apiKey); bytes != nil {
				prettyprint.PrintJson(bytes, slicesutil.Exist(in, "--pretty"))
			}
			return nil
		}

		if slicesutil.Exist(in, syncParam) {
			internal.AddCMDHistory(*p.c, strings.Join(in, " "))

			p.c.Print("INFO", "sync workspace with its collections and environments")
			p.c.Print("INFO", "find workspace \"%s\" ...", slicesutil.FindNextEl(in, syncParam))
			workspace := p.downloadWorkspace(apiKey).Find(slicesutil.FindNextEl(in, syncParam))
			if workspace == nil {
				p.c.Print("WARN", "workspace {%s} not found...", slicesutil.FindNextEl(in, syncParam))
				return nil
			}

			p.c.Print("INFO", "download environments ...")
			envs, _ := p.downloadEnvs(apiKey, *workspace)
			p.c.Print("WARN", "\"%s\" found", slicesutil.ToStringT[postman.Env](envs, func(e postman.Env) *string {
				s := e.GetName()
				return &s
			}, "\", \""))

			p.c.Print("INFO", "download collections ...")
			collections, _ := p.downloadCollections(apiKey, *workspace)
			p.c.Print("WARN", "\"%s\" found", slicesutil.ToStringT[postman.Collection](collections, func(c postman.Collection) *string { return &c.Info.Name }, "\", \""))

			return internal.NewPromptCallback(
				fmt.Sprintf("Update data for the \"%s\" workspace (Yes / No)", workspace.Name),
				[]internal.PromptSuggestCallback{
					internal.NewPromptSuggestCallback("Yes", "Erase current data and reload the collection"),
					internal.NewPromptSuggestCallback("No", "Do nothing")},
				p, *workspace, collections, envs)

		}

		p.c.Print("WARN", "select an available option to continue...")
	}
	return nil
}

func (p PromptPostman) downloadWorkspace(apiKey string) postman.PSTWorkspaces {
	if bytes := p.call(workspacesEndpoint, apiKey); bytes == nil {
		return postman.PSTWorkspaces{}
	} else {
		workspaces, err := jsonsutil.Unmarshal[postman.PSTWorkspaces](bytes)
		if err != nil {
			p.logger.Error(err, "`bytes` cannot be unmarshaled", "data", bytes)
			return postman.PSTWorkspaces{}
		} else {
			return workspaces
		}
	}
}

func (p PromptPostman) downloadEnvs(apiKey string, workspace postman.PSTWorkspace) ([]postman.Env, error) {
	var envs []postman.Env
	if bytes := p.call(fmt.Sprintf("%s?workspace=%s", envsEndpoint, workspace.Id), apiKey); bytes == nil {
		return []postman.Env{}, nil
	} else {
		pstEnvironments, err := jsonsutil.Unmarshal[postman.PSTEnvironments](bytes)
		if err != nil {
			p.logger.Error(err, "`bytes` cannot be unmarshaled", "data", bytes)
			return nil, err
		}
		for _, e := range pstEnvironments.Environments {
			if bytes := p.call(fmt.Sprintf("%s/%s", envsEndpoint, e.Id), apiKey); bytes != nil {
				postmanEnv, err := jsonsutil.Unmarshal[postman.PostmanEnvironment](bytes)
				if err != nil {
					p.logger.Error(err, "`bytes` cannot be unmarshaled", "data", bytes)
					return nil, err
				}
				envs = append(envs, postmanEnv.Environment)
			}
		}
	}
	return envs, nil
}

func (p PromptPostman) downloadCollections(apiKey string, workspace postman.PSTWorkspace) ([]postman.Collection, error) {
	var collections []postman.Collection
	if bytes := p.call(fmt.Sprintf("%s?workspace=%s", collectionsEndpoint, workspace.Id), apiKey); bytes == nil {
		return []postman.Collection{}, nil
	} else {
		pstCollections, err := jsonsutil.Unmarshal[postman.PSTCollections](bytes)
		if err != nil {
			p.logger.Error(err, "`bytes` cannot be unmarshaled", "data", bytes)
			return nil, err
		}
		for _, c := range pstCollections.Collections {
			if bytes := p.call(fmt.Sprintf("%s/%s", collectionsEndpoint, c.Id), apiKey); bytes != nil {
				postmanCollection, err := jsonsutil.Unmarshal[postman.PostmanCollection](bytes)
				if err != nil {
					p.logger.Error(err, "`bytes` cannot be unmarshaled", "data", bytes)
					return nil, err
				}
				collections = append(collections, postmanCollection.Collection)
			}
		}
	}
	return collections, nil
}

func (p PromptPostman) save(workspace postman.PSTWorkspace, collections []postman.Collection, envs []postman.Env) {
	r := true

	workspaceTemporaryFolder := p.buildWorkspaceRootFolder(workspace.Name, "_"+time.Now().Format("2006-01-02_150405"))
	if err := os.Mkdir(workspaceTemporaryFolder, os.ModePerm); err != nil {
		p.logger.Error(err, "folder cannot be created", "resource", workspaceTemporaryFolder)
		p.c.Print("ERROR", "unable to create temporary workspace folder \"%s\"", workspaceTemporaryFolder)
		return
	}
	p.c.Print("INFO", "create temporary folder \"%s\"", workspaceTemporaryFolder)

	for _, env := range envs {
		envName := slug.Make(env.GetName())
		if envName != "" {
			envFileName := workspaceTemporaryFolder + "/" + envName + ".env.json"
			if err := ioutil.Write[postman.Env](env, envFileName, internal.SECRET); err != nil {
				p.logger.Error(err, "file cannot be written", "resource", envFileName)
				r = false
				break
			} else {
				p.c.Print("INFO", "write file \"%s\"", envName+".env.json")
			}
		}
	}

	for _, collection := range collections {
		collectionName := slug.Make(collection.Info.Name)
		if collectionName != "" {
			collectionFileName := workspaceTemporaryFolder + "/" + collectionName + ".collection.json"
			if err := ioutil.Write[postman.Collection](collection, collectionFileName, internal.SECRET); err != nil {
				p.logger.Error(err, "file cannot be written", "resource", collectionFileName)
				r = false
				break
			} else {
				p.c.Print("INFO", "write file \"%s\"", collectionName+".collection.json")
			}
		}
	}

	if !r {
		p.c.Print("WARN", "unable to terminate the updating, remove the temporary folder... retry!")
		os.RemoveAll(workspaceTemporaryFolder)
	} else {
		// remove the current workspace folder if it exists
		if err := p.deleteSafely(p.buildWorkspaceRootFolder(workspace.Name, "")); err != nil {
			r = false
			p.logger.Error(err, "file cannot be deleted", "resource", p.buildWorkspaceRootFolder(workspace.Name, ""))
			p.c.Print("ERROR", "unable to remove \"%s\" folder", p.buildWorkspaceRootFolder(workspace.Name, ""))
		} else {
			// rename the temporary workspace to the workspace folder
			if err := os.Rename(workspaceTemporaryFolder, p.buildWorkspaceRootFolder(workspace.Name, "")); err != nil {
				r = false
				p.logger.Error(err, "file cannot be renamed", "from", workspaceTemporaryFolder, "to", p.buildWorkspaceRootFolder(workspace.Name, ""))
				p.c.Print("ERROR", "unable to rename \"%s\" to \"%s\", do it manually...", workspaceTemporaryFolder, p.buildWorkspaceRootFolder(workspace.Name, ""))
			} else {
				p.c.Print("INFO", "rename temporay folder to \"%s\"", p.buildWorkspaceRootFolder(workspace.Name, ""))
			}
		}
	}

	if r {
		p.c.Print("INFO", "Workspace \"%s\" is up to date!", workspace.Name)
	}
}

func (p PromptPostman) buildWorkspaceRootFolder(workspaceName, suffix string) string {
	return internal.GCLI_4POSTMAN_HOME + "/" + slug.Make(workspaceName) + suffix
}

// remove safely, check if the folder is not the root folder!
func (p PromptPostman) deleteSafely(folder string) error {
	if folder != internal.GCLI_4POSTMAN_HOME && folder != internal.GCLI_4POSTMAN_HOME+"/" {
		if err := os.RemoveAll(folder); err != nil {
			p.logger.Error(err, "folder cannot be deleted", "resource", folder)
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf("unable to remove folder {%s}", folder)
	}
}

// Call executes item request
func (p PromptPostman) call(url string, APIKey string) []byte {
	p.logger.Info("call HTTP request", "resource", url)

	r, err := httpsutil.NewHttpRequest(url, "")

	if err != nil {
		p.logger.Error(err, "request cannot be called", "resource", url)
		return nil
	}

	resp, err := r.
		Header("User-Agent", "PostmanRuntime/7.35.0").
		Header("X-API-Key", APIKey).
		AsJson().Call()

	if err == nil {
		return resp.Body
	} else {
		p.logger.Error(err, "request cannot be called", "resource", url)
		return nil
	}
}

func (p PromptPostman) PromptCallback(in []string, actions []internal.PromptAction, args ...any) {
	if slicesutil.Exist(in, "Yes") {
		p.save(args[0].(postman.PSTWorkspace), args[1].([]postman.Collection), args[2].([]postman.Env))
		p.c.Clean()
	}
}
