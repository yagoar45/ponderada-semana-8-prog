package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		ctx := c.Request.Context()
		tracer := otel.Tracer("hello-tracer")

		_, span := tracer.Start(ctx, "helloHandler")
		defer span.End()

		span.SetAttributes(attribute.String("handler", "hello"))
		c.String(http.StatusOK, "Hello, OpenTelemetry with Gin!")
	})
	return r
}

func TestHelloHandler(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/hello", nil)
	if err != nil {
		t.Fatalf("Erro ao criar a solicitação: %v", err)
	}

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello, OpenTelemetry with Gin!", w.Body.String())
}
