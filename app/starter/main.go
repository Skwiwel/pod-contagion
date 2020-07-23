// The starter acts the same as a generic Podder but starts an InfectionFrenzy by itself
package main

import (
	"flag"
	"time"

	"github.com/skwiwel/pod-contagion/app/podder"
)

var (
	httpAddr     = flag.String("http", "0.0.0.0:80", "HTTP service address.")
	healthAddr   = flag.String("health", "0.0.0.0:81", "Health service address.")
	symptomDelay = flag.Int("symptomDelay", 5000, "The time delay between getting infected "+
		"and starting showing symptoms (infection spreading and kubernetes health status change) in milliseconds.")
	healthDelay = flag.Int("healthDelay", 500, "The time delay between the end of symptomDelay "+
		"and signaling negative health status to Kubernetes probes  in milliseconds.")
	sneezeInterval = flag.Int("sneezeInterval", 500, "The time interval between sneezes in milliseconds. "+
		"The amount of sneezes will depend on how fast the container is killed.")
)

func main() {
	flag.Parse()

	p := podder.MakePodder(*httpAddr, *healthAddr, *symptomDelay, *healthDelay, *sneezeInterval)
	go func() {
		p.Run()
	}()

	time.Sleep(10 * time.Second)
	go p.InfectionFrenzy()

	// Let the frenzy run for some time
	time.Sleep(2 * time.Second)
	// Shutdown
}
