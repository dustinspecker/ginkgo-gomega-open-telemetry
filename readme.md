# ginkgo-gomega-open-telemetry

> Example of using OpenTelemetry tracing through Ginkgo tests

## Usage

1. Start a Jaeger all-in-one instance:

   ```bash
   docker run \
     --interactive \
     --name jaeger \
     --rm \
     --tty \
     --env COLLECTOR_OTLP_ENABLED=true \
     --publish 0.0.0.0:16686:16686 \
     --publish 0.0.0.0:4317:4317 \
     --publish 0.0.0.0:44317:44317 \
     jaegertracing/all-in-one:1.35
   ```

1. Start an open-telemetry-collector instance:

   ```bash
   docker run \
     --interactive \
     --network container:jaeger \
     --rm \
     --tty \
     --volume "$PWD/open-telemetry-config.yaml:/open-telemetry-config.yaml" \
     otel/opentelemetry-collector-contrib:0.104.0 --config /open-telemetry-config.yaml
   ```

1. Start microservice-a:

   ```bash
   go run ./cmd/microservice-a
   ```

1. Start microservice-b:

   ```bash
   go run ./cmd/microservice-b
   ```

1. Install Ginkgo:

   ```bash
   go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
   ```

1. Run Ginkgo tests:

   ```bash
   ginkgo run -v ./test/...
   ```
