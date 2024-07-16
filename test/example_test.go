package test_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/teivah/onecontext"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var _ = Describe("Example", func() {
	It("request to microservice-a should succeed", func(specContext SpecContext) {
		ctx, cancel := onecontext.Merge(testCtx, specContext)
		DeferCleanup(cancel)

		response, err := otelhttp.Get(ctx, "http://localhost:8080")
		Expect(err).ToNot(HaveOccurred(), "error making request to microservice-a")
		DeferCleanup(func() {
			Expect(response.Body.Close()).To(Succeed(), "error closing response body")
		})

		Expect(response).To(HaveHTTPStatus(http.StatusOK), "request to microservice-a should return 200 OK")
	}, SpecTimeout(5*time.Second))

	It("this test always passes", func() {
		Expect(true).To(BeTrue(), "should always be true")
	})
})
