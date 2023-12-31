package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {

	svcName, exists := os.LookupEnv("SVC_NAME")
	if !exists {
		slog.Error("SVC_NAME env var not set")
		panic("SVC_NAME env var not set")
	}

	e := echo.New()

	g := e.Group("/client/api/v1")

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
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(svcName + "/demo/api/v1/hello/" + name)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error",
				"error":   "Error greeting: " + err.Error(),
			})
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error",
				"error":   "Error reading response: " + err.Error(),
			})
		}
		slog.Info("Response from demmo", slog.String("response", string(b)))

		var hello Hello
		json.NewDecoder(bytes.NewReader(b)).Decode(&hello)

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
