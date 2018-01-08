package main

import (
	"math/rand"
	"time"

	optic "github.com/zbiljic/optic/cmd"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	optic.Main()
}
