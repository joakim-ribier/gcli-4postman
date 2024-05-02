package execs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joakim-ribier/gcli-4postman/internal"
	"github.com/joakim-ribier/gcli-4postman/pkg/logger"
)

type UpdateReadmeExec struct {
	c      internal.Context
	logger logger.Logger
}

func NewUpdateReadmeExec(c internal.Context, logger logger.Logger) UpdateReadmeExec {
	return UpdateReadmeExec{
		c:      c,
		logger: logger,
	}
}

// Update updates the README.md using the documentation helper.
func (u UpdateReadmeExec) Update(documentation string) {
	file, err := os.ReadFile("../../README.md")
	if err != nil {
		u.logger.Error(err, "file cannot be read", "resource", "../../README.md")
		u.c.Print("ERROR", err.Error())
		return
	}

	repl := regexp.
		MustCompile(`(?s)(#how-to-use#).*?(#how-to-use#)`).
		ReplaceAllString(string(file[:]), fmt.Sprintf("#how-to-use#\n%s\n#how-to-use#", documentation))

	if err := os.WriteFile("../../README.md", []byte(repl), 0644); err != nil {
		u.logger.Error(err, "file cannot be written", "resource", "../../README.md")
		u.c.Print("ERROR", err.Error())
		return
	}

	u.c.Print("INFO", "README.md updated!")
}
