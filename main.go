package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MaurogDark/pokedex/internal"
)

type Config struct {
	prev string
	next string
}

type CliCommand struct {
	name        string
	description string
	callback    func(*Config, string) error
}

var cache *internal.Cache
var pokedex map[string]Pokemon

func commands() map[string]CliCommand {
	return map[string]CliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Display the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the previous 20 locations",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location. Takes a location parameter",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon. Takes a Pokemon name parameter",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect a Pokemon. Takes a Pokemon name parameter",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show all the Pokemon you caught",
			callback:    commandPokedex,
		},
	}
}

func commandHelp(_ *Config, _ string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, c := range commands() {
		fmt.Println(c.name + ": " + c.description)
	}
	return nil
}

func commandExit(_ *Config, _ string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(c *Config, _ string) error {
	if c.next == "" && c.prev == "" {
		page, new_config := get_map("https://pokeapi.co/api/v2/location-area")
		for _, result := range page.Results {
			fmt.Println(result.Name)
		}
		*c = new_config
	} else {
		if c.next == "" {
			fmt.Println("you're on the last page")
		} else {
			page, new_config := get_map(c.next)
			for _, result := range page.Results {
				fmt.Println(result.Name)
			}
			*c = new_config
		}
	}
	return nil
}

func commandMapb(c *Config, _ string) error {
	if c.next == "" && c.prev == "" {
		page, new_config := get_map("https://pokeapi.co/api/v2/location-area")
		for _, result := range page.Results {
			fmt.Println(result.Name)
		}
		*c = new_config
	} else {
		if c.prev == "" {
			fmt.Println("you're on the first page")
		} else {
			page, new_config := get_map(c.prev)
			for _, result := range page.Results {
				fmt.Println(result.Name)
			}
			*c = new_config
		}
	}
	return nil
}

func commandExplore(_ *Config, param string) error {
	if len(param) < 1 {
		fmt.Println("The syntax is explore [location]")
	} else {
		fmt.Printf("Exploring %s...\n", param)
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", param)
		get_pokemans(url)
	}
	return nil
}

func commandCatch(_ *Config, param string) error {
	if len(param) < 1 {
		fmt.Println("The syntax is catch [pokemon]")
	} else {
		catch_pokemans(param)
	}
	return nil
}

func commandInspect(_ *Config, param string) error {
	if len(param) < 1 {
		fmt.Println("The syntax is inspect [pokemon]")
	} else {
		inspect_pokemans(param)
	}
	return nil
}

func commandPokedex(_ *Config, _ string) error {
	if len(pokedex) < 1 {
		fmt.Println("Your Pokedex is empty!")
	} else {
		fmt.Println("Your Pokedex:")
		for k := range pokedex {
			fmt.Printf("  - %s\n", k)
		}
	}
	return nil
}

type Result struct {
	Name string
	Url  string
}

type ResultPage struct {
	Count    int
	Next     string
	Previous string
	Results  []Result
}

type Pokemon_In_Area struct {
	Name string
	Url  string
}

type Pokemon_Encounter struct {
	Pokemon Pokemon_In_Area
}

type Pokemon_Area struct {
	Pokemon_Encounters []Pokemon_Encounter
}

type Pokemon_Inner_Stat struct {
	Name string
}

type Pokemon_Stat struct {
	Base_Stat int
	Stat      Pokemon_Inner_Stat
}

type Pokemon_Inner_Type struct {
	Name string
}

type Pokemon_Type struct {
	Type Pokemon_Inner_Type
}

type Pokemon struct {
	Id              int
	Name            string
	Base_Experience int
	Height          int
	Is_Default      bool
	Order           int
	Weight          int
	Stats           []Pokemon_Stat
	Types           []Pokemon_Type
}

func inspect_pokemans(pokemon_name string) {
	p, ok := pokedex[pokemon_name]
	if !ok {
		fmt.Printf("You don't have a %s!\n", pokemon_name)
		return
	}

	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Height: %d\n", p.Height)
	fmt.Printf("Weight: %d\n", p.Weight)
	fmt.Println("Stats:")
	for s := range p.Stats {
		fmt.Printf("  -%s: %d\n", p.Stats[s].Stat.Name, p.Stats[s].Base_Stat)
	}
	fmt.Println("Types:")
	if len(p.Types) < 1 {
		fmt.Println(" - none")
	} else {
		for t := range p.Types {
			fmt.Printf("  -%s\n", p.Types[t].Type.Name)
		}
	}
	fmt.Println("")
}

func catch_pokemans(pokemon_name string) {
	pokemon_url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemon_name)

	p, ok := pokedex[pokemon_name]
	if ok {
		fmt.Printf("You already have a %s!\n", pokemon_name)
		return
	}

	body, ok := cache.Get(pokemon_url)
	if !ok {
		res, err := http.Get(pokemon_url)
		if err != nil {
			log.Fatal(err)
		}

		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			fmt.Printf("No such Pokemon (code %d)\n", res.StatusCode)
			return
		}
		if err != nil {
			fmt.Printf("No such Pokemon (error: %s)\n", err)
			return
		}
	}

	p = Pokemon{}

	err := json.Unmarshal(body, &p)
	if err != nil {
		fmt.Printf("Failed to parse Pokemon (error: %s)\n", err)
		return
	}

	cache.Add(pokemon_url, body)

	chance := 100 - p.Base_Experience/3
	if chance < 5 {
		chance = 5
	}

	/*fmt.Print("<<< ")
	fmt.Print(p)
	fmt.Print(" >>>\n\n")*/

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon_name)
	roll := rand.IntN(100)

	fmt.Printf("<<< Need to roll under %d to catch %s, rolled %d >>>\n", chance, pokemon_name, roll)

	if roll < chance {
		fmt.Printf("%s was caught!\n", pokemon_name)
		pokedex[pokemon_name] = p
	} else {
		fmt.Printf("%s escaped!\n", pokemon_name)
	}
}

func get_pokemans(area_url string) {
	body, ok := cache.Get(area_url)
	if !ok {
		res, err := http.Get(area_url)
		if err != nil {
			log.Fatal(err)
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			fmt.Printf("Failed to explore the area (code %d)\n", res.StatusCode)
			return
		}
		if err != nil {
			fmt.Printf("Failed to explore the area (error: %s)\n", err)
			return
		}
	}

	area := Pokemon_Area{}

	err := json.Unmarshal(body, &area)
	if err != nil {
		fmt.Printf("Failed to parse area (error: %s)\n", err)
		return
	}

	cache.Add(area_url, body)

	if len(area.Pokemon_Encounters) < 1 {
		fmt.Println("You didn't find any Pokemon!")
	} else {
		fmt.Println("Found Pokemon:")
		for p := range area.Pokemon_Encounters {
			fmt.Printf(" - %s\n", area.Pokemon_Encounters[p].Pokemon.Name)
		}
	}
}

func get_map(map_url string) (ResultPage, Config) {
	body, ok := cache.Get(map_url)
	if !ok {
		res, err := http.Get(map_url)
		if err != nil {
			log.Fatal(err)
		}
		body, err = io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}
		cache.Add(map_url, body)
	}
	page := ResultPage{}
	err := json.Unmarshal(body, &page)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("MAPPED [" + map_url + "]\nPREV [" + page.Previous + "]\nNEXT [" + page.Next + "]")
	return page, Config{prev: page.Previous, next: page.Next}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cache = internal.NewCache(time.Second * 30)
	pokedex = map[string]Pokemon{}
	c := Config{prev: "", next: ""}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		if len(words) < 1 {
			fmt.Println("Enter help to see a list of commands")
			continue
		}

		command, ok := commands()[words[0]]
		if !ok {
			fmt.Println("Unrecognized command: " + words[0])
		} else {
			param := ""
			if len(words) > 1 {
				param = words[1]
			}
			if !ok {
				param = ""
			}
			command.callback(&c, param)
		}

	}
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
