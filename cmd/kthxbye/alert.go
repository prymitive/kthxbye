package main

import (
	"context"

	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/models"
)

func queryAlerts(ctx context.Context, cfg *ackConfig) ([]*models.GettableAlert, error) {

	alerts := []*models.GettableAlert{}

	withUnprocessed := false
	withActive := false
	withInhibited := false
	withSilenced := true

	alertParams := alert.NewGetAlertsParams().WithContext(ctx).
		WithUnprocessed(&withUnprocessed).
		WithActive(&withActive).
		WithInhibited(&withInhibited).
		WithSilenced(&withSilenced)

	amclient := newAMClient(cfg.alertmanagerURI)

	getOk, err := amclient.Alert.GetAlerts(alertParams)

	if err != nil {
		return alerts, err
	}

	for _, alert := range getOk.Payload {
		alerts = append(alerts, alert)
	}

	return alerts, nil
}
