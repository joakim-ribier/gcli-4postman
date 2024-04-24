package postman

import "github.com/gosimple/slug"

type EnvParam struct {
	Key   string
	Value string
}

type Env struct {
	Name   string
	Params []EnvParam `json:"values"`
}

func (e Env) GetName() string {
	return slug.Make(e.Name)
}

func NewEnv() Env {
	return Env{
		Name:   "",
		Params: []EnvParam{},
	}
}
