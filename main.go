package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Response struct {
	Success      string `json:"success"`
	ListeArrivee string `json:"listeArrivee"`
}

type Arrivals []struct {
	Line            string      `json:"line"`
	LineName        interface{} `json:"lineName"`
	TextColor       string      `json:"textColor"`
	BackgroundColor string      `json:"backgroundColor"`
	Type            string      `json:"type"`
	Mode            string      `json:"mode"`
	Destination     string      `json:"destination"`
	Horaire         string      `json:"horaire"`
	EstApresMinuit  bool        `json:"estApresMinuit"`
	Disruption      struct {
		Category  string `json:"category"`
		Criticity string `json:"criticity"`
	} `json:"disruption"`
	Experimentation interface{} `json:"experimentation"`
	RealTime        bool        `json:"realTime"`
}
type StopsMap map[string]StopJSON

type StopJSON struct {
	Nom     Nom     `json:"nom"`
	CodeSMS CodeSMS `json:"codeSMS"`
}

type CodeSMS struct {
	Value int64 `json:"value"`
}

type Nom struct {
	Value string `json:"value"`
}

type Stop struct {
	Name string
	Code int64
}

func GetStops() []Stop {
	request, err := http.NewRequest("GET", "https://www.cts-strasbourg.eu/system/modules/eu.cts.module.core/actions/getStops.jsp", nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)
	var data StopsMap
	err = json.NewDecoder(res.Body).Decode(&data)
	var stops []Stop
	for _, stop := range data {
		stops = append(stops, Stop{stop.Nom.Value, stop.CodeSMS.Value})
	}
	return stops
}

func filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return ret
}

func MakeRequest(stops []Stop, args []string) Arrivals {
	poincare := func(s Stop) bool { return strings.Contains(s.Name, args[0]) }
	foundStops := filter(stops, poincare)
	if len(foundStops) > 1 {
		panic("Found more than 1 Stop")
	} else if len(foundStops) == 0 {
		panic("Found no Stop")
	}
	stop := foundStops[0]
	t := time.Now()
	body := fmt.Sprintf("smscode=%v&hour=%v&minute=%v&nbHoraire=1&locale=fr", stop.Code, t.Hour(), t.Minute())
	request, err := http.NewRequest("POST", "https://www.cts-strasbourg.eu/system/modules/eu.cts.module.horairetempsreel/actions/action_recherchetempsreel.jsp", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)
	var response Response
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		panic(err)
	}
	var data Arrivals
	err = json.Unmarshal([]byte(response.ListeArrivee), &data)
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	stops := GetStops()
	argsWithoutProg := os.Args[1:]
	arrivals := MakeRequest(stops, argsWithoutProg)

	for _, stop := range arrivals {
		println(stop.Horaire, stop.Destination)
	}
}
