package restyprometheus

import (
	"fmt"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
	"resty.dev/v3"
)

type metrics struct {
	responseTimePerURL    *prometheus.HistogramVec
	responseStatusCounter *prometheus.CounterVec
	successCounterByURL   *prometheus.CounterVec
	failureCounterByURL   *prometheus.CounterVec
}

func NewClient(c *resty.Client, ns string, serviceNamePrefix string) *resty.Client {
	m := &metrics{
		responseTimePerURL: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: ns,
				Name:      serviceNamePrefix + "response_time_seconds",
				Help:      "Response time of requests",
			},
			[]string{"url"},
		),
		responseStatusCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      serviceNamePrefix + "http_requests_status_total",
				Help:      "Total requests per URL, status code and method",
			},
			[]string{"status", "url", "method"},
		),
		successCounterByURL: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      serviceNamePrefix + "ttp_requests_total",
				Help:      "Total requests per URL",
			},
			[]string{"url"},
		),
		failureCounterByURL: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      serviceNamePrefix + "http_requests_failure_total",
				Help:      "Total failure requests per URL",
			},
			[]string{"url"},
		),
	}

	prometheus.MustRegister(
		m.responseTimePerURL,
		m.responseStatusCounter,
		m.successCounterByURL,
		m.failureCounterByURL,
	)
	c.AddResponseMiddleware(func(c *resty.Client, r *resty.Response) error {
		m.collect(r)
		return nil
	})

	return c
}

func (m *metrics) collect(r *resty.Response) {
	// ошибку глушим, так как это метрики и не должны влиять на итоговый результирующий код
	URL, _ := stripQueryKeep(r.Request.URL, []string{})
	if r.IsSuccess() {
		m.successCounterByURL.With(prometheus.Labels{"url": URL}).Inc()
	}
	if r.IsError() {
		m.failureCounterByURL.With(prometheus.Labels{"url": URL}).Inc()
	}

	m.responseTimePerURL.With(prometheus.Labels{"url": URL}).Observe(r.Duration().Seconds())
	m.responseStatusCounter.With(prometheus.Labels{
		"url":    URL,
		"status": fmt.Sprintf("%d", r.RawResponse.StatusCode),
		"method": r.Request.Method,
	}).Inc()
}

// stripQueryKeep убираем все гет параметры, кроме переданных в keep, чтобы сохранить преемственность запросов
// иначе запросы с условным параметром "query" который каждый раз уникальный засрут все метрики
func stripQueryKeep(u string, keep []string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	q := parsed.Query()          // все GET параметры
	newQ := url.Values{}         // новые параметры
	keepSet := map[string]bool{} // делаем set для удобства

	for _, k := range keep {
		keepSet[k] = true
	}

	// переносим только нужные
	for k, v := range q {
		if keepSet[k] {
			newQ[k] = v
		}
	}

	parsed.RawQuery = newQ.Encode()
	return parsed.String(), nil
}
