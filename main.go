package main

import (
	"crypto/tls"
	"golang-template-v3.1/pkg/config"
	loggerV1 "golang-template-v3.1/pkg/middleware/logger/v1"
	utilityV1 "golang-template-v3.1/pkg/middleware/utility/v1"
	"golang-template-v3.1/routers"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v3"
)

func init() {
	// Load environment name
	env := utilityV1.GetEnv("ENVIRONMENT")
	loadedEnv := strings.ToLower(env)

	fmt.Println("ENVIRONMENT:", strings.ToUpper(loadedEnv))
	// Load environment settings
	if envErr := godotenv.Load(fmt.Sprintf("./envs/.env-%s", loadedEnv)); envErr != nil {
		log.Fatal("Error loading env file:", envErr)
	}

	fmt.Println("PROJECT: ", utilityV1.GetEnv("PROJECT"))
	fmt.Println("DESCRIPTION: ", utilityV1.GetEnv("DESCRIPTION"))

	loggerV1.CreateInitialFolder()

	// Connect to DB
	config.ConnectPostgres()
}

func main() {
	app := fiber.New(fiber.Config{
		AppName:          utilityV1.GetEnv("PROJECT"),
		CaseSensitive:    true,
		DisableKeepalive: true,
		JSONEncoder:      json.Marshal,
		JSONDecoder:      json.Unmarshal,
	})

	// CORS configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET,POST,PUT,DELETE"},
		AllowHeaders: []string{"Origin, Content-Type, Accept, Authorization"},
	}))

	app.Use(logger.New())
	app.Use(recover.New())

	// Initialize routes
	routers.AppRoutes(app)

	// TLS Configuration
	if strings.ToUpper(utilityV1.GetEnv("SSL_MODE")) == "ENABLED" {
		fmt.Println("SSL_MODE: ENABLED")
		fmt.Println("CERTIFICATE:", utilityV1.GetEnv("SSL_CERTIFICATE"))
		fmt.Println("KEY:", utilityV1.GetEnv("SSL_KEY"))

		// LOAD CERTIFICATE
		cert, err := tls.LoadX509KeyPair(utilityV1.GetEnv("SSL_CERTIFICATE"), utilityV1.GetEnv("SSL_KEY"))
		if err != nil {
			log.Fatal(err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		// START THE SERVER WITH HTTPS
		tlsPort := fmt.Sprintf(":%s", utilityV1.GetEnv("PORT"))
		listener, err := tls.Listen("tcp", tlsPort, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to create TLS listener: %v", err)
		}

		if err := app.Listener(listener); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	} else {
		fmt.Println("SSL_MODE: DISABLED")
		log.Fatal(app.Listen(fmt.Sprintf(":%s", utilityV1.GetEnv("PORT"))))
	}

}
