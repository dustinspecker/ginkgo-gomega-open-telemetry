package test_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dustinspecker/ginkgo-gomega-open-telemetry/internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	// Note: test cases should not directly use suiteCtx or suiteSpan
	suiteCtx  context.Context
	suiteSpan trace.Span

	// Note: test cases may create children spans by providing testCtx to tracer.Start
	testCtx context.Context
	// Note: test cases may attach attributes to testSpan if desired through testSpan.SetAttributes
	testSpan trace.Span
)

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite")
}

// setup tracer provider, propagator, and create suite-level span
// each test case will have a child span of the suite-level span
var _ = BeforeSuite(func() {
	tracerProvider, err := internal.GetTracerProvider("test")
	Expect(err).NotTo(HaveOccurred(), "error creating tracer provider")
	DeferCleanup(func() {
		Expect(tracerProvider.Shutdown(context.Background())).To(Succeed(), "error shutting down tracer provider")
	})
	otel.SetTracerProvider(tracerProvider)

	propagator := internal.GetPropagator()
	otel.SetTextMapPropagator(propagator)

	suiteCtx, suiteSpan = tracerProvider.Tracer("test").Start(context.Background(), "test-suite")
	DeferCleanup(func() {
		suiteSpan.End()
	})
})

// create a span for the test case
var _ = BeforeEach(func() {
	testCtx, testSpan = otel.GetTracerProvider().Tracer("test").Start(suiteCtx, CurrentSpecReport().FullText())
	DeferCleanup(func() {
		if CurrentSpecReport().Failed() {
			testSpan.SetStatus(codes.Error, CurrentSpecReport().Failure.Message)
		} else {
			testSpan.SetStatus(codes.Ok, "test passed")
		}

		testSpan.End()

		traceLink := fmt.Sprintf("http://localhost:16686/trace/%s", testSpan.SpanContext().TraceID().String())
		GinkgoWriter.Println("visit trace:", traceLink)
	})
})
