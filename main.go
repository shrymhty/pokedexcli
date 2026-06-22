package main

import (
	"bufio"
	"math/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shrymhty/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name		string
	description	string
	callback	func(*config) error	
}

type config struct {
	Next string
	Previous string
	cache *pokecache.Cache
	Area   string
	PokeName string
	Pokedex map[string]pokemonDetails
}


type locationResponse struct {
	Count	 int	 `json:"count"`
	Next	 string	 `json:"next"`
	Previous string	 `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`	
	}
}

type exploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} 	`json:"pokemon"`
	}	`json:"pokemon_encounters"`
}

type pokemonDetails struct {
	Name string	`json:"name"`
	BaseExperience int `json:"base_experience"`
	Height int `json:"height"`
	Weight int `json:"weight"`
	PokeStats []struct {
		BaseStat int `json:"base_stat"`
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

var supportedCommands map[string]cliCommand 

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	cache := pokecache.NewCache(5 * time.Second)

	supportedCommands = map[string]cliCommand{
        "exit": {
            name:        "exit",
            description: "Exiting the pokedex cli",
            callback:    commandExit,
        },
        "help": {
            name:        "help",
            description: "how to use the pokedex cli",
            callback:    commandHelp,
        },
		"map": {
			name: "map",
			description: "displays the names of 20 location areas in the Pokemon world",
			callback: commandMap,
		},
		"mapb": {
			name: "mapb",
			description: "displays the previous 20 locations",
			callback: commandMapb,
		},
		"explore": {
			name: "explore",
			description: "list of all the Pokémon located in the area. Usage: \"explore <area_name>\"",
			callback: commandExplore,
		},
		"catch": {
			name: "catch",
			description: "Try to catch the pokemon. Usage: \"catch <pokemon_name>\"",
			callback: commandCatch,
		},
		"inspect": {
			name: "inspect",
			description: "It takes the name of a Pokemon and prints the name, height, weight, stats and type(s) of the Pokemon. Usage: \"inspect <pokemon_name>\"",
			callback: commandInspect,
		},
		"pokedex": {
			name: "pokedex",
			description: "Prints a list of all the names of the Pokemon the you have caught",
			callback: commandPokedex,
		},
    }

	cfg := &config{
		cache: cache,
		Pokedex: make(map[string]pokemonDetails),
	}

	fmt.Println("Welcome to the Pokedex!")

	for {
		fmt.Print("Pokedex > ")		
		if !scanner.Scan() {
			break // EOF or error
		}

		inputs := scanner.Text()
		input := strings.Fields(inputs)

		command, ok := supportedCommands[input[0]]
		if !ok {
			fmt.Println("Unknown Command")
			continue
		}

		if len(input) > 1 {
			if input[0] == "explore" {
				cfg.Area = input[1]
			} 
			if input[0] == "catch" || input[0] == "inspect" {
				cfg.PokeName = input[1]
			}
		}

		if err := command.callback(cfg); err != nil {
			fmt.Println(err)
		}
		
	}
}

func commandExit(cfg *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config) error {
	fmt.Println("Usage:")
	for _, command := range supportedCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(cfg *config) error {

	url := cfg.Next

	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"
	}

	if val, ok := cfg.cache.Get(url); ok {
		var data locationResponse
		err := json.Unmarshal(val, &data)
		if err != nil {
			return err
		}
		cfg.Next = data.Next
		cfg.Previous = data.Previous

		for _, loc := range data.Results {
			fmt.Println(loc.Name)
		}

		return nil
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

    cfg.cache.Add(url, body)

    var data locationResponse
    if err := json.Unmarshal(body, &data); err != nil {
        return err
    }

    cfg.Next = data.Next
    cfg.Previous = data.Previous

    for _, loc := range data.Results {
        fmt.Println(loc.Name)
    }

    return nil
}

func commandMapb(cfg *config) error {
	url := cfg.Previous

	if url == "" {
		return fmt.Errorf("you are on the first page")
	}

	if val, ok := cfg.cache.Get(url); ok {
		var data locationResponse
		err := json.Unmarshal(val, &data)
		if err != nil {
			return err
		}
		cfg.Next = data.Next
		cfg.Previous = data.Previous

		for _, loc := range data.Results {
			fmt.Println(loc.Name)
		}

		return nil
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
    if err != nil {
        return err
    }

    var data locationResponse
    err = json.Unmarshal(body, &data)
    if err != nil {
        return err
    }

    cfg.cache.Add(url, body)

	cfg.Next = data.Next
    cfg.Previous = data.Previous

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

func commandExplore(cfg *config) error {
	url := "https://pokeapi.co/api/v2/location-area/"+cfg.Area

	if val, ok := cfg.cache.Get(url); ok {
		var data exploreResponse
		err := json.Unmarshal(val, &data)
		if err != nil {
			return err
		}

		fmt.Println("Exploring "+cfg.Area+"...")
		fmt.Println("Found Pokemon:")
		for _, encounter := range data.PokemonEncounters {
			fmt.Println(encounter.Pokemon.Name)
		}
		
		return nil
	}

	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println("Exploring "+cfg.Area+"...")

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	
	var data exploreResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	
	fmt.Println("Found Pokemon:")
	for _, encounter := range data.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}
	cfg.cache.Add(url, body)

	return nil
}

func commandCatch(cfg *config) error {
	url := "https://pokeapi.co/api/v2/pokemon/" + cfg.PokeName

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var details pokemonDetails
	err = json.NewDecoder(res.Body).Decode(&details)
	if err != nil {
		return err
	}

	fmt.Println("Throwing a Pokeball at "+cfg.PokeName+"...")
	threshold := 100 - (details.BaseExperience / 2)

	if threshold <= 5 {
		threshold = 5
	}
	chance := rand.Intn(100)
	if chance <= threshold {
		fmt.Printf("%s was caught!\n", cfg.PokeName)
		cfg.Pokedex[cfg.PokeName] = details
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", cfg.PokeName)
	}

	return nil
}

func commandInspect(cfg *config) error {
	pokemon, ok := cfg.Pokedex[cfg.PokeName]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\nHeight: %d\nWeight: %d\n", pokemon.Name, pokemon.Height, pokemon.Height)	
	fmt.Println("Stats:")
	for _, stat := range pokemon.PokeStats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, poketype := range pokemon.Types {
		fmt.Printf("  -%s\n", poketype.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range cfg.Pokedex {
		fmt.Printf("  - %s\n", pokemon.Name)
	}
	return nil
}