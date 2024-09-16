package opentelemetry

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"tgs-automation/internal/middleware"
	"tgs-automation/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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

func InitialOpenTelemetry(ctx context.Context, config util.TgsConfig) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider) {
	serviceName := os.Getenv("SERVICE_NAME")
	// #region Start OpenTelemetry
	tp, err := initialTracerProvider(ctx, config.OpenTelemetryConfig.OtelCollectorUrl, serviceName, "0.1.0", os.Getenv("ENV"))

	if err != nil {
		log.Printf("Failed to init otel tracer:%v", err)
		return nil, nil
	}

	mp, err := initialPrometheusExporter()
	if err != nil {
		log.Printf("Failed to init otel meter: %v", err)
		return nil, nil
	}
	// #endregion End OpenTelemetry

	routerOtel := gin.Default()
	routerOtel.GET("/metrics", gin.WrapH(promhttp.Handler()))
	routerOtel.Use(otelgin.Middleware(serviceName))
	routerOtel.Use(middleware.MetricsMiddleware(mp))

	server := &http.Server{
		Addr:    ":8889",
		Handler: routerOtel,
	}

	err = server.ListenAndServe()

	if err != nil {
		fmt.Println("Failed to start otel server:", err)
	}

	return tp, mp
}

func initialTracerProvider(ctx context.Context, collectorUrl, service, version, environment string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(collectorUrl),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("OTLP Trace http Creation: %s %v", service, err)
		log.Fatal(err)
		return nil, nil
	}

	resources := getResource(service, version, environment)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func getResource(
	serviceName string,
	serviceVersion string,
	deploymentEnvironment string) *resource.Resource {
	// Defines resource with service name, version, and environment.
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		attribute.String("environment", deploymentEnvironment),
	)
}

// 初始化 Prometheus 导出器
func initialPrometheusExporter() (*sdkmetric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("无法创建 Prometheus 导出器：%v", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(os.Getenv("SERVICE_NAME")), // 设置自定义服务名称
		),
	)
	if err != nil {
		log.Fatalf("无法创建资源: %v", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	return meterProvider, nil
}
