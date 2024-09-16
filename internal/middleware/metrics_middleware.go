package middleware

import (
	"fmt"
	"os"
	"tgs-automation/internal/helper"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	requestTotal     metric.Int64Counter
	requestUVTotal   metric.Int64Counter
	uriRequestTotal  metric.Int64Counter
	requestBodySize  metric.Float64Counter
	responseBodySize metric.Float64Counter
	requestDuration  metric.Float64Histogram

	slowRequest     metric.Int64Counter
	defaultSlowTime = int32(5)
	defaultDuration = []float64{0.1, 0.3, 0.5, 1.2, 5, 10, 30, 60}
	bloomFilter     *helper.BloomFilter
)

func initial(mp *sdkmetric.MeterProvider) {
	meter := mp.Meter(os.Getenv("SERVICE_NAME"))
	bloomFilter = helper.NewBloomFilter()
	requestTotal, _ = meter.Int64Counter("gin_request_total", metric.WithDescription("all the server received request total."))
	requestUVTotal, _ = meter.Int64Counter("gin_request_uv_total", metric.WithDescription("all the server received ip number."))
	uriRequestTotal, _ = meter.Int64Counter("gin_uri_request_total", metric.WithDescription("all the server received request number with erery uri."))
	requestBodySize, _ = meter.Float64Counter("gin_request_body_total", metric.WithDescription("the server received request body size, unit is byte."))
	responseBodySize, _ = meter.Float64Counter("gin_response_body_total", metric.WithDescription("the server response body size, unit is byte."))

	requestDuration, _ = meter.Float64Histogram(
		"gin_request_duration",
		metric.WithUnit("s"),
		metric.WithDescription("the time server took to handle the request."),
		metric.WithExplicitBucketBoundaries(defaultDuration...),
	)

	slowRequest, _ = meter.Int64Counter("gin_slow_request_total", metric.WithDescription(fmt.Sprintf("the server handled slow requests counter, t=%d.", defaultSlowTime)))
}

// authMiddleware validates the JWT token
func MetricsMiddleware(mp *sdkmetric.MeterProvider) gin.HandlerFunc {
	initial(mp)
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		handleMetrics(c, startTime)
	}
}

func handleMetrics(ctx *gin.Context, start time.Time) {
	r := ctx.Request
	w := ctx.Writer
	c := ctx.Request.Context()

	attributes := metric.WithAttributes(
		attribute.String("uri", ctx.FullPath()),
		attribute.String("method", r.Method),
		attribute.String("code", string(w.Status())),
	)

	requestTotal.Add(c, 1)
	uriRequestTotal.Add(c, 1, attributes)

	//請求 IP 數
	if ip := ctx.ClientIP(); !bloomFilter.Contains(ip) {
		bloomFilter.Add(ip)
		requestUVTotal.Add(c, 1)
	}

	if r.ContentLength >= 0 {
		requestBodySize.Add(c, float64(r.ContentLength))
	}

	// set response size
	if w.Size() > 0 {
		responseBodySize.Add(c, float64(w.Size()))
	}

	// set slow request
	latency := time.Since(start)
	if int32(latency.Seconds()) >= defaultSlowTime {
		slowRequest.Add(c, 1, attributes)
	}

	// set request duration
	requestDuration.Record(c, latency.Seconds())
}
