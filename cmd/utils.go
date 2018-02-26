package cmd

import (
	"log"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("TRACE '%s' took %s", name, elapsed)
}
