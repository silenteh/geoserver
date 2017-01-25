package main

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"
// )

// var myIp string

// const reportingApiUrl string = "https://z4a24pdwp6.execute-api.eu-west-1.amazonaws.com/prod/AccessGeoLocations"
// const ipServiceUrl string = "https://api.ipify.org"

// //const geoServiceUrl string = "http://ip-api.com/json/"
// const geoServiceUrl string = "http://ipinfo.io/geo"

// func sendLocation(jsonStr string) {
// 	var jsonByte = []byte(jsonStr)
// 	req, err := http.NewRequest("POST", reportingApiUrl, bytes.NewBuffer(jsonByte))
// 	req.Header.Set("Content-Type", "application/json")
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()
// 	fmt.Println("response Status:", resp.Status)
// }

// func getMyIPAddress() string {
// 	response, err := http.Get(ipServiceUrl)
// 	if err != nil {
// 		log.Println("Unable to retrieve IP: ", err)
// 	} else {
// 		defer response.Body.Close()
// 		buf := new(bytes.Buffer)
// 		buf.ReadFrom(response.Body)
// 		myIp := buf.String()
// 		return myIp
// 	}
// }

// func getCoordinates() map[string]string {

// 	coordinates := make(map[string]string)

// 	//response, err := http.Get(geoServiceUrl)
// 	client := &http.Client{}
// 	req, err := http.NewRequest("GET", geoServiceUrl, nil)
// 	// ...
// 	req.Header.Add("User-Agent", "curl/7.49.1")
// 	response, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Unable to retrieve coordinates:", err)
// 		return coordinates
// 	} else {
// 		defer response.Body.Close()
// 		buf := new(bytes.Buffer)
// 		buf.ReadFrom(response.Body)
// 		var dat map[string]interface{}
// 		//fmt.Println("response data:", buf.String())
// 		if err := json.Unmarshal(buf.Bytes(), &dat); err != nil {
// 			panic(err)
// 		}
// 		res := strings.Split(dat["loc"].(string), ",")
// 		if len(res) > 1 {
// 			coordinates["lat"] = res[0]
// 			coordinates["lon"] = res[1]
// 			coordinates["ip"] = dat["ip"].(string)
// 			//return map[string]string{"lat": lat, "lon": lon, "ip": dat["ip"].(string)}
// 			return coordinates
// 		}
// 	}

// 	return coordinates
// }

// func reportLocation() {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			fmt.Println("Recovered in f", r)
// 		}
// 	}()
// 	coord := getCoordinates()
// 	hostname, _ := os.Hostname()
// 	timestamp := int64(time.Now().Unix())
// 	mapD := map[string]interface{}{
// 		"IP":       coord["ip"],
// 		"hostname": hostname,
// 		"date":     timestamp,
// 		"lat":      coord["lat"],
// 		"lng":      coord["lon"],
// 	}
// 	mapB, _ := json.Marshal(mapD)
// 	fmt.Println("JSON:", string(mapB))
// 	sendLocation(string(mapB))
// }

// func sayhelloName(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()       // parse arguments, you have to call this by yourself
// 	fmt.Println(r.Form) // print form information in server side
// 	fmt.Println("path", r.URL.Path)
// 	fmt.Println("scheme", r.URL.Scheme)
// 	fmt.Println(r.Form["url_long"])
// 	for k, v := range r.Form {
// 		fmt.Println("key:", k)
// 		fmt.Println("val:", strings.Join(v, ""))
// 	}
// 	fmt.Fprintf(w, getZone()) // send data to client side
// }
