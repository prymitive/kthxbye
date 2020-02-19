package main

import (
	"net/http"
	"net/url"
	"path"

	httptransport "github.com/go-openapi/runtime/client"

	"github.com/prometheus/alertmanager/api/v2/client"
)

func setAuth(inner http.RoundTripper, username string, password string) http.RoundTripper {
	return &authRoundTripper{
		inner:    inner,
		Username: username,
		Password: password,
	}
}

type authRoundTripper struct {
	inner    http.RoundTripper
	Username string
	Password string
}

func (art *authRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(art.Username, art.Password)
	return art.inner.RoundTrip(r)
}

func newAMClient(uri string) *client.Alertmanager {
	u, _ := url.Parse(uri)

	transport := httptransport.New(u.Host, path.Join(u.Path, "/api/v2"), []string{u.Scheme})

	if u.User.Username() != "" {
		username := u.User.Username()
		password, _ := u.User.Password()
		transport.Transport = setAuth(transport.Transport, username, password)
	}

	c := client.New(transport, nil)
	return c
}
