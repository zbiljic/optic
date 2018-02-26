package internal

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"
)

// BuildFQName joins the given two name components by "_". Empty name components
// are ignored. If the name parameter itself is empty, an empty string is
// returned, no matter what.
func BuildFQName(namespace, name string) string {
	if name == "" {
		return ""
	}
	switch {
	case namespace != "":
		return strings.Join([]string{namespace, name}, "_")
	}
	return name
}

// RandomSleep will sleep for a random amount of time up to max.
// If the shutdown channel is closed, it will return before it has finished
// sleeping.
func RandomSleep(max time.Duration, shutdown chan struct{}) {
	if max == 0 {
		return
	}
	maxSleep := big.NewInt(max.Nanoseconds())

	var sleepns int64
	if j, err := rand.Int(rand.Reader, maxSleep); err == nil {
		sleepns = j.Int64()
	}

	t := time.NewTimer(time.Nanosecond * time.Duration(sleepns))
	select {
	case <-t.C:
		return
	case <-shutdown:
		t.Stop()
		return
	}
}
