package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmchi"
	"go.uber.org/zap"
)

const (
	Namespace = "geekbrains"

	LabelMethod = "method"
	LabelStatus = "status"
)

var (
	Histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: Namespace,
		Name:      "latency",
		Help:      "time series request",
		Buckets:   []float64{0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000},
	})

	Counter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Name:      "request",
		Help:      "request for '/'",
	})
)

func init() {

	prometheus.MustRegister(Histogram)
	prometheus.MustRegister(Counter)

}

func prometheusHandler() func(w http.ResponseWriter, r *http.Request) {
	h := promhttp.Handler()

	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	cfg := &config.Configuration{
		ServiceName: "ApiTraicer",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "localhost:6831",
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatal(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	logger = logger.With(zap.String("app", "Logger")).With(zap.String("environment", "production"))

	defer func() { _ = logger.Sync() }()

	err = sentry.Init(sentry.ClientOptions{
		Dsn: "https://73ceabf20bb04c2c80053701f2cb52e3@o1054564.ingest.sentry.io/6039999",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)
	r := chi.NewRouter()
	r.Use(apmchi.Middleware())
	go func() {
		r.Get("/metrics", prometheusHandler())
		http.ListenAndServe(":84", r)

	}()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		spanApm, _ := apm.StartSpan(r.Context(), "/getStatus", "request")

		span, _ := opentracing.StartSpanFromContextWithTracer(r.Context(), tracer, "/getStatus")
		startTime := time.Now()

		defer func() {
			Histogram.Observe(float64(time.Since(startTime)) / float64(time.Second))
			Counter.Inc()
		}()
		n := rand.Intn(1)
		time.Sleep(time.Duration(n) * time.Second)
		render.JSON(w, r, `{"status":"ok",}`)
		logger.Info(fmt.Sprintf("Method: %s,  RequestURI: %s,  ResponseTime: %.5f", r.Method, r.RequestURI, float64(time.Since(startTime))/float64(time.Second)))
		//sentry.CaptureMessage(fmt.Sprintf("Method: %s,  RequestURI: %s,  ResponseTime: %.5f", r.Method, r.RequestURI, float64(time.Since(startTime))/float64(time.Second)))
		span.Finish()
		spanApm.End()
	})
	http.ListenAndServe(":83", r)

}
