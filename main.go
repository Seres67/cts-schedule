package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

type Response struct {
	Success  string `json:"success"`
	Arrivals string `json:"listeArrivee"`
}

type Arrivals []struct {
	Line            string      `json:"line"`
	LineName        interface{} `json:"lineName,omitempty"`
	TextColor       string      `json:"textColor"`
	BackgroundColor string      `json:"backgroundColor"`
	Type            string      `json:"type"`
	Mode            string      `json:"mode"`
	Destination     string      `json:"destination"`
	DepartureTime   string      `json:"horaire"`
	AfterMidnight   bool        `json:"estApresMinuit"`
	Disruption      struct {
		Category  string `json:"category"`
		Criticity string `json:"criticity"`
	} `json:"disruption"`
	Experimentation interface{} `json:"experimentation,omitempty"`
	RealTime        bool        `json:"realTime"`
}
type StopsMap map[string]StopJSON

type StopJSON struct {
	Name Nom     `json:"nom"`
	Code CodeSMS `json:"codeSMS"`
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
		stops = append(stops, Stop{stop.Name.Value, stop.Code.Value})
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

// FilterStops TODO: um this isn't the way now that it is an API
func FilterStops(stops []Stop, stopQuery string) Stop {
	stop := func(s Stop) bool { return strings.Contains(strings.ToLower(s.Name), strings.ToLower(stopQuery)) }
	foundStops := filter(stops, stop)
	index := 0
	if len(foundStops) > 1 {
		//for i, foundStop := range foundStops {
		//	fmt.Printf("%d %s\n", i+1, foundStop.Name)
		//}
		//print("Quel arrêt? ")
		//_, err := fmt.Scanf("%d", &index)
		//if err != nil {
		//	panic(err)
		//}
		//index = index - 1
	} else if len(foundStops) == 0 {
		panic("Found no Stop")
	}
	return stops[index]
}

func MakeRequest(stop Stop) Arrivals {
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
	err = json.Unmarshal([]byte(response.Arrivals), &data)
	if err != nil {
		panic(err)
	}
	return data
}

var stops []Stop

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "API is online!",
	})
}

func HandleStop(c *gin.Context) {
	name := c.Param("name")
	stop := FilterStops(stops, name)
	arrivals := MakeRequest(stop)
	c.JSON(http.StatusOK, gin.H{
		"arrivals": arrivals,
	})
}

func HandleStops(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"stops": stops,
	})
}

func main() {
	stops = GetStops()

	r := gin.Default()
	r.GET("/", Status)
	r.GET("/stop/:name", HandleStop)
	r.GET("/stops", HandleStops)

	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}

//TODO: replace every panic with better error handling (we don't want the API to crash every few seconds)
