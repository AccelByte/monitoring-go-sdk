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

package main

import (
	"github.com/AccelByte/monitoring-go-sdk"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type data struct {
	count        int64
	sum          float64
	currentValue chan float64
}

func newData() *data {
	return &data{count: 0, sum: 0, currentValue: make(chan float64)}
}

func main() {
	metricData := newData()
	monitoringClient := monitoring.New(monitoring.Prometheus)
	metrics := []monitoring.Metric{
		{
			Name:           "gauge_metric_test",
			Desc:           "this metric will always be changed on every 3 second",
			MetricType:     monitoring.Gauge,
			OperationType:  monitoring.Set,
			UpdateInterval: time.Second * 3,
			Getter:         metricData.getAverage,
		},
		{
			Name:           "counter_metric_test",
			Desc:           "this metric will always be incremented by 1 every 5 second",
			MetricType:     monitoring.Counter,
			OperationType:  monitoring.Increment,
			UpdateInterval: time.Second * 5,
		},
		{
			Name:           "counter_metric_test_add_by_2",
			Desc:           "this metric will always be incremented by 2 every 10 second",
			MetricType:     monitoring.Counter,
			OperationType:  monitoring.Add,
			UpdateInterval: time.Second * 10,
			Getter: func() float64{ return 3 },
		},
	}

	//
	monitoringClient.Init(metrics)

	go metricData.listenNewData()
	go metricData.updateAverage()

	// this will have /metrics endpoint that serve our 3 metrics
	// visit localhost:2112/metrics to view the metric data
	http.Handle("/metrics", monitoringClient.GetHandler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}

// this function will be called by prometheus job on every interval we define on metric list
func (metricData *data) getAverage() float64 {
	average := metricData.sum / float64(metricData.count)
	log.Printf("returning average: %v", average)
	metricData.sum = 0
	metricData.count = 0
	return average
}

// these 2 functions will update and listen on metric data we want to serve
// implementation can be vary based on how we populate the data

// this will update metricData on every new data updated
func (metricData *data) listenNewData() {
	for {
		select {
		case msg, _ := <-metricData.currentValue:
			metricData.sum += msg
			metricData.count++
			log.Printf("new data: %v, count: %v, sum: %v", msg, metricData.count, metricData.sum)

		}
	}
}

// this will update metricData with random data for every 1 second
func (metricData *data) updateAverage() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			rand.Seed(time.Now().UnixNano())
			min := 0
			max := 5
			random := float64(rand.Intn(max-min+1) + min)
			metricData.currentValue <- random
		}
	}
}
