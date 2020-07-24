package podder

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

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
	serviceAddr          string
	health               health.Manager
	symptomDelay         time.Duration
	healthDelay          time.Duration
	sneezeInterval       time.Duration
	// stopServerChan will close as a sign to close http servers
	stopServerChan chan struct{}
	// these ensure only one infection frenzy is occuring
	infected bool
	iMux     sync.Mutex
}

// MakePodder is the constructor for a default Podder
// sneezeTimeInterval is time in milliseconds
func MakePodder(httpAddr, healthAddr string, symptomDelay, healthDelay, sneezeInterval int) Podder {
	// Try to obtain the kubernetes service address
	// It should be located in an environment variable
	var serviceAddr string
	tempAddr := os.Getenv("PODDER_SERVICE_HOST")
	tempPort := os.Getenv("PODDER_SERVICE_PORT")
	if tempAddr != "" && tempPort != "" {
		serviceAddr = fmt.Sprintf("%s:%s", tempAddr, tempPort)
	} else {
		// If there is no such env variable then just assume it's the http address.
		// The Podder will likely not be able to communicate with others, though.
		log.Println("could not acquire Kubernetes service address; assuming http address")
		serviceAddr = httpAddr
	}

	p := podder{
		httpAddr:       httpAddr,
		healthAddr:     healthAddr,
		serviceAddr:    serviceAddr,
		health:         health.MakeHealthManager(),
		symptomDelay:   time.Duration(symptomDelay) * time.Millisecond,
		healthDelay:    time.Duration(healthDelay) * time.Millisecond,
		sneezeInterval: time.Duration(sneezeInterval) * time.Millisecond,
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

	log.Printf("Podder operational.\n")

	for {
		select {
		case <-p.stopServerChan:
			ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
			defer cancel()
			if err := httpServer.Shutdown(ctx); err != nil {
				//log.Printf("error on http shutdown: %v\n", err)
			}
			p.stopServerChan = nil
		case err := <-errChan:
			if err != nil && err.Error() != "http: Server closed" {
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
			fmt.Fprintf(w, "could not parse http POST form: %v", err)
			return
		}

		switch action := r.FormValue("action"); action {
		case "achoo":
			// This Podder is infected now
			fmt.Fprintf(w, "eww\n")

			// check if already infected
			p.iMux.Lock()
			if p.infected {
				p.iMux.Unlock()
				return
			}
			p.infected = true
			p.iMux.Unlock()

			go func() {
				time.Sleep(p.symptomDelay)
				go p.InfectionFrenzy()
				time.Sleep(p.healthDelay)
				p.health.SetLivenessStatus(http.StatusTeapot)
				p.health.SetReadinessStatus(http.StatusTeapot)
			}()
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
	log.Println("sniff")
	// wait for the http server to reply
	if p.symptomDelay < 100*time.Millisecond {
		time.Sleep(100 * time.Millisecond)
	}
	// close the http server
	close(p.stopServerChan)
	// sneeze on some Podders
	for {
		go p.sneeze()
		time.Sleep(p.sneezeInterval)
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
	resp, err := http.PostForm(fmt.Sprintf("http://%s/face", p.serviceAddr), formData)
	if err != nil {
		// The log spam is atrocious, so better to disable this logging.
		//log.Printf("could not sneeze: %v\n", err)
		return
	}
	defer resp.Body.Close()
}
