package main

import (
	"fmt"
	"hstin/gocoder/config"
	"hstin/gocoder/generate"
	"hstin/gocoder/geocoder"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func StartServer() {
	startTime := time.Now()
	gCoder, err := geocoder.NewGeocoder(config.Database)
	defer gCoder.Close()

	if err != nil {
		panic(err)
	}

	fmt.Println("Time to initialize:", time.Since(startTime))

	app := fiber.New()

	// CORS Middleware
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token")
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {

		if !config.EnableForward {
			return c.JSON(fiber.Map{
				"error": "Forward search is disabled",
			})
		}

		q := c.Query("q")
		maxResults := c.QueryInt("max", 10)
		complete := c.QueryBool("complete", false)
		useCache := c.QueryBool("cache", true)
		lang := c.Query("lang", "name")

		if complete {
			maxResults = -1
			useCache = false
		}

		result, cacheHit := gCoder.Search(q, maxResults, useCache, lang)
		if cacheHit {
			c.Set("X-Geocache", "HIT")
		} else {
			c.Set("X-Geocache", "MISS")
		}
		return c.JSON(result)
	})

	app.Get("/reverse", func(c *fiber.Ctx) error {

		if !config.EnableReverse {
			return c.JSON(fiber.Map{
				"error": "Reverse search is disabled",
			})
		}

		lat := c.Query("lat")
		lng := c.Query("lng")
		lang := c.Query("lang")

		latFloat, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			return err
		}

		lngFloat, err := strconv.ParseFloat(lng, 64)
		if err != nil {
			return err
		}

		return c.JSON(gCoder.Reverse(latFloat, lngFloat, lang))
	})

	app.Get("/node/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return err
		}

		lang := c.Query("lang")

		return c.JSON(gCoder.GetNode(id, lang))
	})

	_ = app.Listen(":3000")

}

func main() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "generate":
		generate.GenerateDatabase()
	case "server":
		StartServer()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Usage: %s <command> [options]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  generate    Generate files or resources")
	fmt.Println("  server      Start the HTTP server")
	fmt.Println("")
}
