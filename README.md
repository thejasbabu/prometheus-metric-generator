# Prometheus metric generator

A HTTP server which exposes the metrics which are configured in the metric.config file and also exposes api to override the metric value

## Config file structure

```
metrics:
  - name: test_metric
    help: Testing metric used for test
    type: gauge
  - name: another_test_metric
    help: Another testing metric used for test
    type: gauge
```

## Usage

The below endpoint emits all the metrics exposed

```
curl localhost:8080/metrics
```

The below endpoint allows one to override the metric value

```
curl -X POST \
  http://localhost:8080/metric \
  -H 'Content-Type: application/json' \
  -H 'Host: localhost:8080' \
  -d '{
	"name": "another_test_metric",
	"value": 1232
  }'
```