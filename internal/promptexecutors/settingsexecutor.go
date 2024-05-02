package promptexecutors

import (
	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/internal/promptexecutors/execs"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

// Executor for settings action.
type SettingsExecutor struct {
	c      internal.Context
	logger logger.Logger
}

// NewSettingsExecutor builds executor for settings action.
func NewSettingsExecutor(c internal.Context, logger logger.Logger) SettingsExecutor {
	return SettingsExecutor{
		c:      c,
		logger: logger,
	}
}

// UpdateReadme updates the README.md using the documentation helper.
func (s SettingsExecutor) UpdateReadme(documentation string) {
	execs.NewUpdateReadmeExec(s.c, s.logger).Update(documentation)
}

// EnableSecureMode (re)encrypts data on disk with the new {secret}.
func (s SettingsExecutor) EnableSecureMode(secret string) bool {
	return execs.NewSecureModeExec(s.c, s.logger).Encrypt(secret)
}

// DisableSecureMode decrypts data on disk.
func (s SettingsExecutor) DisableSecureMode() bool {
	return execs.NewSecureModeExec(s.c, s.logger).Decrypt()
}
