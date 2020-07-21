// The generic is a Podder runtime
package main

import (
	"flag"

	"github.com/skwiwel/pod-contagion/app/podder"
)

var (
	httpAddr   = flag.String("http", "0.0.0.0:80", "HTTP service address.")
	healthAddr = flag.String("health", "0.0.0.0:81", "Health service address.")
)

func main() {
	flag.Parse()

	p := podder.MakePodder(*httpAddr, *healthAddr)
	p.Run()
}
