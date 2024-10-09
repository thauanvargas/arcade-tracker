package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
	"xabbo.b7c.io/goearth/shockwave/profile"
	"xabbo.b7c.io/goearth/shockwave/room"
	"xabbo.b7c.io/goearth/shockwave/trade"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Arcade Tracker",
	Description: "TSA's Arcade Tracker",
	Version:     "1.1",
	Author:      "Thauan",
})

var profileMgr = profile.NewManager(ext)
var tradeMgr = trade.NewManager(ext)
var roomMgr = room.NewManager(ext)
var authToken = ""
var isAuthenticating = false
var enabled = true
var host string
var currentMotto string

func main() {

	ext.Intercept(out.CHAT, out.SHOUT, out.WHISPER).With(handleCommand)

	ext.Connected(func(e g.ConnectArgs) {
		authToken = ""

		if e.Host == "game-obr.habbo.com" {
			host = "BR"
		}
		if e.Host == "game-ous.habbo.com" {
			host = "COM"
		}
		if e.Host == "game-oes.habbo.com" {
			host = "ES"
		}

		profileMgr.Updated(func(e profile.Args) {
			if authToken != "" || isAuthenticating {
				return
			}
			isAuthenticating = true
			currentMotto = profileMgr.CustomData
			go authenticateLegacy(profileMgr.Name, host)
		})

	})

	ext.Run()

}

func handleCommand(e *g.Intercept) {
	msg := e.Packet.ReadString()
	args := strings.Split(msg, " ")
	if args[0] == ":tracker" {

		if authToken == "" {
			ext.Send(in.SYSTEM_BROADCAST, []byte("You must authenticate first, type :authenticate to try authenticating."))
			return
		}

		if !enabled {
			enabled = true
			ext.Send(in.SYSTEM_BROADCAST, []byte("[TSA] Tracker enabled."))
		} else {
			enabled = false
			ext.Send(in.SYSTEM_BROADCAST, []byte("[TSA] Tracker disabled."))
		}
		e.Block()
	}

	if args[0] == ":authenticate" {
		currentMotto = profileMgr.CustomData
		go authenticateLegacy(profileMgr.Name, host)
		e.Block()
	}

}
func authenticateLegacy(habboName, server string) {
	log.Println("Starting authenticateLegacy function")
	data := map[string]string{
		"habboName": habboName,
		"server":    server,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}
	log.Println("JSON marshaled successfully:", string(dataBytes))

	resp, err := http.Post("https://legacyhabbo.me/generate", "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		log.Println("Error sending POST request:", err)
		return
	}
	defer resp.Body.Close()
	log.Println("POST request sent successfully")

	if resp.StatusCode != http.StatusOK {
		ext.Send(out.UPDATE, int16(6), currentMotto)
		ext.Send(out.UPDATE, int16(6), currentMotto)
		ext.Send(out.UPDATE, int16(6), currentMotto)
		log.Printf("Received non-OK response: %s\n", resp.Status)
		return
	}
	log.Println("Received OK response from server")

	var response map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println("Error decoding response JSON:", err)
		return
	}
	log.Println("Response JSON decoded successfully:", response)

	oldMotto, ok := response["previousMotto"]
	if !ok {
		log.Println("Motto not found in response")
		return
	}

	motto, ok := response["motto"]
	if !ok {
		log.Println("Motto not found in response")
		return
	}
	log.Println("Motto found in response:", motto)

	ext.Send(out.UPDATE, int16(6), motto)
	log.Println("Sent UPDATE with motto " + motto)

	loginData := map[string]string{
		"habboName": habboName,
		"server":    server,
		"motto":     motto,
	}
	loginDataBytes, err := json.Marshal(loginData)
	if err != nil {
		log.Println("Error marshaling login JSON:", err)
		return
	}
	log.Println("Login JSON marshaled successfully:", string(loginDataBytes))

	loginResp, err := http.Post("https://legacyhabbo.me/login", "application/json", bytes.NewBuffer(loginDataBytes))
	if err != nil {
		ext.Send(out.UPDATE, int16(6), oldMotto)
		ext.Send(out.UPDATE, int16(6), oldMotto)
		ext.Send(out.UPDATE, int16(6), oldMotto)
		ext.Send(in.SYSTEM_BROADCAST, []byte("There was an error authenticating you in TSA Tracker."))
		log.Println("Error sending login POST request:", err)
		return
	}
	defer loginResp.Body.Close()
	log.Println("Login POST request sent successfully")

	if loginResp.StatusCode != http.StatusOK {

		log.Printf("Login received non-OK response for login: %s\n", loginResp.Status)

	} else {
		log.Println("Successfully sent login request")

		var loginResponse map[string]string
		if err := json.NewDecoder(loginResp.Body).Decode(&loginResponse); err != nil {
			log.Println("Error decoding login response JSON:", err)
			return
		}
		log.Println("Login response JSON decoded successfully:", loginResponse)

		if authToken, ok = loginResponse["authToken"]; ok {
			log.Println("authToken found in response:", authToken)
		} else {
			log.Println("authToken not found in response")
		}

		log.Println("Sent UPDATE with motto" + oldMotto)
		ext.Send(in.SYSTEM_BROADCAST, []byte("You have been authenticated successfully in TSA Tracker and Tracker was enabled automatically, disable it by typing :tracker."))
		ext.Send(out.UPDATE, int16(6), oldMotto)
		ext.Send(out.UPDATE, int16(6), oldMotto)
		ext.Send(out.UPDATE, int16(6), oldMotto)
	}
}
