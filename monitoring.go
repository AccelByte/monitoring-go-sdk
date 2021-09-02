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
	"log"
	"net/http"
	"time"
)

// supported metric & monitoring application
const (
	// Prometheus https://prometheus.io/
	Prometheus = iota
)

// Metric Type
const (
	// Counter is a cumulative metric, can only increase or be reset to zero on restart
	Counter = iota
	// Gauge is a single numerical value that can arbitrarily go up and down
	Gauge
)

// List all available operation
// Some operation may not available for some metric
const (
	Set = iota
	Add
	Sub
	Increment
	Decrement
)

// Metrics hold all the metrics
type Metrics struct {
	metrics []Metric
}

// Metric is the metric we want to serve in the service
type Metric struct {
	Name string
	Desc string

	// MetricType or metric data type
	MetricType int

	// OperationType is a type of operation that will be used for the specific metric
	OperationType int

	// UpdateInterval is interval to get the data from getter and then update the metric value
	UpdateInterval time.Duration

	// Getter is a function that return value for the metric
	Getter func() float64
}

type Client interface {
	Init(labels []Metric)
	GetHandler() http.Handler
}

func New(app int) Client {
	switch app {
	case Prometheus:
		return newPrometheus()
	default:
		log.Fatal("unknown metric instrument")
	}
	return nil
}
