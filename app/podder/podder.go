package podder

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/skwiwel/pod-contagion/app/health"
)

// Podder represents the functionality of an image replica inside Kubernetes engine.
// It's purpose is to test the Kubernetes liveness and shutdown mechanisms.
type Podder interface {
	Run()
	InfectionFrenzy()
}

type podder struct {
	httpAddr, healthAddr string
	health               health.Manager
	// stopServerChan will close as a sign to close http servers
	stopServerChan chan struct{}
}

// MakePodder is the constructor for a default Podder
func MakePodder(httpAddr, healthAddr string) Podder {
	p := podder{
		httpAddr:       httpAddr,
		healthAddr:     healthAddr,
		health:         health.MakeHealthManager(),
		stopServerChan: make(chan struct{}),
	}
	return &p
}

// Run brings the Podder to live by initiating 2 http servers:
// one for the Kubernetes health probe handling
// the other for listening to other Podders
func (p *podder) Run() {
	errChan := make(chan error, 10)

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/liveness", p.health.LivenessHandler)
	healthMux.HandleFunc("/readiness", p.health.ReadinessHandler)
	healthServer := &http.Server{Addr: p.healthAddr, Handler: healthMux}

	go func() {
		errChan <- healthServer.ListenAndServe()
	}()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/face", p.faceHandler)
	httpServer := &http.Server{Addr: p.httpAddr, Handler: httpMux}

	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	for {
		select {
		case <-p.stopServerChan:
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

func (p *podder) faceHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Println("sniff")
			p.InfectionFrenzy()
		case "":
			fmt.Fprintf(w, "Do something!\n")
		default:
			fmt.Fprintf(w, "I don't understand what you're doing.\n")
		}

	default:
		fmt.Fprintf(w, "Stop bothering me, please.")
	}
}

// InfectionFrenzy makes the Podder respond negatively to Kubernetes
// liveness probes and begins sneezing at other Podders
func (p *podder) InfectionFrenzy() {
	if p.health.LivenessStatus() != http.StatusOK {
		return
	}
	p.health.SetReadinessStatus(http.StatusTeapot)
	p.health.SetLivenessStatus(http.StatusTeapot)
	// close the http server
	close(p.stopServerChan)
	// sneeze on some Podders
	for i := 0; i < 10; i++ {
		go p.sneeze()
	}
}

// Sneeze sneezes on a Podder.
// Since this executable has no way of knowing other Podders exist
// it will sneeze on the address of it's face.
// Kubernetes' load balancing service should ensure that whatever
// the number of Podders, they will get sneezed on an equal amount.
func (p *podder) sneeze() {
	formData := url.Values{
		"action": {"achoo"},
	}
	resp, err := http.PostForm(fmt.Sprintf("http://%s/face", p.httpAddr), formData)
	if err != nil {
		log.Printf("could not sneeze: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
