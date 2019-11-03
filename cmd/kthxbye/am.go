package main

import (
	"github.com/prometheus/alertmanager/api/v2/client"

	clientruntime "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

func newAMClient(hostPort string, apiPath string) *client.Alertmanager {
	schemes := []string{"http"}
	cr := clientruntime.New(hostPort, apiPath, schemes)

	return client.New(cr, strfmt.Default)
}
