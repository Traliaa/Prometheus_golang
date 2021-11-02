package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
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
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	go func() {

		r.GET("/metrics", prometheusHandler())
		r.Run(":84")
	}()

	r.GET("/", func(c *gin.Context) {
		startTime := time.Now()

		defer func() {
			Histogram.Observe(float64(time.Since(startTime)) / float64(time.Second))
			Counter.Inc()
		}()
		n := rand.Intn(10)
		time.Sleep(time.Duration(n) * time.Second)
		c.JSON(200, gin.H{
			"status": "ok",
		})
		logger.Error(fmt.Sprintf("Method: %s,  RequestURI: %s,  ResponseTime: %.5f", c.Request.Method, c.Request.RequestURI, float64(time.Since(startTime))/float64(time.Second)))
		sentry.CaptureMessage(fmt.Sprintf("Method: %s,  RequestURI: %s,  ResponseTime: %.5f", c.Request.Method, c.Request.RequestURI, float64(time.Since(startTime))/float64(time.Second)))

	})
	err = r.Run(":83")
	if err != nil {
		log.Fatal(err)
	}

}
