package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var ServerURL = "http://localhost:8080"

// --- Models ---
type RegisterResponse struct {
	UserID   int    `json:"user_id"`
	SystemID string `json:"system_id"`
	Status   string `json:"status"`
}

type StatusResponse struct {
	UUID   string `json:"uuid"`
	Tick   int    `json:"tick"`
	Leader string `json:"leader"`
}

func main() {
	if url := os.Getenv("OWNWORLD_SERVER"); url != "" {
		ServerURL = url
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("OwnWorld Federation Client v2.1")
	fmt.Printf("Connected to: %s\n", ServerURL)
	fmt.Println("Commands: register, status, build, burn, launch, quit")

	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		parts := strings.Fields(text)

		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]

		switch cmd {
		case "status":
			doStatus()
		case "register":
			if len(parts) < 3 {
				fmt.Println("Usage: register <username> <password>")
				continue
			}
			doRegister(parts[1], parts[2])
		case "build":
			if len(parts) < 4 {
				fmt.Println("Usage: build <colony_id> <structure> <amount>")
				continue
			}
			amt, _ := strconv.Atoi(parts[3])
			colID, _ := strconv.Atoi(parts[1])
			doBuild(colID, parts[2], amt)
		case "burn":
			if len(parts) < 4 {
				fmt.Println("Usage: burn <colony_id> <item> <amount>")
				continue
			}
			amt, _ := strconv.Atoi(parts[3])
			colID, _ := strconv.Atoi(parts[1])
			doBurn(colID, parts[2], amt)
		case "launch":
			// New Feature from v1 Client
			if len(parts) < 4 {
				fmt.Println("Usage: launch <fleet_id> <dest_system_uuid> <distance>")
				continue
			}
			fleetID, _ := strconv.Atoi(parts[1])
			dist, _ := strconv.Atoi(parts[3])
			doLaunch(fleetID, parts[2], dist)
		case "quit", "exit":
			fmt.Println("Disconnecting...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command.")
		}
	}
}

func doStatus() {
	resp, err := http.Get(ServerURL + "/api/status")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	
	var s StatusResponse
	json.Unmarshal(body, &s)
	// Safety check for empty response
	if s.Leader == "" { s.Leader = "Unknown" }
	if s.UUID == "" { s.UUID = "Unknown" }
	
	// Truncate UUIDs for display if they are long enough
	leaderDisp := s.Leader
	if len(s.Leader) > 8 { leaderDisp = s.Leader[:8] }
	uuidDisp := s.UUID
	if len(s.UUID) > 8 { uuidDisp = s.UUID[:8] }

	fmt.Printf("Tick: %d | Leader: %s | UUID: %s\n", s.Tick, leaderDisp, uuidDisp)
}

func doRegister(user, pass string) {
	payload := map[string]string{"username": user, "password": pass}
	data, _ := json.Marshal(payload)
	
	resp, err := http.Post(ServerURL+"/api/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("Failed: %s\n", string(body))
		return
	}

	var r RegisterResponse
	json.Unmarshal(body, &r)
	fmt.Printf("Success! User ID: %d, System: %s\n", r.UserID, r.SystemID)
}

func doBuild(colID int, structure string, amount int) {
	payload := map[string]interface{}{
		"colony_id": colID,
		"structure": structure,
		"amount":    amount,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(ServerURL+"/api/build", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
}

func doBurn(colID int, item string, amount int) {
	payload := map[string]interface{}{
		"colony_id": colID,
		"item":      item,
		"amount":    amount,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(ServerURL+"/api/bank/burn", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Bank Receipt: %s\n", string(body))
}

func doLaunch(fleetID int, dest string, distance int) {
	payload := map[string]interface{}{
		"fleet_id":    fleetID,
		"dest_system": dest,
		"distance":    distance,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(ServerURL+"/api/fleet/launch", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Mission Status: %s\n", string(body))
}
