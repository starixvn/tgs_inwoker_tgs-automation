package opentelemetry

import (
	"context"
	"log"

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

func InitTracerProvider(ctx context.Context, collectorUrl, service, version, environment string) (*sdktrace.TracerProvider, error) {
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

func InitMetricProvider(ctx context.Context, collectorUrl, serviceName string) (*sdkmetric.MeterProvider, error) {
	// secureOption := otlpmetricgrpc.WithInsecure()
	// exporter, err := otlpmetricgrpc.New(ctx, secureOption, otlpmetricgrpc.WithEndpoint(collectorUrl))

	// if err != nil {
	// 	return nil, err
	// }

	// resource := resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))

	// mp := sdkmetric.NewMeterProvider(
	// 	sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))),
	// 	sdkmetric.WithResource(resource),
	// )

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("无法创建 Prometheus 导出器：%v", err)
	}

	mp := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(mp)
	return mp, nil
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

// func InitPrometheusExporter() {
// 	exporter, err := prometheus.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 創建全局 MeterProvider
// 	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
// 	otel.SetMeterProvider(meterProvider)

// 	// 暴露 /metrics 端點
// 	http.Handle("/metrics", promhttp.Handler())
// 	err = http.ListenAndServe("8888", nil)

// 	if err != nil {
// 		log.Fatalf("启动指标服务器失败: %v", err)
// 	}

// 	runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
// 	host.Start()

// 	return meterProvider, nil
// }

// 初始化 Prometheus 导出器
func InitPrometheusExporter() {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatalf("无法创建 Prometheus 导出器：%v", err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(meterProvider)
}
