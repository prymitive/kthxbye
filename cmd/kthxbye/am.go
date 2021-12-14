package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

func newAMClient(uri string) http.Client {
	client := http.Client{}

	u, _ := url.Parse(uri)
	if u.User.Username() != "" {
		username := u.User.Username()
		password, _ := u.User.Password()
		client.Transport = setAuth(client.Transport, username, password)
	}

	return client
}

func joinURI(base, path string) string {
	if strings.HasSuffix(base, "/") {
		return fmt.Sprintf("%s%s", base, path)
	}
	return fmt.Sprintf("%s/%s", base, path)
}
