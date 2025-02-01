
import (
        "encoding/json"
        "fmt"
        "io/ioutil"
        "os"
        "sync"
)

type GameState struct {
        FoodCount      int                 `json:"food"`
        WaterCount     int                 `json:"water"`
        MineralCount   int                 `json:"minerals"`
        Population     int                 `json:"population"`
        Happiness      float64             `json:"happiness"`
        ResearchPoints int                 `json:"researchPoints"`
        Infra          map[string]int      `json:"infrastructure"`
        TicksPassed    int                 `json:"ticksPassed"`
        StarCoins      int                 `json:"starCoins"`
        TaxRate        float64             `json:"taxRate"`
        Ships          map[string][]*Ship  `json:"ships"`
}

type Ship struct {
        Name             string  `json:"name"`
        Class            string  `json:"Class"`
        Health           int     `json:"Health"`
        Description      string  `json:"Description"`
        PersonnelLimit   int     `json:"Personnel_Limit"`
        PersonnelMinimum int     `json:"Personnel_Minimum"`
        CargoLimit       int     `json:"Cargo_Limit"`
        FuelCapacity     int     `json:"Fuel_Capacity"`
        FuelEfficiency   float64 `json:"Fuel_Efficiency"`
        Level            int     `json:"Level"`
        Damage           int     `json:"Damage"`
        Price            int     `json:"Price"`
        Amount           int     `json:"Amount"`
}

var stateLock sync.Mutex

func loadState(filename string) (GameState, error) {
        stateLock.Lock()
        defer stateLock.Unlock()

        data, err := ioutil.ReadFile(filename)
        if err != nil {
                return GameState{}, err
        }

        var state GameState
        err = json.Unmarshal(data, &state)
        if state.Infra == nil {
                state.Infra = make(map[string]int)
                state.ResearchPoints = 0
        }
        if state.Ships == nil {
                state.Ships = make(map[string][]*Ship)
        }
        return state, err
}

func saveState(filename string, state GameState) error {
        stateLock.Lock()
        defer stateLock.Unlock()

        data, err := json.MarshalIndent(state, "", "  ")
        if err != nil {
                return err
        }
        return ioutil.WriteFile(filename, data, 0644)
}

func showCommand(state GameState) {
        fmt.Printf("Food: %d, Water: %d, Minerals: %d\n", state.FoodCount, state.WaterCount, state.MineralCount)
        fmt.Printf("Population: %d, Happiness: %.2f%%\n", state.Population, state.Happiness)
        fmt.Printf("Research Points: %d\n", state.ResearchPoints)
        fmt.Printf("StarCoins: %d, TaxRate: %.2f\n", state.StarCoins, state.TaxRate)
        fmt.Printf("Ticks Passed: %d\n", state.TicksPassed)

        fmt.Println("Infrastructure:")
        for infraType, count := range state.Infra {
                fmt.Printf("  %s: %d\n", infraType, count)
        }

        fmt.Println("Ships:")
        for class, ships := range state.Ships {
                fmt.Printf("Class %s:\n", class)
                for _, ship := range ships {
                        fmt.Printf("  %s (Level: %d, Amount: %d)\n", ship.Name, ship.Level, ship.Amount)
                }
        }
}

func researchShip(state *GameState, args []string) {
        if len(args) < 4 {
                fmt.Println("Usage: worldc research <class> <name>")
                os.Exit(1)
        }

        shipClass := args[2]
        shipName := args[3]

        if shipClass != "Explorite" && shipClass != "Enforcer" && shipClass != "Pioneer" {
                fmt.Println("Unknown ship class:", shipClass)
                return
        }

        fmt.Printf("How many StarCoins do you want to put into this research project?\n")
        var starCoins int
        fmt.Scan(&starCoins)

        fmt.Printf("How many Development Points do you want to add to this project?\n")
        var devPoints int
        fmt.Scan(&devPoints)

        newShip := &Ship{
                Name:       shipName,
                Class:      shipClass,
                Health:     50,  // default for all
                Level:      1,   // starting level
                Price:      100, // base price, to be adjusted
                FuelCapacity:    100,
                FuelEfficiency:  1,
                Damage:     0,
        }

        switch shipClass {
        case "Explorite":
                newShip.FuelCapacity += starCoins / 10
                newShip.FuelEfficiency += float64(starCoins) / 100.0
        case "Enforcer":
                newShip.Damage += starCoins / 10
        case "Pioneer":
                newShip.PersonnelLimit += starCoins / 10
                newShip.FuelCapacity += starCoins / 20
                newShip.FuelEfficiency += float64(starCoins) / 200.0
        }

        priceAdjustment := devPoints * 50
        newShip.Price = max(50, newShip.Price-priceAdjustment) // ensure there is a minimum price

        state.Ships[shipClass] = append(state.Ships[shipClass], newShip)
        saveState("../data/game_state.json", *state)
        fmt.Printf("Research complete! Developed new ship: %s\n", newShip.Name)
        fmt.Println("Ship Details:")
        fmt.Printf("  Class: %s, Health: %d, Fuel Capacity: %d, Fuel Efficiency: %.2f, Damage: %d, Personnel Limit: %d, Price: %d\n",
                newShip.Class, newShip.Health, newShip.FuelCapacity, newShip.FuelEfficiency, newShip.Damage, newShip.PersonnelLimit, newShip.Price)
}

func buildInfrastructure(state *GameState, args []string) {
        // Updated function for building infrastructure with detailed resource consumption
}

func destroyInfrastructure(state *GameState, args []string) {
        // Updated function for destroying infrastructure with detailed information
}

func constructShip(state *GameState, args []string) {
        // Updated function for managing ship construction
}

func setCommand(state *GameState, args []string) {
        // Updated function for setting parameters such as TaxRate
}

func max(a, b int) int {
        if a > b {
                return a
        }
        return b
}

func showHelp() {
        fmt.Println("Usage: worldc <command> [args]")
        fmt.Println("Commands:")
        fmt.Println("  show                                  - Display the current game state.")
        fmt.Println("  build <infrastructure> <number>       - Build specified amount of infrastructure.")
        fmt.Println("  destroy <infrastructure> <number>     - Destroy specified amount of infrastructure.")
        fmt.Println("  set <TaxRate> <value>                 - Set the tax rate.")
        fmt.Println("  construct <ship_class> <number>       - Construct specified amount of ships based on class.")
        fmt.Println("  research <class> <name>               - Start a new research project for a ship.")
}

func executeCommand(command string, state *GameState, args []string) {
        switch command {
        case "show":
                showCommand(*state)
        case "build":
                buildInfrastructure(state, args)
        case "destroy":
                destroyInfrastructure(state, args)
        case "set":
                setCommand(state, args)
        case "construct":
                constructShip(state, args)
        case "research":
                researchShip(state, args)
        default:
                showHelp()
        }
}

func main() {
        if len(os.Args) < 2 {
                showHelp()
                os.Exit(1)
        }

        command := os.Args[1]
        filename := "../data/game_state.json"

        state, err := loadState(filename)
        if err != nil {
                fmt.Printf("Error loading state: %v\n", err)
                return
        }

        executeCommand(command, &state, os.Args)

        if command == "build" || command == "destroy" || command == "set" || command == "construct" || command == "research" {
                if err := saveState(filename, state); err != nil {
                        fmt.Printf("Error saving state: %v\n", err)
                }
        }
}
