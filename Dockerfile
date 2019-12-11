FROM golang:1.13-alpine as builder
RUN mkdir -p /prometheus-metric-generator
COPY . /prometheus-metric-generator
WORKDIR /prometheus-metric-generator
RUN go build

FROM alpine
COPY --from=builder /prometheus-metric-generator/prometheus-metric-generator ./prometheus-metric-generator
CMD ./prometheus-metric-generator