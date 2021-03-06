package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"html/template"
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
	host          string
	isAdmin       bool
	admin         *adminPanel
	z             *zone
}

// Create a new API struct
func NewAPI(host, port string, provider CloudProvider, isAdmin bool, admin *adminPanel, z *zone) *api {
	return &api{
		cloudProvider: provider,
		port:          port,
		host:          host,
		isAdmin:       isAdmin,
		admin:         admin,
		z:             z,
	}
}

// Start the API
func (api *api) Start() {

	http.HandleFunc("/favicon.ico", api.faviconHandlerFunc)

	if api.isAdmin {
		// Add zone handler
		http.HandleFunc("/ping", api.pingHandlerFunc)

		http.HandleFunc("/", api.adminIndexHandlerFunc)

		http.HandleFunc("/services", api.adminHandlerFunc)

	} else {

		// Add zone handler
		http.HandleFunc("/", api.zoneIndexHandlerFunc)

		// Service to get the data
		http.HandleFunc("/location", api.zoneHandlerFunc)

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
	portHost := fmt.Sprintf("%s:%s", api.host, api.port)

	// show the vars
	if api.isAdmin {
		log.Println("Admin mode - HTTP listening on:", portHost, "for provider", api.cloudProvider)
	} else {
		log.Println("HTTP listening on:", portHost, "for provider", api.cloudProvider)
	}

	// set the zone as ready
	api.z.Ready = isReady

	// socket listening
	log.Fatal(http.ListenAndServe(portHost, nil))
}

func (api *api) adminIndexHandlerFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		log.Fatal("Error parsing admin template files ", err)
	}
	t.Execute(w, nil)
}

func (api *api) zoneIndexHandlerFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("Error parsing admin template files ", err)
	}
	t.Execute(w, nil)
}

func (api *api) adminHandlerFunc(w http.ResponseWriter, r *http.Request) {
	var vms []*zone
	if err := api.admin.getAll(&vms); err != nil {
		log.Println("Error getting VMs", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	finalList := vms
	now := time.Now().UTC().Add(-1 * time.Minute)
	for index, vm := range vms {
		if now.After(vm.Timestamp) {
			finalList = remove(vms, index)
		}
	}

	data, _ := json.Marshal(&finalList)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (api *api) pingHandlerFunc(w http.ResponseWriter, r *http.Request) {

	log.Println("Got heartbeat:", getIPAdress(r))

	if r.Body == nil {
		http.Error(w, "Missing request body", 400)
		return
	}

	//serve admin index
	decoder := json.NewDecoder(r.Body)

	var z zone
	err := decoder.Decode(&z)

	if err != nil {
		log.Println("Could not deserialize zone", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update in botldb
	if err := api.admin.ping(&z); err != nil {
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
	log.Println("Used KILL SWITCH")
	os.Exit(1)
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
	api.z.Ready = isReady
	mutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

// Disable the service
func (api *api) disableHandlerFunc(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	isReady = false
	api.z.Ready = isReady
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
	host, _, _ := net.SplitHostPort(getIPAdress(r))
	if host != "" {
		coord := getIPCoordinates(host)
		api.z.ClientIpAddress = &coord
		log.Printf("Got client IP coordinates: %s\n", api.z.ClientIpAddress)
	}

	// serve the data back
	if data := api.z.toJson(); len(data) > 0 {
		w.Write(data)
		return
	}
	log.Println("Error serialize VM information to JSON")

	// if z := getVMData(api.cloudProvider, getIPAdress(r)); z != nil {

	// }

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

func remove(slice []*zone, s int) []*zone {
	return append(slice[:s], slice[s+1:]...)
}
