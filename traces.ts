// package main

// import (
// 	"context"
// 	"fmt"

// 	"go.opentelemetry.io/otel"
// 	// "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
// 	// "go.opentelemetry.io/otel/sdk/resource"
// 	// "go.opentelemetry.io/otel/sdk/trace"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	// semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
// 	// "go.opentelemetry.io/otel/trace"
// )

// var tp *sdktrace.TracerProvider

// func initTracer() error{
// 	// exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
// 	ctx := context.Background()
// 	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
// 	if err != nil {
// 		return fmt.Errorf("Failed to initialize console exporter %w", err)
// 	}

// 	bsp := sdktrace.NewBatchSpanProcessor(exp)

// 	tp = sdktrace.NewTracerProvider(
// 		sdktrace.WithSampler(sdktrace.AlwaysSample()),
// 		sdktrace.WithSpanProcessor(bsp),
// 	)

// 	otel.SetTracerProvider(tp)
// 	return nil
// }

