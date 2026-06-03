package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"

	"qris-payment/internal/handler"
	"qris-payment/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[ENV] No .env file, using OS environment")
	}

	storagePath := getEnv("STORAGE_FILE", "./data/payments.json")
	store, err := service.NewStore(storagePath)
	if err != nil {
		log.Fatalf("[Store] Init failed: %v", err)
	}
	log.Printf("[Store] ✅ JSON storage: %s", storagePath)

	paymentSvc := service.NewPaymentService(store)
	paymentHandler := handler.NewPaymentHandler(paymentSvc)

	engine := html.New("./templates", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views:   engine,
		AppName: "QRIS Payment Testing",
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path}\n",
	}))
	app.Use(cors.New())

	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:order_id", fiberws.New(paymentHandler.WebSocket))

	simulateHandler := handler.NewSimulateHandler(paymentSvc)

	api := app.Group("/api")
	api.Post("/payment/create", paymentHandler.CreatePayment)
	api.Get("/payment/:order_id", paymentHandler.GetStatus)

	sim := api.Group("/simulate")
	sim.Post("/create", simulateHandler.CreateSimulate)
	sim.Post("/pay/:order_id", simulateHandler.PaySimulate)
	sim.Get("/pay/:order_id", simulateHandler.PaySimulate)
	sim.Get("/status/:order_id", simulateHandler.GetSimulateStatus)

	app.Post("/webhook/midtrans", paymentHandler.Webhook)

	// Serve React frontend SPA at the root
	app.Static("/", "./frontend/dist")
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./frontend/dist/index.html")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	host := getEnv("APP_HOST", "0.0.0.0")
	port := getEnv("APP_PORT", "3002")
	addr := fmt.Sprintf("%s:%s", host, port)

	publicURL := getEnv("PUBLIC_URL", "http://"+addr)

	log.Println("╔══════════════════════════════════════════════╗")
	log.Println("║     QRIS Payment Testing — GoFiber           ║")
	log.Println("╠══════════════════════════════════════════════╣")
	log.Printf("║  Web    : http://%s\n", addr)
	log.Printf("║  WS     : ws://%s/ws/:order_id\n", addr)
	log.Printf("║  Webhook: %s/webhook/midtrans\n", publicURL)
	log.Println("╚══════════════════════════════════════════════╝")

	if err := app.Listen(addr); err != nil {
		log.Fatalf("[APP] Server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
