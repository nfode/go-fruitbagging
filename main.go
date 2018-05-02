package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"time"
)

var url = "https://fruitbagger.herokuapp.com/"

var apiKey = ""
var session = ""
var bag = ""
var client *http.Client

func CreateSession() {
	req, err := http.NewRequest("POST", url+"api/session", nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("auth", apiKey)
	resp, err := client.Do(req)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	session = bodyString
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func CreateBag() {
	req, err := http.NewRequest("POST", url+"api/bag/"+session, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("auth", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Create bag" + err.Error())
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	bag = bodyString
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

}

var fruits []Fruit

type FruitResponse map[string]int

type Fruit struct {
	key    string
	weight int
}

func GetLookahead() bool {
	req, err := http.NewRequest("GET", url+"api/fruits/"+session, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("auth", apiKey)
	for {
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("GetLookahead:" + err.Error())
		}
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if resp == nil {
			return false
		}
		if resp.StatusCode == http.StatusBadRequest {
			return false
		}
		if resp.StatusCode == http.StatusNoContent {
			return true
		}
		var res FruitResponse
		err = json.Unmarshal(bodyBytes, &res)
		if err != nil {
			fmt.Println(err)
		}
		for i, k := range res {
			fruit := Fruit{key: i, weight: k}
			fruits = append(fruits, fruit)
		}

		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}
func FindBestValue(weight int) (string, int) {
	sort.Slice(fruits, func(i, j int) bool {
		return fruits[i].weight > fruits[j].weight
	})
	index := -1
	best := ""
	last := math.MaxInt64
	for i, fruit := range fruits {
		newWeight := fruit.weight + weight
		if newWeight < 1000 {
			fruits = deleteFromArray(fruits, i)
			return fruit.key, newWeight
		}
		if newWeight > 1000 && newWeight < last {
			index = i
			best = fruit.key
			last = newWeight
		}
	}
	fruits = deleteFromArray(fruits, index)
	return best, last
}

func deleteFromArray(arr []Fruit, key int) []Fruit {
	return append(arr[:key], arr[key+1:]...)
}

func AddToBag(key string) {
	req, err := http.NewRequest("POST", url+"api/bagging/"+session+"/"+bag+"/"+key, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("auth", apiKey)
	resp, err := client.Do(req)
	if resp == nil {
		return
	}
	if err != nil {
		fmt.Print("add to bag" + err.Error())
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func CloseBag() {
	req, err := http.NewRequest("PUT", url+"api/bag/"+session+"/"+bag, nil)
	if err != nil {
		fmt.Print("close bag:" + err.Error())
	}
	req.Header.Add("auth", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print("close bag:" + err.Error())
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

func CloseSession() {
	req, err := http.NewRequest("PUT", url+"api/session/"+session, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("auth", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print("close session" + err.Error())
	}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

}
func main() {
	url = os.Getenv("FRUITBAGGER_URL")
	apiKey = os.Getenv("FRUITBAGGER_API_KEY")
	timeout := time.Duration(10 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}
	CreateSession()
	last := false
	for !last {
		CreateBag()
		var weight = 0
		for weight < 1000 {
			last = GetLookahead()
			newKey, newWeight := FindBestValue(weight)
			AddToBag(newKey)
			weight = newWeight
		}
		CloseBag()
	}
	CloseSession()
}
