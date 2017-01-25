package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

// Define URLs for gathering cloud provider metadata
const (
	// Google compute engine
	kGCEExternalIP = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"
	kGCEZoneUrl    = "http://metadata.google.internal/computeMetadata/v1/instance/zone"

	// AWS
	kAWSExternalIP = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"
	kAWSZoneUrl    = "http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/forwarded-ips"

	// USING Public geoip API service
	geoServiceInternalIPUrl = "http://ipinfo.io/geo"
	geoServiceExternalIPUrl = "http://ipinfo.io/%s/geo"
)

// define http client to re-use
var zoneHttpClient http.Client

// defines struct to return to the client
// Contains
// - the virtual machine public IP address
// - the client IP address hitting this virtual machine
// - the virtual machine zone info
// https://cloud.google.com/compute/docs/regions-zones/regions-zones
type zone struct {
	CloudProvider   CloudProvider
	Name            string
	IpAddress       *coordinates
	ClientIpAddress *coordinates
}

// coordinates information of the ip address
type coordinates struct {
	Ip      string `json:"ip"`
	LatLong string `json:"loc"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}

// main methid which retrieves additional info
func getVMData(provider CloudProvider, clientIp string) *zone {
	z := zone{
		CloudProvider: provider,
	}

	getVMInfo(provider, &z)

	if clientIp != "" {
		// split /host /port
		host, _, _ := net.SplitHostPort(clientIp)
		coord := getIPCoordinates(host)
		z.ClientIpAddress = &coord
		log.Printf("Got client IP coordinates: %s\n", z.ClientIpAddress)
	}

	return &z
}

func (z *zone) toJson() []byte {
	var data []byte
	var err error

	data, err = json.Marshal(z)
	if err != nil {
		log.Println("Could not serialize zone to JSON", err)
	}

	return data
}

func ZoneFromJson(data []byte) (*zone, error) {
	var z *zone
	err := json.Unmarshal(data, z)
	return z, err

}

// helper which fills in all the VM info into the zone struct
func getVMInfo(provider CloudProvider, z *zone) {

	var zoneUrl string
	var publicIPUrl string

	switch provider {
	case GCE:
		zoneUrl = kGCEZoneUrl
		publicIPUrl = kGCEExternalIP
		break
	case AWS:
		zoneUrl = ""
		publicIPUrl = ""
		break
	}

	// gathers the external ip assigned to the VM
	ip := getExternalIP(provider, publicIPUrl)
	log.Println("Got external VM ip address", ip)
	if ip != "" {
		// Resolves the geo location info
		coord := getIPCoordinates(ip)
		z.IpAddress = &coord
		log.Printf("Got external VM ip address GEO: %s\n", z.IpAddress)
	}

	// get the zone information
	z.Name = getZoneInfo(provider, zoneUrl)

}

// Gets the VM external IP address assigned to it
func getExternalIP(provider CloudProvider, url string) string {
	// define vars
	var ip string
	var bodyBytes []byte

	// cerate new request
	req, _ := http.NewRequest("GET", url, nil)

	// add specific header
	switch provider {
	case GCE:
		req.Header.Set("Metadata-Flavor", "Google")
		break
	case AWS:
		break
	}

	// make the call
	resp, err := zoneHttpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			bodyBytes, err = ioutil.ReadAll(resp.Body)
			if err == nil {
				return string(bodyBytes)
			}
		} else {
			log.Println("VM External IP: got", resp.StatusCode, "status code.")
		}

	}

	log.Println("Failed to get external IP", url, err)
	return ip

}

// Calls the geo ip web service to resolve the geolocation info
// http://ipinfo.io/developers/jsonp-requests
func getIPCoordinates(ip string) coordinates {

	var coord coordinates

	// format the URL
	ipAddressUrl := fmt.Sprintf(geoServiceExternalIPUrl, ip)
	log.Println("Calling GEO IP Services", ipAddressUrl)

	// create request
	req, err := http.NewRequest("GET", ipAddressUrl, nil)
	if err != nil {
		log.Println("Error building request for getting IP coordinates", ipAddressUrl, err)
		return coord
	}

	// add specific CURL user agent header necessary for the api
	req.Header.Add("User-Agent", "curl/7.49.1")

	// make the request
	resp, err := zoneHttpClient.Do(req)
	if err != nil {
		// not good...
		log.Println("Unable to retrieve external GEO IP information:", ipAddressUrl, err)
	} else {
		// Parse the response
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			// bodyBytes, err := ioutil.ReadAll(resp.Body)
			// if err == nil {
			// 	log.Println("GEO string", string(bodyBytes))
			// }
			enc := json.NewDecoder(resp.Body)
			if err := enc.Decode(&coord); err != nil {
				log.Println("Failed to parse external GEO ip json response", ipAddressUrl, err)
				return coord
			}
		} else {
			log.Println("IP Coordinates service: got", resp.StatusCode, "status code")
		}
		log.Printf("Geo IP response: %s\n", coord)
		return coord
	}

	return coord
}

// Gets the provider specific zone information
func getZoneInfo(provider CloudProvider, url string) string {
	// define vars
	var zoneInfo string
	var bodyBytes []byte

	// create a new request
	req, _ := http.NewRequest("GET", url, nil)

	// add specific headers if necessary
	switch provider {
	case GCE:
		req.Header.Set("Metadata-Flavor", "Google")
		break
	case AWS:
		break
	}

	// makje the http call
	resp, err := zoneHttpClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			return string(bodyBytes)
		}
	}

	log.Println("Failed to get VM zone information", url)
	return zoneInfo
}
