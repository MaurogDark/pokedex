package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	callback    func(*Config) error
}

var cache *internal.Cache

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
	}
}

func commandHelp(_ *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, c := range commands() {
		fmt.Println(c.name + ": " + c.description)
	}
	return nil
}

func commandExit(_ *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandMap(c *Config) error {
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

func commandMapb(c *Config) error {
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

type result struct {
	Name string
	Url  string
}

type resultPage struct {
	Count    int
	Next     string
	Previous string
	Results  []result
}

func get_map(map_url string) (resultPage, Config) {
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
	page := resultPage{}
	err := json.Unmarshal(body, &page)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("MAPPED [" + map_url + "]\nPREV [" + page.Previous + "]\nNEXT [" + page.Next + "]")
	return page, Config{prev: page.Previous, next: page.Next}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cache = internal.NewCache(time.Second * 30)
	c := Config{prev: "", next: ""}
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		words := cleanInput(scanner.Text())
		command, ok := commands()[words[0]]
		if ok {
			command.callback(&c)
		} else {
			fmt.Println("Unrecognized command: " + words[0])
		}
	}
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
