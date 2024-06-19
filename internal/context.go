package internal

import (
	"os"

	"github.com/gosimple/slug"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

var GCLI_4POSTMAN_HOME = os.Getenv("GCLI_4POSTMAN_HOME")
var APP_MODE = "user"
var HTTP_BODY_SIZE_LIMIT = 5000
var SECRET = ""
var FILE_LOG = ""
var SEP_CHARACTER = " "
var ENCLOSE_CHARACTER = "'"
var MAX_CMD_HISTORISE = 50

type Context struct {
	WorkspaceName  string
	CollectionName string

	Collection *postman.Collection
	Env        *postman.Env
	Envs       []postman.Env

	CollectionHistoryRequests postman.CollectionHistoryItemsLight
	CMDsHistory               CMDHistories

	Log   logger.Logger
	Print func(string, string, ...any)
}

func NewContext(
	log logger.Logger,
	print func(string, string, ...any)) *Context {

	return &Context{
		Log:   log,
		Print: print,
	}
}

func (c *Context) GetEnvName() string {
	if c.Env != nil {
		return c.Env.GetName()
	} else {
		return ""
	}
}

func (c *Context) Clean() {
	c.WorkspaceName = ""
	c.CollectionName = ""

	c.Collection = nil

	c.Envs = nil
	c.CollectionHistoryRequests = nil

	c.Env = nil
}

func (c *Context) GetCollectionHistoryPathFolder() string {
	return GetHomeWorkspaceFilePath(c.WorkspaceName, c.CollectionName+"-history")
}

func (c *Context) GetCollectionPath() string {
	return GetHomeWorkspaceFilePath(c.WorkspaceName, c.CollectionName+".collection.json")
}

func (c *Context) GetEnvPath(env postman.Env) string {
	return GetHomeWorkspaceFilePath(c.WorkspaceName, slug.Make(env.GetName())+".env.json")
}

func (c *Context) GetCMDHistoryPath() string {
	return GetHomeFilePath("gcli-4postman_cmd.json")
}

func (c *Context) GetWorkspacePath() string {
	return GetHomeWorkspacePath(c.WorkspaceName)
}

func GetHomeWorkspacePath(workspaceName string) string {
	return GCLI_4POSTMAN_HOME + "/" + workspaceName
}

func GetHomeWorkspaceFilePath(workspaceName, file string) string {
	return GetHomeWorkspacePath(workspaceName) + "/" + file
}

func GetHomeFilePath(file string) string {
	return GCLI_4POSTMAN_HOME + "/" + file
}
