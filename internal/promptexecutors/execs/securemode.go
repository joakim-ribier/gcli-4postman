package execs

import (
	"os"
	"strings"

	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/gcli-4postman/pkg/ioutil"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

type SecureModeExec struct {
	tmpSuffix string
	c         internal.Context
	logger    logger.Logger
}

func NewSecureModeExec(c internal.Context, logger logger.Logger) SecureModeExec {
	return SecureModeExec{
		tmpSuffix: "._safe",
		c:         c,
		logger:    logger,
	}
}

// Encrypt (re)encrypts data on disk with the new {secret}.
func (s SecureModeExec) Encrypt(secret string) bool {
	return s.encryptOrDecrypt(secret)
}

// Decrypt decrypts data on disk.
func (s SecureModeExec) Decrypt() bool {
	return s.encryptOrDecrypt("")
}

func (s SecureModeExec) encryptOrDecrypt(secret string) bool {
	if err := s.overwrite(secret); err != nil {
		s.c.Print("WARN", "data could not be overwritten on disk")
		s.clean(false)
		return false
	} else {
		s.c.Print("INFO", "data overwritten on disk")
		s.clean(true)
		return true
	}
}

func (s SecureModeExec) clean(confirm bool) {

	replaceOrDelete := func(filePath string) {
		if strings.Contains(filePath, s.tmpSuffix) {
			if confirm {
				if err := os.Rename(filePath, strings.ReplaceAll(filePath, s.tmpSuffix, "")); err != nil {
					s.logger.Error(err, "file cannot be renamed", "from", filePath, "to", strings.ReplaceAll(filePath, s.tmpSuffix, ""))
				}
			} else {
				if err := os.Remove(filePath); err != nil {
					s.logger.Error(err, "file cannot be deleted", "resource", filePath)
				}
			}
		}
	}

	files, err := os.ReadDir(internal.GCLI_4POSTMAN_HOME)
	if err != nil {
		s.logger.Error(err, "folder cannot be read", "resource", internal.GCLI_4POSTMAN_HOME)
	} else {
		for _, file := range files {
			if file.IsDir() { // is a workspace
				workspace := file.Name()
				files, err := os.ReadDir(internal.GetHomeWorkspacePath(workspace))
				if err != nil {
					s.logger.Error(err, "folder cannot be read", "resource", internal.GetHomeWorkspacePath(workspace))
				} else {
					for _, file := range files {
						if !file.IsDir() {
							replaceOrDelete(internal.GetHomeWorkspaceFilePath(workspace, file.Name()))
						}
					}
				}
			} else {
				replaceOrDelete(internal.GetHomeFilePath(file.Name()))
			}
		}
	}
}

func (s SecureModeExec) overwrite(secret string) error {
	files, err := os.ReadDir(internal.GCLI_4POSTMAN_HOME)
	if err != nil {
		s.logger.Error(err, "folder cannot be read", "resource", internal.GCLI_4POSTMAN_HOME)
		s.c.Print("ERROR", "unable to access files in $$GCLI_4POSTMAN_HOME directory %s", internal.GCLI_4POSTMAN_HOME)
		return err
	} else {
		for _, file := range files {
			if file.Name() == "cmd.json" {
				if err := overwriteT[internal.CMDHistories](internal.GetHomeFilePath(file.Name()), s.tmpSuffix, secret, s.logger); err != nil {
					s.c.Print("ERROR", "unable to overwrite cmd history %s", file.Name())
					return err
				}
			} else {
				if file.IsDir() {
					if err := s.overwriteWorkspace(file.Name(), secret); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

func (s SecureModeExec) overwriteWorkspace(workspace, secret string) error {
	files, err := os.ReadDir(internal.GetHomeWorkspacePath(workspace))
	if err != nil {
		s.logger.Error(err, "folder cannot be read", "resource", internal.GetHomeWorkspacePath(workspace))
		s.c.Print("ERROR", "unable to access files in workspace directory %s", workspace)
		return err
	} else {
		for _, file := range files {
			filePath := internal.GetHomeWorkspaceFilePath(workspace, file.Name())
			if strings.Contains(file.Name(), ".collection.json") {
				if err := overwriteT[postman.Collection](filePath, s.tmpSuffix, secret, s.logger); err != nil {
					s.c.Print("ERROR", "unable to overwrite collection %s", filePath)
					return err
				}
			}
			if strings.Contains(file.Name(), ".env.json") {
				if err := overwriteT[postman.Env](filePath, s.tmpSuffix, secret, s.logger); err != nil {
					s.c.Print("ERROR", "unable to overwrite environment %s", filePath)
					return err
				}
			}
			if file.IsDir() && strings.Contains(file.Name(), "-history") {
				if err := os.RemoveAll(filePath); err != nil {
					s.c.Print("ERROR", "unable to remove collection history %s", filePath)
					return err
				}
			}
		}
		return nil
	}
}

func overwriteT[T any](pathFile, tmpSuffix, newSecret string, logger logger.Logger) error {
	if t, err := ioutil.Load[T](pathFile, internal.SECRET); err != nil {
		logger.Error(err, "file cannot be loaded", "resource", pathFile)
		return err
	} else {
		err := ioutil.Write[T](t, pathFile+tmpSuffix, newSecret)
		if err != nil {
			logger.Error(err, "file cannot be written", "resource", pathFile+tmpSuffix)
		}
		return err
	}
}
