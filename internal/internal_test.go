package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildFQName(t *testing.T) {
	scenarios := []struct{ namespace, name, result string }{
		{"a", "b", "a_b"},
		{"", "b", "b"},
		{"a", "", ""},
		{"a", "", ""},
		{"", "", ""},
		{" ", "", ""},
	}

	for i, s := range scenarios {
		if want, got := s.result, BuildFQName(s.namespace, s.name); want != got {
			t.Errorf("%d. want %s, got %s", i, want, got)
		}
	}
}

func TestRandomSleep(t *testing.T) {
	// test that zero max returns immediately
	s := time.Now()
	RandomSleep(time.Duration(0), make(chan struct{}))
	elapsed := time.Since(s)
	assert.True(t, elapsed < time.Millisecond)

	// test that max sleep is respected
	s = time.Now()
	RandomSleep(time.Millisecond*50, make(chan struct{}))
	elapsed = time.Since(s)
	assert.True(t, elapsed < time.Millisecond*100)

	// test that shutdown is respected
	s = time.Now()
	shutdown := make(chan struct{})
	go func() {
		time.Sleep(time.Millisecond * 100)
		close(shutdown)
	}()
	RandomSleep(time.Second, shutdown)
	elapsed = time.Since(s)
	assert.True(t, elapsed < time.Millisecond*150)
}
