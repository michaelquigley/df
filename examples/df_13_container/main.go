package main

import (
	"fmt"
	"log"

	"github.com/michaelquigley/df"
)

// configuration struct
type appConfig struct {
	ServerPort int    `df:"port"`
	DBHost     string `df:"db_host"`
	AppName    string `df:"app_name"`
}

// simple database service
type database struct {
	host      string
	connected bool
}

func (db *database) connect() {
	fmt.Printf("connecting to database at %s...\n", db.host)
	db.connected = true
}

// web server service
type server struct {
	port int
	name string
	db   *database
}

func (s *server) start() {
	fmt.Printf("starting %s server on port %d\n", s.name, s.port)
	if s.db != nil && s.db.connected {
		fmt.Println("server has database connection")
	}
}

// application that coordinates everything
type app struct {
	server *server
}

func (a *app) run() {
	fmt.Println("running application...")
	a.server.start()
}

func main() {
	fmt.Println("df container example")
	fmt.Println("====================")

	// configuration data that would typically come from a config file
	configData := map[string]any{
		"port":     8080,
		"db_host":  "localhost:5432",
		"app_name": "my awesome app",
	}

	// phase 1: bind configuration
	builder := df.NewBuilder()
	builder, err := builder.BindFrom(configData, &appConfig{})
	if err != nil {
		log.Fatalf("failed to bind config: %v", err)
	}
	fmt.Println("✓ configuration bound")

	// phase 2: register factories
	df.Factory(builder, func(c *df.Container) (*database, error) {
		config, found := df.Get[*appConfig](c)
		if !found {
			return nil, fmt.Errorf("config not found")
		}
		return &database{host: config.DBHost}, nil
	})

	df.Factory(builder, func(c *df.Container) (*server, error) {
		config, found := df.Get[*appConfig](c)
		if !found {
			return nil, fmt.Errorf("config not found")
		}
		return &server{
			port: config.ServerPort,
			name: config.AppName,
		}, nil
	})

	df.Factory(builder, func(c *df.Container) (*app, error) {
		return &app{}, nil
	})

	fmt.Println("✓ factories registered")

	// phase 3: create all objects
	builder, err = builder.Create()
	if err != nil {
		log.Fatalf("failed to create objects: %v", err)
	}
	fmt.Println("✓ objects created")

	// phase 4: wire dependencies
	builder.Link(func(c *df.Container) error {
		// connect database
		db, found := df.Get[*database](c)
		if !found {
			return fmt.Errorf("database not found")
		}
		db.connect()
		return nil
	}).Link(func(c *df.Container) error {
		// wire server with database
		server, found := df.Get[*server](c)
		if !found {
			return fmt.Errorf("server not found")
		}
		db, found := df.Get[*database](c)
		if !found {
			return fmt.Errorf("database not found")
		}
		server.db = db
		return nil
	}).Link(func(c *df.Container) error {
		// wire app with server
		app, found := df.Get[*app](c)
		if !found {
			return fmt.Errorf("app not found")
		}
		server, found := df.Get[*server](c)
		if !found {
			return fmt.Errorf("server not found")
		}
		app.server = server
		return nil
	})

	builder, err = builder.Wire()
	if err != nil {
		log.Fatalf("failed to wire dependencies: %v", err)
	}
	fmt.Println("✓ dependencies wired")

	// phase 5: get final container and run application
	container := builder.Build()

	application, found := df.Get[*app](container)
	if !found {
		log.Fatal("application not found in container")
	}

	fmt.Println()
	application.run()
}