package httputil

import (
	"github.com/joakim-ribier/gcli-4postman/internal/postman"
	"github.com/joakim-ribier/go-utils/pkg/httpsutil"
)

// Call executes collection {item} API request on a specific environment and returns the {httpsutil.HttpResponse}.
func Call(item postman.Item, env *postman.Env, params []postman.Param) (*httpsutil.HttpResponse, error) {
	r, err := httpsutil.NewHttpRequest(item.Request.Url.Get(env, params), item.Request.Body.Get(env, params))
	if err != nil {
		return nil, err
	}

	if item.Request.Auth.Type == "basic" {
		username, password := item.Request.Auth.GetAuthBasicCredential(env, params)
		r.SetBasicAuth(username, password)
	}

	return r.
		Method(item.Request.Method).
		Headers(item.Request.Header.Get(env, params)).
		AsJson().Call()
}
