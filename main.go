package main

import (
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)
const (
    Namespace = "geekbrains"

    LabelMethod = "method"
    LabelStatus = "status"
)
var (
	Histogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Namespace: Namespace,
	Name:  	"latency",
	Help:  	"time series request",
	Buckets:   []float64{0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000},
})



Counter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: Namespace,
	Name:  	"request",
	Help:  	"request for '/'",
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

func main()  {
	

	go func ()  {
		r := gin.Default()
		r.GET("/metrics", prometheusHandler())
		r.Run(":84")
	}()


	r := gin.Default()
	
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
	})
	r.Run(":83")



}


