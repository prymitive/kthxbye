package main

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
)

func querySilences(ctx context.Context, cfg *ackConfig) ([]*models.GettableSilence, error) {
	silences := []*models.GettableSilence{}

	silenceParams := silence.NewGetSilencesParams().WithContext(ctx)

	amclient := newAMClient(cfg.alertmanagerURI)

	getOk, err := amclient.Silence.GetSilences(silenceParams)
	if err != nil {
		return silences, err
	}

	for _, sil := range getOk.Payload {
		if time.Time(*sil.EndsAt).Before(time.Now()) {
			continue
		}
		silences = append(silences, sil)
	}

	return silences, nil
}

func updateSilence(ctx context.Context, cfg *ackConfig, sil *models.GettableSilence) {
	ps := &models.PostableSilence{
		ID:      *sil.ID,
		Silence: sil.Silence,
	}

	amclient := newAMClient(cfg.alertmanagerURI)

	silenceParams := silence.NewPostSilencesParams().WithContext(ctx).WithSilence(ps)
	postOk, err := amclient.Silence.PostSilences(silenceParams)
	if err != nil {
		log.Printf("Silence update failed: %s", err)
	}

	metricsSincesExtended.Inc()
	log.Printf("Silence updated: %s => %s", *sil.ID, postOk.Payload.SilenceID)
}
