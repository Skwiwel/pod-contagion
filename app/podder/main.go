package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/skwiwel/pod-contagion/health"
)

func main() {
	var (
		httpAddr   = flag.String("http", "0.0.0.0:80", "HTTP service address.")
		healthAddr = flag.String("health", "0.0.0.0:81", "Health service address.")
	)
	flag.Parse()

	log.Println("Starting server...")
	log.Printf("Health service listening on %s", *healthAddr)
	log.Printf("HTTP service listening on %s", *httpAddr)

	errChan := make(chan error, 10)

	healthServer := http.NewServeMux()
	healthServer.HandleFunc("/liveness", health.LivenessHandler)
	healthServer.HandleFunc("/readiness", health.ReadinessHandler)
	//http.HandleFunc("/health/status", HealthzStatusHandler)
	//http.HandleFunc("/readiness/status", ReadinessStatusHandler)

	go func() {
		errChan <- http.ListenAndServe(*healthAddr, healthServer)
	}()

	httpServer := http.NewServeMux()
	httpServer.HandleFunc("/face", faceHandler)

	go func() {
		errChan <- http.ListenAndServe(*httpAddr, httpServer)
	}()

	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Fatal(err)
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
		case "":
			fmt.Fprintf(w, "Do something!\n")
		default:
			fmt.Fprintf(w, "I don't understand what you're doing.\n")
		}

	default:
		fmt.Fprintf(w, "Stop bothering me, please.")
	}
}
