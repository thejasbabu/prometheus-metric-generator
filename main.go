package main

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


type Config struct {
	Port int64 `default:"8080"`
	ConfigFile string `default:"./metrics.config"`
}

type Metric struct {
	Name string `yaml:"name"`
	HelpDesc string `yaml:"help"`
	Type string `yaml:"type"`
}

type PrometheusMetrics struct {
	Name string
	Type string
	Metric interface{}
}

type MetricRequest struct {
	Name string `json:"name"`
	Value float64 `json:"value"`
}

var registry = prometheus.NewRegistry()
var prometheusMetrics []PrometheusMetrics

func main() {
	var config Config
	err := envconfig.Process("PMG", &config)
	if err != nil {
		log.Fatalf("error parsing env variables: %s", err.Error())
	}

	metrics, err := getMetrics(config.ConfigFile)
	if err!= nil {
		log.Fatalf("error reading metric from config file %s: %s", config.ConfigFile, err.Error())
	}

	for _, metric := range metrics {
		switch metric.Type {
		case "gauge":
			promMetric := initGaugeMetric(metric.Name, metric.HelpDesc)
			registry.MustRegister(promMetric)
			prometheusMetrics = append(prometheusMetrics, PrometheusMetrics{Name: metric.Name, Metric: &promMetric, Type: "gauge"})
		default:
			log.Printf("%s metric type not supported", metric.Type)
		}
	}

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/metric", handleMetricUpdate)
	addr := fmt.Sprintf(":%d", config.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func initGaugeMetric(name, helpDesc string) prometheus.Gauge {
	return prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
			Help: helpDesc,
		},
	)
}

type MetricConfig struct {
	Metrics []Metric `yaml:"metrics"`
}

func getMetrics(configFilePath string) ([]Metric, error) {
	content, err := ioutil.ReadFile(configFilePath)
	if err!= nil {
		return nil, err
	}
	var metricConfig MetricConfig
	err = yaml.Unmarshal(content, &metricConfig)
	if err!= nil {
		return nil, err
	}
	return metricConfig.Metrics, nil
}

func handleMetricUpdate(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var metricReq MetricRequest
	err := decoder.Decode(&metricReq)
	if err != nil {
		log.Printf("error parsing request: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	found := false
	for _, metric := range prometheusMetrics {
		if metric.Name == metricReq.Name {
			found = true
			switch metric.Type {
			case "gauge":
				promMetric, ok := metric.Metric.(*prometheus.Gauge)
				if !ok {
					log.Printf("error getting gauge metric: %s", metric.Name)
					w.WriteHeader(http.StatusInternalServerError)
				}
				(*promMetric).Set(metricReq.Value)
				w.WriteHeader(http.StatusOK)
			default:
				log.Printf("metric type not supported: %s", metric.Type)
				w.WriteHeader(http.StatusInternalServerError)
			}
			break
		}
	}
	if !found {
		log.Printf("metric not found: %s", metricReq.Name)
		w.WriteHeader(http.StatusNotFound)
	}
	return
}