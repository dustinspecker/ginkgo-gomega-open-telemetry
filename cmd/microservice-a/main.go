package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dustinspecker/ginkgo-gomega-open-telemetry/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type Response struct {
	Message string `json:"message"`
}

func getMessage(ctx context.Context) (string, error) {
	url := "http://localhost:8081/"
	jsonMessageResponse, err := otelhttp.Get(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	defer jsonMessageResponse.Body.Close()

	if jsonMessageResponse.StatusCode >= 300 {
		return "", fmt.Errorf("failed request %s: %s", url, jsonMessageResponse.Status)
	}

	response := Response{}
	if err := json.NewDecoder(jsonMessageResponse.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed decoding response: %w", err)
	}

	return response.Message, nil
}

func main() {
	tracerProvider, err := internal.GetTracerProvider("microservice-a")
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
		message, err := getMessage(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := Response{
			Message: fmt.Sprintf("message from microservice-a: %s", message),
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(jsonResponse)
	})))

	err = http.ListenAndServe(":8080", otelhttp.NewHandler(mux, "server", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents)))

	tracerProvider.Shutdown(context.Background())

	// err should never be nil since ListenAndServe always returns a non-nil error
	if err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}
