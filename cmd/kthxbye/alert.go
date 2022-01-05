package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Alert struct {
	Status struct {
		SilencedBy []string `json:"silencedBy"`
	} `json:"status"`
}

func queryAlerts(cfg ackConfig) (alerts []Alert, err error) {
	uri := fmt.Sprintf(
		"%s?silenced=true&inhibited=false&active=false&unprocessed=false",
		joinURI(cfg.alertmanagerURI, "api/v2/alerts"))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.alertmanagerTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	client := newAMClient(cfg.alertmanagerURI)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if tok != json.Delim('[') {
		return nil, fmt.Errorf("invalid JSON token, expected [, got %s", tok)
	}

	var alert Alert
	for dec.More() {
		alert.Status.SilencedBy = []string{}
		if err = dec.Decode(&alert); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if tok != json.Delim(']') {
		return nil, fmt.Errorf("invalid JSON token, expected ], got %s", tok)
	}

	return alerts, nil
}
