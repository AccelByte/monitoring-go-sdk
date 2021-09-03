/*
 * Copyright 2021 AccelByte Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type PrometheusClient struct {
	Handler http.Handler
	Metrics []Metric
}

func newPrometheus() Client {
	return &PrometheusClient{Handler: promhttp.Handler()}
}

func (client *PrometheusClient) SetMetrics(metrics []Metric) Client{
	log.Printf("\nsetting metrics %+v\n", metrics)
	client.Metrics = metrics
	return client
}

func (client *PrometheusClient) GetHandler() http.Handler{
	return client.Handler
}

func (client *PrometheusClient) Init() {
	for _, metric := range client.Metrics {
		prometheusJob := client.newJob(metric.MetricType, metric.Name, metric.Desc)
		go client.runJob(prometheusJob, metric.UpdateInterval, metric.OperationType, metric.Getter)
	}
}

func (client *PrometheusClient) newJob(t int, name, desc string) job {
	switch t {
	case Gauge:
		return &GaugeJob{promauto.NewGauge(prometheus.GaugeOpts{Name:name, Help: desc})}
	case Counter:
		return &CounterJob{promauto.NewCounter(prometheus.CounterOpts{Name: name, Help: desc})}
	}
	return nil
}

func (client *PrometheusClient) runJob(job job, interval time.Duration, operationType int, getter func() float64){
	ticker := time.NewTicker(interval)
	for{
		select {
		case <-ticker.C:
			value := float64(0)
			if getter != nil {
				value = getter()
			}

			// get operation and run it directly
			client.getOperation(operationType, job, value)()
		}
	}
}

func (client *PrometheusClient) getOperation(operationType int, prometheusJob job, value float64) func(){
	switch operationType {
	case Set:
		return func() { prometheusJob.Set(value) }
	case Increment:
		return func() { prometheusJob.Increment() }
	case Decrement:
		return func() { prometheusJob.Decrement() }
	case Add:
		return func() { prometheusJob.Add(value) }
	case Sub:
		return func() { prometheusJob.Subtract(value) }
	default:
		log.Fatal("unknown operation type")
	}
	return nil
}


type job interface {
	Set(val float64)
	Increment()
	Decrement()
	Add(val float64)
	Subtract(val float64)
}

type GaugeJob struct {
	// Job is interface of available operations in GaugeJob
	Job interface{
		Set(value float64)
		Inc()
		Dec()
		Add(value float64)
		Sub(value float64)
	}
}

func (job *GaugeJob) Increment() { job.Job.Inc() }
func (job *GaugeJob) Decrement() { job.Job.Dec() }
func (job *GaugeJob) Set(val float64) { job.Job.Set(val) }
func (job *GaugeJob) Add(val float64) { job.Job.Add(val) }
func (job *GaugeJob) Subtract(val float64) { job.Job.Sub(val) }

type CounterJob struct {
	// Job is interface of available operations in CounterJob
	Job interface{
		Inc()
		Add(value float64)
	}
}
func (job *CounterJob) Set(val float64) { log.Fatal("no set function from Counter metric") }
func (job *CounterJob) Increment() { job.Job.Inc() }
func (job *CounterJob) Decrement() { log.Fatal("no dec function from Counter metrics") }
func (job *CounterJob) Add(val float64) { job.Job.Add(val) }
func (job *CounterJob) Subtract(val float64) { log.Fatal("no substitute function from Counter metrics") }
