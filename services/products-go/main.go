package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type AppServer struct {
	server   *http.Server
	tracer   trace.Tracer
	meter    metric.Meter
	rootCtx  context.Context
	shutdown func(ctx context.Context) error
}

func NewAppServer(port string, rootCtx context.Context) *AppServer {
	server := &http.Server{
		Addr:         ":" + port,
		BaseContext:  func(_ net.Listener) context.Context { return rootCtx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}

	return &AppServer{
		server:  server,
		tracer:  otel.GetTracerProvider().Tracer("products-service"),
		meter:   otel.GetMeterProvider().Meter("products-service"),
		rootCtx: rootCtx,
	}
}

func (app *AppServer) Start() {
	// Set up OpenTelemetry
	cleanup, err := setupOTelSDK(app.rootCtx)
	if err != nil {
		log.Fatalf("failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		fmt.Println("cleaning up")
		if err := cleanup(context.Background()); err != nil {
			log.Fatalf("failed to cleanup OpenTelemetry: %v", err)
		}
	}()

	// Start the server
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", app.server.Addr)
		serverErr <- app.server.ListenAndServe()
	}()

	select {
	case <-app.rootCtx.Done():
		// Waiting for CTRL+C signal
		app.Stop()
		return
	case err := <-serverErr:
		log.Fatalf("server error: %s", err.Error())
		return
	}
}

func (app *AppServer) Stop() {
	// Gracefully shutdown the server, waiting max 30 seconds for current operations to complete.
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	app.server.Shutdown(ctx)
	log.Println("server gracefully stopped")
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	// handle is a replacement for mux.Handle
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handle := func(pattern string, handler http.Handler) {
		// Configure the "http.route" for the HTTP instrumentation.
		wrappedHandler := otelhttp.WithRouteTag(pattern, handler)
		mux.Handle(pattern, wrappedHandler)
	}

	// Register handlers.
	handle("GET /products", newRandomLatencyMW(0, 500, newProbabilisticFailureMW(0.1, http.HandlerFunc(productsHandler))))

	handle("GET /products/{id}", newRandomLatencyMW(0, 200, newProbabilisticFailureMW(0.1, http.HandlerFunc(getProductByIDHandler))))

	handle("POST /products", newRandomLatencyMW(0, 300, newProbabilisticFailureMW(0.2, http.HandlerFunc(addProductHandler))))

	handle("GET /might-fail", newProbabilisticFailureMW(0.1, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})))

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}

// func main() {
// 	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
// 	defer stop()

// 	cleanup, err := setupOTelSDK(ctx)
// 	if err != nil {
// 		log.Fatalf("failed to initialize OpenTelemetry: %v", err)
// 	}
// 	defer func() {
// 		if err := cleanup(context.Background()); err != nil {
// 			log.Fatalf("failed to cleanup OpenTelemetry: %v", err)
// 		}
// 	}()

// 	port := os.Getenv("APP_PORT")
// 	if port == "" {
// 		port = "8080"
// 	}
// 	srv := &http.Server{
// 		Addr:         ":" + port,
// 		BaseContext:  func(_ net.Listener) context.Context { return ctx },
// 		ReadTimeout:  time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		Handler:      newHTTPHandler(),
// 	}

// 	serverErr := make(chan error, 1)
// 	go func() {
// 		log.Printf("listening on :%s", port)
// 		serverErr <- srv.ListenAndServe()
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		// Waiting for CTRL+C signal
// 		// Gracefully shutdown the server, waiting max 30 seconds for current operations to complete.
// 		log.Println("shutting down server...")
// 		stop()
// 		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 		defer cancel()
// 		srv.Shutdown(ctx)
// 		log.Println("server gracefully stopped")
// 		return
// 	case err := <-serverErr:
// 		log.Fatalf("server error: %s", err.Error())
// 		return
// 	}

// }

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	server := NewAppServer(port, ctx)
	server.Start()
}
