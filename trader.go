package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
	"xabbo.b7c.io/goearth/shockwave/trade"
)

var lastTrade trade.Offers
var messageSent bool

func init() {
	tradeMgr.Completed(handleTradeComplete)
}

type TrackerData struct {
	Furni            string `json:"furni"`
	Note             string `json:"note"`
	Quantity         int    `json:"quantity"`
	Result           string `json:"result"`
	AddedByExtension bool   `json:"addedByExtension"`
}

type InventoryItem struct {
	ItemName  string `json:"itemName"`
	ItemType  string `json:"itemType"`
	ItemClass string `json:"itemClass"`
	ItemProps string `json:"itemProps"`
	Count     int    `json:"count"`
}

func sendTrackerData(offers trade.Offers) {
	itemCounts := make(map[string]map[string]int)

	// Log details of offers
	for _, offer := range offers {
		for _, item := range offer.Items {
			if itemCounts[item.Class] == nil {
				itemCounts[item.Class] = make(map[string]int)
			}
			itemCounts[item.Class][offer.Name]++
		}
	}

	for className, owners := range itemCounts {
		for owner, count := range owners {
			result := "won"
			if owner == profileMgr.Name {
				result = "lose"
			}

			currentTime := time.Now().Format("2006-01-02 15:04:05")

			note := ""

			if offers[1].Name == profileMgr.Name {
				note = "Traded with " + offers[0].Name + " at " + currentTime + " at room "
			} else {
				note = "Traded with " + offers[1].Name + " at " + currentTime + " at room "
			}

			data := TrackerData{
				Furni:            className,
				Note:             note,
				Quantity:         count,
				Result:           result,
				AddedByExtension: true,
			}

			dataBytes, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling JSON:", err)
				continue
			}

			req, err := http.NewRequest("POST", "https://legacyhabbo.me/tracker/add", bytes.NewBuffer(dataBytes))
			if err != nil {
				log.Println("Error creating POST request:", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-AUTH-TOKEN", authToken)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error sending POST request:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				ext.Send(in.SYSTEM_BROADCAST, []byte("There was a issue with the tracker, contact the developer."))
				log.Printf("Received non-OK response: %s\n", resp.Status)
			}
		}
	}
}
func handleTradeComplete(args trade.Args) {
	lastTrade = args.Offers

	if enabled {
		go sendTrackerData(lastTrade)

		if lastTrade.Tradee().Items == nil {
			ext.Send(out.TRADE_OPEN)
			return
		}

		if !messageSent {
			ext.Send(in.SYSTEM_BROADCAST, []byte("Tracker is enabled, so this trade was tracked, you can disable by typing :tracker."))
			messageSent = true
		}
	}

}
