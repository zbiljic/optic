package testutil

import (
	"net"
	"net/url"
	"os"
	"time"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/optic/logline"
	"github.com/zbiljic/optic/optic/metric"
	"github.com/zbiljic/optic/optic/raw"
)

var localhost = "localhost"

// GetLocalHost returns the DOCKER_HOST environment variable, parsing out any
// scheme or ports so that only the IP address is returned.
func GetLocalHost() string {
	if dockerHostVar := os.Getenv("DOCKER_HOST"); dockerHostVar != "" {
		u, err := url.Parse(dockerHostVar)
		if err != nil {
			return dockerHostVar
		}

		// split out the ip addr from the port
		host, _, err := net.SplitHostPort(u.Host)
		if err != nil {
			return dockerHostVar
		}

		return host
	}
	return localhost
}

// TestRaw Returns a simple test raw event:
//     source -> "test1"
//     value -> value
//     tags -> "tag1":"value1"
//     fields -> nil
func TestRaw(value []byte) optic.Raw {
	source := "test1"
	tags := map[string]string{"tag1": "value1"}
	r, _ := raw.New(
		source,
		value,
		tags,
		nil,
	)
	return r
}

// TestMetric Returns a simple test metric event:
//     namespace -> "test1" or name
//     tags -> "tag1":"value1"
//     value -> value
//     time -> time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
func TestMetric(value interface{}, name ...string) optic.Metric {
	if value == nil {
		panic("Cannot use a nil value")
	}
	namespace := "test1"
	if len(name) > 0 {
		namespace = name[0]
	}
	tags := map[string]string{"tag1": "value1"}
	m, _ := metric.New(
		namespace,
		tags,
		map[string]interface{}{"value": value},
		time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	return m
}

// TestLogLine Returns a simple test logline event:
//     path -> "test1"
//     content -> content
//     tags -> "tag1":"value1"
//     fields -> nil
func TestLogLine(content string) optic.LogLine {
	path := "test1"
	tags := map[string]string{"tag1": "value1"}
	ll, _ := logline.New(
		path,
		content,
		tags,
		nil,
	)
	return ll
}

// MockRaw returns a mock `[]optic.Raw`` object for using in unit tests.
func MockRaw() []optic.Raw {
	rawEvents := make([]optic.Raw, 0)
	rawEvents = append(rawEvents, TestRaw([]byte("raw")))
	return rawEvents
}

// MockMetrics returns a mock `[]optic.Metric`` object for using in unit tests.
func MockMetrics() []optic.Metric {
	metricsEvents := make([]optic.Metric, 0)
	metricsEvents = append(metricsEvents, TestMetric(1.0))
	return metricsEvents
}

// MockLogLines returns a mock `[]optic.LogLine`` object for using in unit tests.
func MockLogLines() []optic.LogLine {
	loglineEvents := make([]optic.LogLine, 0)
	loglineEvents = append(loglineEvents, TestLogLine("logline"))
	return loglineEvents
}

// MockEvents returns a mock `[]optic.Event`` object for using in unit tests.
func MockEvents() []optic.Event {
	events := make([]optic.Event, 0)
	events = append(events, TestRaw([]byte("raw")))
	events = append(events, TestMetric(1.0))
	events = append(events, TestLogLine("logline"))
	return events
}
