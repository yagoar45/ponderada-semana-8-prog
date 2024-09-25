package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestTracerInitialization(t *testing.T) {
	cleanup := initTracer()
	defer cleanup(context.Background())

	tp := otel.GetTracerProvider()
	assert.NotNil(t, tp, "TracerProvider não deve ser nulo")

	tpSdk, ok := tp.(*trace.TracerProvider)
	assert.True(t, ok, "Deve ser uma instância de *trace.TracerProvider")
	assert.NotNil(t, tpSdk, "TracerProvider SDK não deve ser nulo")

	propagator := otel.GetTextMapPropagator()
	assert.NotNil(t, propagator, "TextMapPropagator não deve ser nulo")
}
