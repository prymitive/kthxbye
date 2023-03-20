package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	metricsCycles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kthxbye_cycles_total",
		Help: "The total number of silence check cycles",
	})
	metricsCycleFailrues = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kthxbye_cycles_failed_total",
		Help: "The total number of failed cycles",
	})
	metricsCycleStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kthxbye_last_cycle_success",
		Help: "Status of the last cycle, 0: failure, 1: success",
	})
	metricsSincesExtended = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kthxbye_silences_extended_total",
		Help: "The total number of silence that got extended",
	})
	metricsSilencesTracked = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kthxbye_silences_tracked",
		Help: "Current number of silences that match prefix pattern and will be tracked",
	})
	metricsSilencesExpiring = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kthxbye_silences_expiring",
		Help: "Current number of silences that are tracked but don't match any alert",
	})
)

type ackConfig struct {
	alertmanagerURI     string
	alertmanagerTimeout time.Duration
	loopInterval        time.Duration
	extendIfExpiringIn  time.Duration
	extendBy            time.Duration
	extendWithPrefix    string
	maxDuration         time.Duration
	logJSON             bool
}

func main() {
	addr := flag.String("listen", ":8080", "The address to listen on for HTTP requests.")

	cfg := ackConfig{}
	flag.StringVar(&cfg.alertmanagerURI, "alertmanager.uri", "http://localhost:9093", "Alertmanager URI to use")
	flag.DurationVar(&cfg.alertmanagerTimeout, "alertmanager.timeout", time.Minute, "Alertmanager request timeout")
	flag.DurationVar(&cfg.loopInterval, "interval", time.Duration(time.Second*45), "Silence check interval")
	flag.DurationVar(&cfg.extendIfExpiringIn, "extend-if-expiring-in", time.Duration(time.Minute*5), "Extend silences that are about to expire in the next DURATION seconds")
	flag.DurationVar(&cfg.extendBy, "extend-by", time.Duration(time.Minute*15), "Extend silences by adding DURATION seconds")
	flag.StringVar(&cfg.extendWithPrefix, "extend-with-prefix", "ACK!", "Extend silences with comment starting with PREFIX string")
	flag.DurationVar(&cfg.maxDuration, "max-duration", 0, "Maximum duration of a silence, it won't be extended anymore after reaching it")
	flag.BoolVar(&cfg.logJSON, "log-json", false, "Format logged messages as JSON")

	flag.Parse()

	if !cfg.logJSON {
		setupLogger()
	}

	if cfg.extendBy.Seconds() <= cfg.extendIfExpiringIn.Seconds() {
		log.Fatal().Msg("-extend-by value must be greater than -extend-if-expiring-in")
	}

	go ackLoop(cfg)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", index)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}

func index(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte(`
			<html>
			<head><title>kthxbye</title></head>
			<body>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>
			`))
	if err != nil {
		log.Error().Err(err).Msg("Failed to write a response")
	}
}
