package api

import (
	_ "embed"
	"net/http"
)

//go:embed swagger.yaml
var swagger []byte

func OapiSchemaHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write(swagger)
}
