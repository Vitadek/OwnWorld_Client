package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- Structs ---

type Colony struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Location  []int          `json:"location"`
	Buildings map[string]int `json:"buildings"`
	Food      int            `json:"food"`
	Water     int            `json:"water"`
	Pop       int            `json:"population"`
}

type ColonySummary struct {
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	Location  []int  `json:"location"`
}

type GameStateResponse struct {
	MyColonies []Colony       `json:"my_colonies"`
	WorldIndex []ColonySummary `json:"world_index"`
	StarCoins  int            `json:"starCoins"`
	Ticks      int            `json:"ticksPassed"`
}

// --- Globals ---

var serverURL = "http://localhost:8080"
var currentUserID = 0
var currentUsername = ""

// --- Networking ---

func postAuth(endpoint, username, password string) (int, error) {
	payload := map[string]string{"username": username, "password": password}
	jsonPayload, _ := json.Marshal(payload)
	
	resp, err := http.Post(serverURL + endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil { return 0, err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("auth failed")
	}

	var res map[string]int
	json.NewDecoder(resp.Body).Decode(&res)
	return res["user_id"], nil
}

func getServerState() (GameStateResponse, error) {
	req, _ := http.NewRequest("GET", serverURL + "/state", nil)
	req.Header.Set("X-User-ID", strconv.Itoa(currentUserID))
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil { return GameStateResponse{}, err }
	defer resp.Body.Close()

	var state GameStateResponse
	json.NewDecoder(resp.Body).Decode(&state)
	return state, nil
}

func postBuild(colonyID int, structure string) error {
	payload := map[string]interface{}{
		"user_id": currentUserID,
		"colony_id": colonyID,
		"structure": structure,
	}
	jsonPayload, _ := json.Marshal(payload)
	
	resp, err := http.Post(serverURL + "/build", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil { return err }
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("build failed (not owner?)")
	}
	return nil
}

// --- Menus ---

func authMenu(reader *bufio.Reader) bool {
	fmt.Println("\n=== AUTHENTICATION ===")
	fmt.Println("1. Login")
	fmt.Println("2. Register")
	fmt.Println("3. Exit")
	fmt.Print("Select: ")
	
	opt, _ := reader.ReadString('\n')
	opt = strings.TrimSpace(opt)
	
	if opt == "3" { os.Exit(0) }
	
	fmt.Print("Username: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	
	fmt.Print("Password: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)
	
	endpoint := "/login"
	if opt == "2" { endpoint = "/register" }
	
	id, err := postAuth(endpoint, user, pass)
	if err != nil {
		fmt.Println("Authentication failed:", err)
		return false
	}
	
	currentUserID = id
	currentUsername = user
	fmt.Printf("Welcome, Commander %s (ID: %d)\n", user, id)
	return true
}

func mainMenu() {
	reader := bufio.NewReader(os.Stdin)

	// Auth Loop
	for currentUserID == 0 {
		if !authMenu(reader) {
			time.Sleep(1 * time.Second)
		}
	}

	for {
		state, err := getServerState()
		if err != nil {
			fmt.Println("Connection lost:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("\n=== COMMAND CENTER ===")
		fmt.Printf("User: %s | Ticks: %d\n", currentUsername, state.Ticks)
		fmt.Println("--------------------------")
		fmt.Println("1. My Colonies")
		fmt.Println("2. World Index (Map)")
		fmt.Println("3. Refresh")
		fmt.Println("4. Logout")
		fmt.Print("Select: ")

		input, _ := reader.ReadString('\n')
		switch strings.TrimSpace(input) {
		case "1":
			myColoniesMenu(reader, state)
		case "2":
			worldIndexMenu(state)
		case "3":
			continue
		case "4":
			currentUserID = 0
			mainMenu() // Recursively restart to auth
		}
	}
}

func myColoniesMenu(reader *bufio.Reader, state GameStateResponse) {
	fmt.Println("\n--- MY COLONIES ---")
	if len(state.MyColonies) == 0 {
		fmt.Println("No colonies found.")
		return
	}
	
	for i, c := range state.MyColonies {
		fmt.Printf("%d. %s [Pop: %d | Food: %d]\n", i+1, c.Name, c.Pop, c.Food)
	}
	fmt.Println("B. Build")
	fmt.Println("R. Return")
	
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToUpper(input))
	
	if input == "B" {
		fmt.Print("Colony # to build in: ")
		numStr, _ := reader.ReadString('\n')
		num, _ := strconv.Atoi(strings.TrimSpace(numStr))
		if num > 0 && num <= len(state.MyColonies) {
			col := state.MyColonies[num-1]
			fmt.Print("Structure (farm/well): ")
			str, _ := reader.ReadString('\n')
			postBuild(col.ID, strings.TrimSpace(str))
		}
	}
}

func worldIndexMenu(state GameStateResponse) {
	fmt.Println("\n--- WORLD INDEX ---")
	for _, c := range state.WorldIndex {
		fmt.Printf("- %s (Owner: %s) @ %v\n", c.Name, c.OwnerName, c.Location)
	}
	fmt.Println("(Press Enter to return)")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func main() {
	urlPtr := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()
	serverURL = *urlPtr

	fmt.Println("Connecting to", serverURL)
	mainMenu()
}
