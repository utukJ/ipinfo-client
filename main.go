package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ipinfo/go/v2/ipinfo"
)

type InfoResp struct {
	ClientIP string `json:"client_ip"`
	Location string `json:"location"`
	Greeting string `json:"greeting"`
}

type WeatherDesc struct {
	Main        string `json:"main"`
	Description string `json:"description"`
}

type WeatherDeets struct {
	Temp float32 `json:"temp"`
}

type OpenWeatherResp struct {
	Weather []WeatherDesc `json:"weather"`
	Main    WeatherDeets  `json:"main"`
	Base    string        `json:"base"`
	Dt      int           `json:"dt"`
}

func kevinToCelcius(k float32) float32 {
	return k - 273
}

func main() {
	fmt.Println("Running ipinfo-client!")
	/*
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading env file: ", err)
		}
	*/

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})

	http.HandleFunc("/getinfo", func(w http.ResponseWriter, r *http.Request) {

		clientIP := r.Header.Get("X-Forwarded-For")

		ipInfoToken := os.Getenv("IPINFO_TOKEN")
		client := ipinfo.NewClient(nil, nil, ipInfoToken)
		ipinfo, err := client.GetIPInfo(net.ParseIP(clientIP))
		if err != nil {
			log.Fatalln(err)
		}

		visitorName := r.URL.Query().Get("visitor_name")
		if visitorName == "" {
			visitorName = "Nameless User"
		}

		openWeatherToken, ok := os.LookupEnv("OPENWEATHER_TOKEN")
		if !ok {
			log.Fatalln("could not find token")
		}
		openWeatherURL := fmt.Sprintf(
			"https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s",
			ipinfo.City,
			openWeatherToken)

		fmt.Println("Making request to: ", openWeatherURL)

		resp, err := http.Get(openWeatherURL)
		if err != nil {
			log.Fatalln(err)
		}

		decodedResp := OpenWeatherResp{}
		json.NewDecoder(resp.Body).Decode(&decodedResp)

		info := InfoResp{
			ClientIP: r.RemoteAddr,
			Location: ipinfo.City,
			Greeting: fmt.Sprintf(
				"Hi %s. The weather is %.2f degrees celcius in %s with %s.",
				visitorName,
				kevinToCelcius(decodedResp.Main.Temp),
				ipinfo.City,
				decodedResp.Weather[0].Description,
			),
		}

		fmt.Println("Delivering getinfo response")
		json.NewEncoder(w).Encode(info)
	})

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalln(err)
	}

}
