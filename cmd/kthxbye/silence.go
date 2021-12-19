package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type Silence struct {
	ID     string `json:"id,omitempty"`
	Status struct {
		State string `json:"state"`
	} `json:"status"`
	Matchers []struct {
		IsEqual bool   `json:"isEqual"`
		IsRegex bool   `json:"isRegex"`
		Name    string `json:"name"`
		Value   string `json:"value"`
	} `json:"matchers"`
	CreatedBy string    `json:"createdBy"`
	Comment   string    `json:"comment"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SilenceResponse struct {
	SilenceID string `json:"silenceID"`
}

func querySilences(cfg ackConfig) (silences []Silence, err error) {
	uri := joinURI(cfg.alertmanagerURI, "api/v2/silences")

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

	var silence Silence
	for dec.More() {
		if err = dec.Decode(&silence); err != nil {
			return nil, err
		}
		if silence.Status.State == "active" {
			silences = append(silences, silence)
		}
	}

	tok, err = dec.Token()
	if err != nil {
		return nil, err
	}
	if tok != json.Delim(']') {
		return nil, fmt.Errorf("invalid JSON token, expected ], got %s", tok)
	}

	_, _ = io.Copy(io.Discard, resp.Body)

	return silences, nil
}

func updateSilence(cfg ackConfig, sil Silence) error {
	payload, err := json.Marshal(sil)
	if err != nil {
		return err
	}

	uri := joinURI(cfg.alertmanagerURI, "api/v2/silences")

	client := newAMClient(cfg.alertmanagerURI)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.alertmanagerTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong response status received: %s", resp.Status)
	}

	var sr SilenceResponse
	err = json.NewDecoder(resp.Body).Decode(&sr)
	if err != nil {
		return err
	}

	metricsSincesExtended.Inc()
	log.Info().
		Str("id", sil.ID).
		Str("replacedBy", sr.SilenceID).
		Strs("matchers", silenceMatchersToLogField(sil)).
		Msg("Silence updated")
	return nil
}
