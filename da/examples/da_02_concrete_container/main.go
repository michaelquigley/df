package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/michaelquigley/df/da"
)

// configuration for our application
type Config struct {
	AppName     string `yaml:"app_name" json:"app_name"`
	DatabaseURL string `yaml:"database_url" json:"database_url"`
	CacheURL    string `yaml:"cache_url" json:"cache_url"`
	Port        int    `yaml:"port" json:"port"`
}

// concrete container with explicit types
type App struct {
	Config *Config `da:"-"` // skip - not a component

	// infrastructure components (ordered first)
	Database *Database `da:"order=1"`
	Cache    *Cache    `da:"order=2"`

	// services (ordered after infrastructure)
	Services struct {
		Users *UserService `da:"order=10"`
		Auth  *AuthService `da:"order=20"`
	}

	// api server (ordered last)
	API *APIServer `da:"order=100"`
}

// database connection
type Database struct {
	URL       string
	connected bool
}

func NewDatabase(url string) *Database {
	return &Database{URL: url}
}

func (d *Database) Start() error {
	fmt.Printf("connecting to database: %s\n", d.URL)
	d.connected = true
	return nil
}

func (d *Database) Stop() error {
	fmt.Printf("disconnecting from database: %s\n", d.URL)
	d.connected = false
	return nil
}

func (d *Database) IsConnected() bool {
	return d.connected
}

// cache connection
type Cache struct {
	URL       string
	connected bool
}

func NewCache(url string) *Cache {
	return &Cache{URL: url}
}

func (c *Cache) Start() error {
	fmt.Printf("connecting to cache: %s\n", c.URL)
	c.connected = true
	return nil
}

func (c *Cache) Stop() error {
	fmt.Printf("disconnecting from cache: %s\n", c.URL)
	c.connected = false
	return nil
}

// user service with dependencies
type UserService struct {
	db    *Database
	cache *Cache
}

func NewUserService() *UserService {
	return &UserService{}
}

// Wire implements Wireable[App] - receives container for dependency wiring
func (s *UserService) Wire(app *App) error {
	s.db = app.Database
	s.cache = app.Cache
	fmt.Println("user service wired to database and cache")
	return nil
}

func (s *UserService) Start() error {
	fmt.Println("starting user service")
	return nil
}

func (s *UserService) Stop() error {
	fmt.Println("stopping user service")
	return nil
}

func (s *UserService) GetUser(id string) string {
	return fmt.Sprintf("user-%s (db connected: %v)", id, s.db.IsConnected())
}

// auth service with dependencies
type AuthService struct {
	db    *Database
	users *UserService
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Wire implements Wireable[App]
func (s *AuthService) Wire(app *App) error {
	s.db = app.Database
	s.users = app.Services.Users
	fmt.Println("auth service wired to database and user service")
	return nil
}

func (s *AuthService) Start() error {
	fmt.Println("starting auth service")
	return nil
}

func (s *AuthService) Stop() error {
	fmt.Println("stopping auth service")
	return nil
}

// api server
type APIServer struct {
	port  int
	db    *Database
	users *UserService
	auth  *AuthService
}

func NewAPIServer(port int) *APIServer {
	return &APIServer{port: port}
}

// Wire implements Wireable[App]
func (s *APIServer) Wire(app *App) error {
	s.db = app.Database
	s.users = app.Services.Users
	s.auth = app.Services.Auth
	fmt.Println("api server wired to all services")
	return nil
}

func (s *APIServer) Start() error {
	fmt.Printf("starting api server on port %d\n", s.port)
	return nil
}

func (s *APIServer) Stop() error {
	fmt.Printf("stopping api server on port %d\n", s.port)
	return nil
}

func main() {
	// create a temporary config file for this example
	configPath := createTempConfig()
	defer os.Remove(configPath)

	// load configuration
	cfg := &Config{}
	if err := da.Config(cfg,
		da.FileLoader(configPath),
	); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("loaded config: %+v\n\n", cfg)

	// build container explicitly with all components
	app := &App{
		Config:   cfg,
		Database: NewDatabase(cfg.DatabaseURL),
		Cache:    NewCache(cfg.CacheURL),
		API:      NewAPIServer(cfg.Port),
	}
	app.Services.Users = NewUserService()
	app.Services.Auth = NewAuthService()

	// wire dependencies (calls Wire on all Wireable[App] components)
	fmt.Println("=== wiring ===")
	if err := da.Wire(app); err != nil {
		log.Fatal(err)
	}

	// start all components (calls Start on all Startable components)
	fmt.Println("\n=== starting ===")
	if err := da.Start(app); err != nil {
		da.Stop(app)
		log.Fatal(err)
	}

	// use the services
	fmt.Println("\n=== using services ===")
	user := app.Services.Users.GetUser("123")
	fmt.Printf("got user: %s\n", user)

	// stop all components (calls Stop on all Stoppable components in reverse order)
	fmt.Println("\n=== stopping ===")
	if err := da.Stop(app); err != nil {
		log.Printf("error during shutdown: %v", err)
	}

	fmt.Println("\n=== done ===")
}

func createTempConfig() string {
	content := `app_name: concrete container example
database_url: postgres://localhost:5432/mydb
cache_url: redis://localhost:6379
port: 8080
`
	path := filepath.Join(os.TempDir(), "da_example_config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		log.Fatal(err)
	}
	return path
}
