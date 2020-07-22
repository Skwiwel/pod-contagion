// The starter acts the same as a generic Podder but starts an InfectionFrenzy by itself
package main

import (
	"flag"
	"time"

	"github.com/skwiwel/pod-contagion/app/podder"
)

var (
	httpAddr   = flag.String("http", "0.0.0.0:80", "HTTP service address.")
	healthAddr = flag.String("health", "0.0.0.0:81", "Health service address.")
)

func main() {
	flag.Parse()

	p := podder.MakePodder(*httpAddr, *healthAddr)
	go func() {
		p.Run()
	}()

	time.Sleep(10 * time.Second)
	go p.InfectionFrenzy()

	// Let the frenzy run for some time
	time.Sleep(2 * time.Second)
	// Shutdown
}
