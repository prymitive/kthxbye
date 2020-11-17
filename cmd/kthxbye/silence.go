package main

import (
	"context"
	"time"

	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Msg("Silence update failed")
	}

	metricsSincesExtended.Inc()
	log.Info().
		Str("id", *sil.ID).
		Str("replacedBy", postOk.Payload.SilenceID).
		Strs("matchers", silenceMatchersToLogField(sil)).
		Msg("Silence updated")
}
