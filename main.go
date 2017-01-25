package main

import (
	"log"
	"os"
	"strings"
	//"time"

	"github.com/boltdb/bolt"
)

func main() {

	// default the provider to GCE
	var providerId = GCE

	// read the ENV vars
	// provider
	provider := strings.ToLower(os.Getenv("PROVIDER"))

	// HTTP port
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")

	// check in case the provider is not GCE
	switch provider {
	case "aws":
		providerId = AWS
	}

	// if the port is emoty then set it to 8080
	if port == "" {
		port = "8080"
	}

	if host == "" {
		host = "0.0.0.0"
	}

	// ==============================================================
	// get VM info
	z := getVMData(providerId, "")
	// ==============================================================

	remoteIp := os.Getenv("REMOTE_IP")
	remotePort := os.Getenv("REMOTE_PORT")
	interval := os.Getenv("INTERVAL")

	// ==============================================================

	isAdminConfig := os.Getenv("ADMIN")
	isAdmin := false
	var adm *adminPanel
	if isAdminConfig == "1" {
		isAdmin = true
		// Open the my.db data file in your current directory.
		// It will be created if it doesn't exist.
		db, err := bolt.Open("my.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		adm = NewAdminPanel(db)

	} else {
		// Heartbit
		h := NewHeartBit(remoteIp, remotePort, interval, z)
		h.Start()
	}

	// ==============================================================

	// New API
	api := NewAPI(host, port, providerId, isAdmin, adm, z)
	// Start the API
	api.Start()

	// go reportLocation()
	// ticker := time.NewTicker(10 * time.Second)
	// quit := make(chan struct{})
	// go func() {
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			go reportLocation()
	// 		case <-quit:
	// 			ticker.Stop()
	// 			return
	// 		}
	// 	}
	// }()
	// http.HandleFunc("/", sayhelloName)           // set router
	// log.Fatal(http.ListenAndServe(":9090", nil)) // set listen port
}
