package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/skwiwel/pod-contagion/app/health"
)

var (
	httpAddr   = flag.String("http", "0.0.0.0:80", "HTTP service address.")
	healthAddr = flag.String("health", "0.0.0.0:81", "Health service address.")
)

// stopServerChan will close as a sign to close http servers
var stopServerChan = make(chan struct{})

func main() {
	flag.Parse()

	log.Println("Starting server...")
	log.Printf("Health service listening on %s", *healthAddr)
	log.Printf("HTTP service listening on %s", *httpAddr)

	errChan := make(chan error, 10)

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/liveness", health.LivenessHandler)
	healthMux.HandleFunc("/readiness", health.ReadinessHandler)
	//http.HandleFunc("/health/status", HealthzStatusHandler)
	//http.HandleFunc("/readiness/status", ReadinessStatusHandler)
	healthServer := &http.Server{Addr: *healthAddr, Handler: healthMux}

	go func() {
		errChan <- healthServer.ListenAndServe()
	}()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/face", faceHandler)
	httpServer := &http.Server{Addr: *httpAddr, Handler: httpMux}

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	for {
		select {
		case <-stopServerChan:
			if err := httpServer.Close(); err != nil {
				log.Printf("error on closing: %v\n", err)
			}
		case err := <-errChan:
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func faceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		switch action := r.FormValue("action"); action {
		case "achoo":
			// This Podder is infected now
			fmt.Fprintf(w, "eww\n")
			health.SetLivenessStatus(http.StatusTeapot)
			log.Println("sniff")
			// close the http server
			close(stopServerChan)
			// sneeze on some Podders
			for i := 0; i < 10; i++ {
				go sneeze()
			}
		case "":
			fmt.Fprintf(w, "Do something!\n")
		default:
			fmt.Fprintf(w, "I don't understand what you're doing.\n")
		}

	default:
		fmt.Fprintf(w, "Stop bothering me, please.")
	}
}

// sneeze sneezes on a Podder.
// Since this executable has no way of knowing other Podders exist
// it will sneeze on the address of it's face.
// Kubernetes' load balancing service should ensure that whatever
// the number of Podders, they will get sneezed on an equal amount.
func sneeze() {
	formData := url.Values{
		"action": {"achoo"},
	}
	resp, err := http.PostForm(fmt.Sprintf("http://%s/face", *httpAddr), formData)
	if err != nil {
		log.Printf("could not sneeze: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
