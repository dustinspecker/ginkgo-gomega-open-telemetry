package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"

	"github.com/dustinspecker/ginkgo-gomega-open-telemetry/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	tracerProvider, err := internal.GetTracerProvider("microservice-b")
	if err != nil {
		log.Fatalf("failed to get tracer: %v", err)
	}
	// Shutdown tracerProvider at the end of main since log.Fatalf is used when
	// when listener stops

	otel.SetTracerProvider(tracerProvider)

	prop := internal.GetPropagator()
	otel.SetTextMapPropagator(prop)

	mux := http.NewServeMux()

	mux.Handle("/", otelhttp.WithRouteTag("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rand.Intn(2) == 1 {
			span := trace.SpanFromContext(r.Context())
			span.RecordError(errors.New("random error"))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		response := Response{
			Message: "Hello, World!",
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonResponse)
	})))

	err = http.ListenAndServe(":8081", otelhttp.NewHandler(mux, "server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents)))

	tracerProvider.Shutdown(context.Background())

	// err should never be nil since ListenAndServe always returns a non-nil error
	if err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}
