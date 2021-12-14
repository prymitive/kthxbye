package main

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func extendACKs(cfg ackConfig) error {
	silences, err := querySilences(cfg)
	if err != nil {
		return err
	}

	alerts, err := queryAlerts(cfg)
	if err != nil {
		return err
	}

	metricsSilencesTracked.Set(float64(len(silences)))

	extendIfBefore := time.Now().UTC().Add(cfg.extendIfExpiringIn)

	silencesExpiring := 0
	for _, sil := range silences {
		if !strings.HasPrefix(sil.Comment, cfg.extendWithPrefix) {
			continue
		}

		usedBy := 0
		for _, alert := range alerts {
			for _, silenceID := range alert.Status.SilencedBy {
				if silenceID == sil.ID {
					usedBy++
				}
			}
		}
		if usedBy > 0 {
			if sil.EndsAt.Before(extendIfBefore) {
				duration := time.Time(sil.EndsAt).Sub(time.Time(sil.StartsAt))
				if cfg.maxDuration > 0 && duration > cfg.maxDuration {
					log.Info().
						Str("id", sil.ID).
						Strs("matchers", silenceMatchersToLogField(sil)).
						Str("maxDuration", cfg.maxDuration.String()).
						Msgf("Silence is used by %d alert(s) but it already reached the maximum duration, letting it expire", usedBy)
				} else {
					log.Info().
						Str("id", sil.ID).
						Strs("matchers", silenceMatchersToLogField(sil)).
						Msgf("Silence expires in %s and matches %d alert(s), extending it by %s",
							sil.EndsAt.Sub(time.Now().UTC()), usedBy, cfg.extendBy)
					sil.EndsAt = time.Now().UTC().Add(cfg.extendBy)
					err = updateSilence(cfg, sil)
					if err != nil {
						log.Error().Err(err).Msg("Silence update failed")
					}
				}
			}
		} else {
			log.Info().
				Str("id", sil.ID).
				Strs("matchers", silenceMatchersToLogField(sil)).
				Msg("Silence is not used by any alert, letting it expire")
			silencesExpiring++
		}
	}
	metricsSilencesExpiring.Set(float64(silencesExpiring))

	return nil
}

func ackLoop(cfg ackConfig) {
	metricsCycleStatus.Set(1)
	for {
		err := extendACKs(cfg)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process silences")
			metricsCycleFailrues.Inc()
			metricsCycleStatus.Set(0)
		} else {
			metricsCycleStatus.Set(1)
		}
		metricsCycles.Inc()
		time.Sleep(cfg.loopInterval)
	}
}
