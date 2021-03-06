// This file allows the geo server to report its status to a central admin service
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var urlFormatWithPort = "http://%s:%s/ping"
var urlFormat = "http://%s/ping"

type heartbit struct {
	provider   CloudProvider
	remoteIP   string
	remotePort string
	interval   int
	client     http.Client
	z          *zone
}

func NewHeartBit(remoteIP, remotePort, interval string, z *zone) *heartbit {

	intervalNum := 10

	if intNum, err := strconv.Atoi(interval); err == nil {
		intervalNum = intNum
	}

	return &heartbit{
		remoteIP:   remoteIP,
		remotePort: remotePort,
		interval:   intervalNum,
		client:     http.Client{},
		z:          z,
	}

}

func (h *heartbit) Start() {

	if h.remoteIP == "" {
		log.Println("REMOTE_IP env variable is not set. Not sending heartbits.")
		return
	}

	// build the url
	var url string
	if h.remotePort == "" {
		url = fmt.Sprintf(urlFormat, h.remoteIP)
	} else {
		url = fmt.Sprintf(urlFormatWithPort, h.remoteIP, h.remotePort)
	}

	// start the ticker
	ticker := time.NewTicker(time.Duration(h.interval) * time.Second)

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				h.ping(url)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (h *heartbit) ping(url string) {
	// serve back the VM info and the client remote IP info

	if data := h.z.toJson(); len(data) > 0 {

		// set the timestamp
		h.z.Timestamp = time.Now().UTC()

		req, err := http.NewRequest("POST", url, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			log.Println("Response Status:", url, resp.Status)
		} else {
			log.Println("Could not send heartbit", err)
		}
		return
	}
	log.Println("Error serialize VM information to JSON")

}
