package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsCycles = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kthxbye_cycles_total",
		Help: "The total number of silence check cycles",
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
	alertmanagerHostPort string
	alertmanagerAPIPath  string
	loopInterval         time.Duration
	extendIfExpiringIn   time.Duration
	extendBy             time.Duration
	extendWithPrefix     string
}

func main() {
	addr := flag.String("listen", ":8080", "The address to listen on for HTTP requests.")

	cfg := ackConfig{}
	flag.StringVar(&cfg.alertmanagerHostPort, "alertmanager.addr", "localhost:9093", "The address of the alertmanager")
	flag.StringVar(&cfg.alertmanagerAPIPath, "alertmanager.api", "/api/v2", "Base path for the alertmanager API")
	flag.DurationVar(&cfg.loopInterval, "interval", time.Duration(time.Second*45), "Silence check interval")
	flag.DurationVar(&cfg.extendIfExpiringIn, "extend-if-expiring-in", time.Duration(time.Minute*5), "Extend silences that are about to expire in the next DURATION seconds")
	flag.DurationVar(&cfg.extendBy, "extend-by", time.Duration(time.Minute*15), "Extend silences by adding DURATION seconds")
	flag.StringVar(&cfg.extendWithPrefix, "extend-with-prefix", "ACK!", "Extend silences with comment starting with PREFIX string")

	flag.Parse()

	if cfg.extendBy.Seconds() < cfg.extendIfExpiringIn.Seconds() {
		log.Fatal("-extend-by value must be greater than -extend-if-expiring-in")
	}

	go ackLoop(&cfg)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", index)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(`
			<html>
			<head><title>kthxbye</title></head>
			<body>
			<p><a href='/metrics'>Metrics</a></p>
			</body>
			</html>
			`))
	if err != nil {
		log.Println(err)
	}
}
