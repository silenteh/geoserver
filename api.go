package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Defines the currently supported cloud providers
// TODO: AWS
const (
	AWS CloudProvider = iota + 1
	GCE
)

// Cloud Provider ENUM
type CloudProvider uint8

var (
	mutex   = sync.Mutex{}
	isReady = true
)

// ==============================

// API Service
type api struct {
	cloudProvider CloudProvider
	port          string
	isAdmin       bool
	admin         *adminPanel
}

// Create a new API struct
func NewAPI(port string, provider CloudProvider, isAdmin bool, admin *adminPanel) *api {
	return &api{
		cloudProvider: provider,
		port:          port,
		isAdmin:       isAdmin,
		admin:         admin,
	}
}

// Start the API
func (api *api) Start() {

	http.HandleFunc("/favicon.ico", api.faviconHandlerFunc)

	if api.isAdmin {
		// Add zone handler
		http.HandleFunc("/", api.adminHandlerFunc)

	} else {

		// Add zone handler
		http.HandleFunc("/", api.zoneHandlerFunc)

		// used for signaling that the conatiner is up and running
		http.HandleFunc("/live", api.liveHandlerFunc)

		// used for signaling that the conatiner is ready to receive requests
		http.HandleFunc("/ready", api.readyHandlerFunc)

		// disable the readyness
		http.HandleFunc("/disable", api.disableHandlerFunc)

		// enable the readyness
		http.HandleFunc("/enable", api.enableHandlerFunc)

		// kill ther app
		http.HandleFunc("/kill", api.killHandlerFunc)

	}

	// start the HTTP server
	portHost := fmt.Sprintf(":%s", api.port)

	// show the vars
	log.Println("HTTP listening on:", portHost, "for provider", api.cloudProvider)

	// socket listening
	log.Fatal(http.ListenAndServe(portHost, nil))
}

func (api *api) adminHandlerFunc(w http.ResponseWriter, r *http.Request) {
	//serve admin index
	decoder := json.NewDecoder(r.Body)

	var z *zone
	err := decoder.Decode(z)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update in botldb
	if err := api.admin.ping(z); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := z.toJson()
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

}

func (api *api) faviconHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (api *api) killHandlerFunc(w http.ResponseWriter, r *http.Request) {
	log.Panic("Used KILL SWITCH")
}

// checks whether the service is up and runnning
func (api *api) liveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// check if the service is ready to serve requests
func (api *api) readyHandlerFunc(w http.ResponseWriter, r *http.Request) {
	api.writeReadyness(w)
}

// Enable the service
func (api *api) enableHandlerFunc(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	isReady = true
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// Disable the service
func (api *api) disableHandlerFunc(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	isReady = false
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// Returns the datacenter zone information about the running process
func (api *api) zoneHandlerFunc(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	if !isReady {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// serve back the VM info and the client remote IP info
	if z := getVMData(api.cloudProvider, getIPAdress(r)); z != nil {
		if data := z.toJson(); len(data) > 0 {
			w.Write(data)
			return
		}
		log.Println("Error serialize VM information to JSON")
	}

	// otherwise return error
	w.WriteHeader(http.StatusServiceUnavailable)

}

func (api *api) writeReadyness(w http.ResponseWriter) {

	mutex.Lock()
	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	mutex.Unlock()

}
