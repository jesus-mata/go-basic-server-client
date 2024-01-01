package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {

	if len(os.Args) <= 1 {
		runServer()
	}

	action := os.Args[1]

	switch action {
	case "healthcheck":
		healthCheck("/client/api/v1/health")
	case "run":
		runServer()
	default:
		fmt.Println("Usage: ./main <action>")
		fmt.Println("action: run | healthcheck")
		fmt.Println("Example: ./main run or ./main healthcheck")
		fmt.Println("If no action is provided, then the application will run")
		panic("Unknown action")
	}
}

func runServer() {
	svcName, exists := os.LookupEnv("SVC_NAME")
	if !exists {
		slog.Error("SVC_NAME env var not set")
		panic("SVC_NAME env var not set")
	}

	e := echo.New()

	g := e.Group("/client/api/v1")

	g.GET("/health", func(c echo.Context) error {
		slog.Info("Health check")

		return c.JSON(http.StatusOK, map[string]string{
			"status": "OK",
		})
	})

	g.GET("/product/:id", func(c echo.Context) error {
		id := c.Param("id")

		slog.Info("Fetching product",
			slog.String("id", id),
		)

		client := http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get("https://dummyjson.com/products/" + id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error",
				"error":   "Error fetching products: " + err.Error(),
			})
		}
		defer resp.Body.Close()

		var product Product
		json.NewDecoder(resp.Body).Decode(&product)

		return c.JSON(http.StatusOK, product)
	})

	g.GET("/greet/:name", func(c echo.Context) error {
		name := c.Param("name")

		slog.Info("Greeting user",
			slog.String("name", name),
		)

		client := http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(svcName + "/demo/api/v1/hello/" + name)
		if err != nil {
			slog.Error("Error greeting: " + err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error",
				"error":   "Error greeting: " + err.Error(),
			})
		}
		defer resp.Body.Close()

		var hello Hello
		json.NewDecoder(resp.Body).Decode(&hello)

		return c.JSON(http.StatusOK, hello)
	})

	slog.Info("Starting server")
	slog.Info("Service name: " + svcName)
	e.Logger.Fatal(e.Start(":8080"))
}

type Product struct {
	ID                 int     `json:"id"`
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	Price              int     `json:"price"`
	DiscountPercentage float64 `json:"discountPercentage"`
	Rating             float64 `json:"rating"`
	Stock              int     `json:"stock"`
	Brand              string  `json:"brand"`
	Category           string  `json:"category"`
}

type Hello struct {
	Message string `json:"message"`
}

// Check the health of the application
// Make a request to the health endpoint and check the status code
// If the status code is 200, then the application is healthy and return exit code 0
// If the status code is not 200, then the application is not healthy and return exit code 1
func healthCheck(healthEndpoint string) {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := "http://localhost:8080/" + strings.TrimPrefix(healthEndpoint, "/")
	//demo/api/v1/health
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	fmt.Println("Status Code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}

}
