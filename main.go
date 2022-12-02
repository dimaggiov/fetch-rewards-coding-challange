package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json: price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type ID struct {
	ID string `json:"id"`
}

type Points struct {
	Points int `json:"points"`
}

var receiptsMap = make(map[string]Receipt)

func processReceipts(w http.ResponseWriter, r *http.Request) {

	//generate UUID for key
	key := uuid.New().String()

	decoder := json.NewDecoder(r.Body)
	var readReceipt Receipt
	decodeErr := decoder.Decode(&readReceipt)
	invalid := false
	if decodeErr != nil {
		invalid = true
	}

	//validate input
	if (readReceipt.Retailer == "") ||
		(readReceipt.PurchaseDate == "") ||
		(readReceipt.PurchaseTime == "") ||
		(readReceipt.Total == "") {
		invalid = true
	}

	for i := 0; i < len(readReceipt.Items); i++ {
		if (readReceipt.Items[i].ShortDescription == "") ||
			(readReceipt.Items[i].Price == "") {
			invalid = true
		}
	}

	if invalid {
		//give code 400 if error occured
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["description"] = "The receipt is invalid"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	receiptsMap[key] = readReceipt

	idToReturn := ID{
		ID: key,
	}

	json.NewEncoder(w).Encode(idToReturn)
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	receipt, receiptExists := receiptsMap[id]

	//give error 404 if id is not in system
	if !receiptExists {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["description"] = "No receipt found for that id"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
		return
	}

	points := calculatePoints(receipt)

	pointsToReturn := Points{
		Points: points,
	}

	json.NewEncoder(w).Encode(pointsToReturn)
}

func calculatePoints(receipt Receipt) int {
	points := 0

	//alphaNumeric check
	points += countAlphaNumeric(receipt.Retailer)

	//Round dollar amount and multiple of 25 check
	totalStrLen := len(receipt.Total)
	cents, _ := strconv.Atoi(receipt.Total[totalStrLen-2:])

	if cents == 0 {
		points += 50
	}

	if cents%25 == 0 {
		points += 25
	}

	//check for pairs of items
	points += (len(receipt.Items) / 2) * 5

	//check item descriptions
	var items []Item = receipt.Items

	for i := 0; i < len(items); i++ {
		if len(items[i].ShortDescription)%3 == 0 {
			priceAsFloat, _ := strconv.ParseFloat(items[i].Price, 8)
			priceAsFloat *= 0.2
			points += int(math.Ceil(priceAsFloat))

		}
	}

	//check purchase date
	dateStrLen := len(receipt.PurchaseDate)
	day, _ := strconv.Atoi(receipt.PurchaseDate[dateStrLen-2:])

	if day%2 == 1 {
		points += 6
	}

	//check time
	hour, _ := strconv.Atoi(receipt.PurchaseTime[0:2])

	if (hour >= 14) && (hour < 16) {
		points += 10
	}

	return points
}

func countAlphaNumeric(str string) int {
	count := 0

	for i := 0; i < len(str); i++ {
		nextChar := str[i]
		if ('a' <= nextChar && nextChar <= 'z') ||
			('A' <= nextChar && nextChar <= 'Z') ||
			('0' <= nextChar && nextChar <= '9') {
			count++
		}
	}

	return count
}

func handleRequests() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/receipts/process", processReceipts).Methods("POST")
	router.HandleFunc("/receipts/{id}/points", getPoints).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", router))
}

func main() {
	handleRequests()
}
