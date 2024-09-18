package opentelemetry

import (
	"context"
	"log"
	"net/http"
	"os"

	"tgs-automation/internal/util"

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

func InitialOpenTelemetry(ctx context.Context, config util.TgsConfig, port string) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider) {
	serviceName := os.Getenv("SERVICE_NAME")
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

	routerOtel := gin.Default()
	routerOtel.GET("/metrics", gin.WrapH(promhttp.Handler()))
	server := &http.Server{
		Addr:    port,
		Handler: routerOtel,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("Err", err)
			panic(err)
		}
	}()

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
