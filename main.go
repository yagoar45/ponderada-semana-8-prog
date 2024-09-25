package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// Configura variáveis de ambiente
var (
	serviceName  = os.Getenv("SERVICE_NAME")
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure     = os.Getenv("INSECURE_MODE")
)

// Função para inicializar o OpenTelemetry e configurar o tracer
func initTracer() func(context.Context) error {
	// Define se a conexão com o coletor é segura ou insegura
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	// Cria o exportador para enviar os dados de rastreamento para o SigNoz
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)
	if err != nil {
		log.Fatalf("Falha ao criar o exportador: %v", err)
	}

	// Define os recursos do tracer
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Falha ao definir recursos: %v", err)
	}

	// Configura o tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Retorna a função de encerramento do tracer
	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}
}

func main() {
	// Inicializa o tracer
	cleanup := initTracer()
	defer cleanup(context.Background())

	// Configura o Gin com middleware do OpenTelemetry
	r := gin.Default()
	r.Use(otelgin.Middleware(serviceName))

	// Define a rota e o handler
	r.GET("/hello", func(c *gin.Context) {
		ctx := c.Request.Context()
		tracer := otel.Tracer("hello-tracer")

		_, span := tracer.Start(ctx, "helloHandler")
		defer span.End()

		span.SetAttributes(attribute.String("handler", "hello"))
		c.String(http.StatusOK, "Hello, OpenTelemetry with Gin!")
	})

	// Executa o servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor iniciando na porta %s", port)
	log.Fatal(r.Run(":" + port))
}
