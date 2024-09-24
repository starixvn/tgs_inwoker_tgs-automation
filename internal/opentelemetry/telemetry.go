package opentelemetry

import (
	"context"
	"log"
	"net/http"
	"os"

	"tgs-automation/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

// Telemetry 結構體封裝了所有 OpenTelemetry 相關的邏輯
type Telemetry struct {
	Config         config.TgsConfig
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Server         *http.Server
}

// NewTelemetry 是 Telemetry 的構造函數
func NewTelemetry(config config.TgsConfig) *Telemetry {
	return &Telemetry{
		Config: config,
	}
}

// Initialize 對 OpenTelemetry 進行初始化，包括 Trace 和 Metric
func (t *Telemetry) Initialize(ctx context.Context, port string) error {
	serviceName := os.Getenv("SERVICE_NAME")

	// 初始化 TracerProvider
	tp, err := t.initTracerProvider(ctx, t.Config.OpenTelemetryConfig.OtelCollectorUrl, serviceName, "0.1.0", os.Getenv("ENV"))
	if err != nil {
		log.Printf("Failed to initialize TracerProvider: %v", err)
		return err
	}
	t.TracerProvider = tp

	// 初始化 PrometheusExporter
	mp, err := t.initPrometheusExporter()
	if err != nil {
		log.Printf("Failed to initialize MeterProvider: %v", err)
		return err
	}
	t.MeterProvider = mp

	// 初始化 Gin 並監聽 Prometheus 指標路徑
	routerOtel := gin.Default()
	routerOtel.GET("/metrics", gin.WrapH(promhttp.Handler()))
	t.Server = &http.Server{
		Addr:    port,
		Handler: routerOtel,
	}

	// 啟動 Prometheus Server
	go func() {
		if err := t.Server.ListenAndServe(); err != nil {
			log.Println("Failed to start Prometheus server:", err)
		}
	}()

	return nil
}

// initTracerProvider 初始化 TracerProvider
func (t *Telemetry) initTracerProvider(ctx context.Context, collectorUrl, service, version, environment string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(collectorUrl),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create OTLP Trace HTTP exporter: %v", err)
		return nil, err
	}

	resources := t.getResource(service, version, environment)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

// getResource 創建 OpenTelemetry Resource
func (t *Telemetry) getResource(serviceName, serviceVersion, deploymentEnvironment string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		attribute.String("environment", deploymentEnvironment),
	)
}

// initPrometheusExporter 初始化 Prometheus Exporter
func (t *Telemetry) initPrometheusExporter() (*sdkmetric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(os.Getenv("SERVICE_NAME")),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

// Shutdown 關閉所有的 OpenTelemetry 組件
func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.Server != nil {
		if err := t.Server.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
