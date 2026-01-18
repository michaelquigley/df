package da

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test container for concrete container tests
type testConcreteApp struct {
	Config   *testConcreteConfig `da:"-"`
	Database *testConcreteDB     `da:"order=1"`
	Cache    *testConcreteCache  `da:"order=2"`
	Services struct {
		Auth *testConcreteAuth `da:"order=10"`
		API  *testConcreteAPI  `da:"order=20"`
	}
}

type testConcreteConfig struct {
	DatabaseURL string `json:"database_url" yaml:"database_url"`
	CacheURL    string `json:"cache_url" yaml:"cache_url"`
	Port        int    `json:"port" yaml:"port"`
}

type testConcreteDB struct {
	url     string
	wired   bool
	started bool
	stopped bool
}

func (d *testConcreteDB) Wire(app *testConcreteApp) error {
	d.url = app.Config.DatabaseURL
	d.wired = true
	return nil
}

func (d *testConcreteDB) Start() error {
	d.started = true
	return nil
}

func (d *testConcreteDB) Stop() error {
	d.stopped = true
	return nil
}

type testConcreteCache struct {
	url     string
	wired   bool
	started bool
	stopped bool
}

func (c *testConcreteCache) Wire(app *testConcreteApp) error {
	c.url = app.Config.CacheURL
	c.wired = true
	return nil
}

func (c *testConcreteCache) Start() error {
	c.started = true
	return nil
}

func (c *testConcreteCache) Stop() error {
	c.stopped = true
	return nil
}

type testConcreteAuth struct {
	db      *testConcreteDB
	wired   bool
	started bool
	stopped bool
}

func (a *testConcreteAuth) Wire(app *testConcreteApp) error {
	a.db = app.Database
	a.wired = true
	return nil
}

func (a *testConcreteAuth) Start() error {
	a.started = true
	return nil
}

func (a *testConcreteAuth) Stop() error {
	a.stopped = true
	return nil
}

type testConcreteAPI struct {
	db      *testConcreteDB
	cache   *testConcreteCache
	auth    *testConcreteAuth
	wired   bool
	started bool
	stopped bool
}

func (a *testConcreteAPI) Wire(app *testConcreteApp) error {
	a.db = app.Database
	a.cache = app.Cache
	a.auth = app.Services.Auth
	a.wired = true
	return nil
}

func (a *testConcreteAPI) Start() error {
	a.started = true
	return nil
}

func (a *testConcreteAPI) Stop() error {
	a.stopped = true
	return nil
}

func TestWire(t *testing.T) {
	app := &testConcreteApp{
		Config:   &testConcreteConfig{DatabaseURL: "postgres://localhost", CacheURL: "redis://localhost"},
		Database: &testConcreteDB{},
		Cache:    &testConcreteCache{},
	}
	app.Services.Auth = &testConcreteAuth{}
	app.Services.API = &testConcreteAPI{}

	err := Wire(app)
	assert.NoError(t, err)

	assert.True(t, app.Database.wired)
	assert.Equal(t, "postgres://localhost", app.Database.url)
	assert.True(t, app.Cache.wired)
	assert.Equal(t, "redis://localhost", app.Cache.url)
	assert.True(t, app.Services.Auth.wired)
	assert.Equal(t, app.Database, app.Services.Auth.db)
	assert.True(t, app.Services.API.wired)
	assert.Equal(t, app.Database, app.Services.API.db)
	assert.Equal(t, app.Cache, app.Services.API.cache)
	assert.Equal(t, app.Services.Auth, app.Services.API.auth)
}

func TestStart(t *testing.T) {
	app := &testConcreteApp{
		Config:   &testConcreteConfig{},
		Database: &testConcreteDB{},
		Cache:    &testConcreteCache{},
	}
	app.Services.Auth = &testConcreteAuth{}
	app.Services.API = &testConcreteAPI{}

	err := Start(app)
	assert.NoError(t, err)

	assert.True(t, app.Database.started)
	assert.True(t, app.Cache.started)
	assert.True(t, app.Services.Auth.started)
	assert.True(t, app.Services.API.started)
}

func TestStop(t *testing.T) {
	app := &testConcreteApp{
		Config:   &testConcreteConfig{},
		Database: &testConcreteDB{},
		Cache:    &testConcreteCache{},
	}
	app.Services.Auth = &testConcreteAuth{}
	app.Services.API = &testConcreteAPI{}

	err := Stop(app)
	assert.NoError(t, err)

	assert.True(t, app.Database.stopped)
	assert.True(t, app.Cache.stopped)
	assert.True(t, app.Services.Auth.stopped)
	assert.True(t, app.Services.API.stopped)
}

// test ordering
type testOrderedApp struct {
	First  *testOrderedComponent `da:"order=1"`
	Second *testOrderedComponent `da:"order=2"`
	Third  *testOrderedComponent `da:"order=3"`
}

type testOrderedComponent struct {
	name         string
	startOrder   int
	stopOrder    int
	startCounter *int
	stopCounter  *int
}

func (c *testOrderedComponent) Start() error {
	*c.startCounter++
	c.startOrder = *c.startCounter
	return nil
}

func (c *testOrderedComponent) Stop() error {
	*c.stopCounter++
	c.stopOrder = *c.stopCounter
	return nil
}

func TestStartOrder(t *testing.T) {
	startCounter := 0
	stopCounter := 0

	app := &testOrderedApp{
		First:  &testOrderedComponent{name: "first", startCounter: &startCounter, stopCounter: &stopCounter},
		Second: &testOrderedComponent{name: "second", startCounter: &startCounter, stopCounter: &stopCounter},
		Third:  &testOrderedComponent{name: "third", startCounter: &startCounter, stopCounter: &stopCounter},
	}

	err := Start(app)
	assert.NoError(t, err)

	assert.Equal(t, 1, app.First.startOrder)
	assert.Equal(t, 2, app.Second.startOrder)
	assert.Equal(t, 3, app.Third.startOrder)
}

func TestStopOrder(t *testing.T) {
	startCounter := 0
	stopCounter := 0

	app := &testOrderedApp{
		First:  &testOrderedComponent{name: "first", startCounter: &startCounter, stopCounter: &stopCounter},
		Second: &testOrderedComponent{name: "second", startCounter: &startCounter, stopCounter: &stopCounter},
		Third:  &testOrderedComponent{name: "third", startCounter: &startCounter, stopCounter: &stopCounter},
	}

	err := Stop(app)
	assert.NoError(t, err)

	// stop should be in reverse order
	assert.Equal(t, 3, app.First.stopOrder)
	assert.Equal(t, 2, app.Second.stopOrder)
	assert.Equal(t, 1, app.Third.stopOrder)
}

// test skip tag
type testSkipApp struct {
	Config   *testSkipComponent `da:"-"`
	Included *testSkipComponent
}

type testSkipComponent struct {
	started bool
}

func (c *testSkipComponent) Start() error {
	c.started = true
	return nil
}

func TestSkipTag(t *testing.T) {
	app := &testSkipApp{
		Config:   &testSkipComponent{},
		Included: &testSkipComponent{},
	}

	err := Start(app)
	assert.NoError(t, err)

	assert.False(t, app.Config.started, "config should be skipped")
	assert.True(t, app.Included.started, "included should be started")
}

// test slice traversal
type testSliceApp struct {
	Workers []*testSliceWorker
}

type testSliceWorker struct {
	name    string
	started bool
}

func (w *testSliceWorker) Start() error {
	w.started = true
	return nil
}

func TestSliceTraversal(t *testing.T) {
	app := &testSliceApp{
		Workers: []*testSliceWorker{
			{name: "worker1"},
			{name: "worker2"},
			{name: "worker3"},
		},
	}

	err := Start(app)
	assert.NoError(t, err)

	for _, w := range app.Workers {
		assert.True(t, w.started, "worker %s should be started", w.name)
	}
}

// test map traversal
type testMapApp struct {
	Handlers map[string]*testMapHandler
}

type testMapHandler struct {
	name    string
	started bool
}

func (h *testMapHandler) Start() error {
	h.started = true
	return nil
}

func TestMapTraversal(t *testing.T) {
	app := &testMapApp{
		Handlers: map[string]*testMapHandler{
			"get":    {name: "get"},
			"post":   {name: "post"},
			"delete": {name: "delete"},
		},
	}

	err := Start(app)
	assert.NoError(t, err)

	for _, h := range app.Handlers {
		assert.True(t, h.started, "handler %s should be started", h.name)
	}
}

// test wire error propagation
type testWireErrorApp struct {
	Failing *testWireErrorComponent
}

type testWireErrorComponent struct{}

func (c *testWireErrorComponent) Wire(app *testWireErrorApp) error {
	return errors.New("wire failed")
}

func TestWireError(t *testing.T) {
	app := &testWireErrorApp{
		Failing: &testWireErrorComponent{},
	}

	err := Wire(app)
	assert.Error(t, err)
	assert.Equal(t, "wire failed", err.Error())
}

// test start error propagation
type testStartErrorApp struct {
	Failing *testStartErrorComponent
}

type testStartErrorComponent struct{}

func (c *testStartErrorComponent) Start() error {
	return errors.New("start failed")
}

func TestStartError(t *testing.T) {
	app := &testStartErrorApp{
		Failing: &testStartErrorComponent{},
	}

	err := Start(app)
	assert.Error(t, err)
	assert.Equal(t, "start failed", err.Error())
}

// test stop continues on error
type testStopErrorApp struct {
	First  *testStopErrorComponent `da:"order=1"`
	Second *testStopErrorComponent `da:"order=2"`
}

type testStopErrorComponent struct {
	name    string
	fail    bool
	stopped bool
}

func (c *testStopErrorComponent) Stop() error {
	c.stopped = true
	if c.fail {
		return errors.New(c.name + " stop failed")
	}
	return nil
}

func TestStopContinuesOnError(t *testing.T) {
	app := &testStopErrorApp{
		First:  &testStopErrorComponent{name: "first", fail: false},
		Second: &testStopErrorComponent{name: "second", fail: true},
	}

	err := Stop(app)
	assert.Error(t, err)
	assert.Equal(t, "second stop failed", err.Error())

	// both should have been stopped even though second failed
	assert.True(t, app.First.stopped)
	assert.True(t, app.Second.stopped)
}

// test config loading
func TestConfigFromJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	content := `{"database_url": "postgres://localhost", "port": 8080}`
	err := os.WriteFile(configPath, []byte(content), 0644)
	assert.NoError(t, err)

	cfg := &testConcreteConfig{}
	err = Config(cfg, FileLoader(configPath))
	assert.NoError(t, err)

	assert.Equal(t, "postgres://localhost", cfg.DatabaseURL)
	assert.Equal(t, 8080, cfg.Port)
}

func TestConfigFromYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `database_url: postgres://localhost
cache_url: redis://localhost
port: 9090`
	err := os.WriteFile(configPath, []byte(content), 0644)
	assert.NoError(t, err)

	cfg := &testConcreteConfig{}
	err = Config(cfg, FileLoader(configPath))
	assert.NoError(t, err)

	assert.Equal(t, "postgres://localhost", cfg.DatabaseURL)
	assert.Equal(t, "redis://localhost", cfg.CacheURL)
	assert.Equal(t, 9090, cfg.Port)
}

func TestConfigOptionalMissingFile(t *testing.T) {
	cfg := &testConcreteConfig{Port: 3000}
	err := Config(cfg, OptionalFileLoader("/nonexistent/config.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port) // should keep default
}

func TestConfigRequiredMissingFile(t *testing.T) {
	cfg := &testConcreteConfig{}
	err := Config(cfg, FileLoader("/nonexistent/config.yaml"))
	assert.Error(t, err)
}

func TestConfigChainLoader(t *testing.T) {
	tempDir := t.TempDir()

	basePath := filepath.Join(tempDir, "base.yaml")
	baseContent := `database_url: postgres://base
port: 3000`
	err := os.WriteFile(basePath, []byte(baseContent), 0644)
	assert.NoError(t, err)

	overridePath := filepath.Join(tempDir, "override.yaml")
	overrideContent := `port: 8080`
	err = os.WriteFile(overridePath, []byte(overrideContent), 0644)
	assert.NoError(t, err)

	cfg := &testConcreteConfig{}
	err = Config(cfg, ChainLoader(
		FileLoader(basePath),
		FileLoader(overridePath),
	))
	assert.NoError(t, err)

	assert.Equal(t, "postgres://base", cfg.DatabaseURL) // from base
	assert.Equal(t, 8080, cfg.Port)                     // overridden
}

func TestConfigMultipleLoaders(t *testing.T) {
	tempDir := t.TempDir()

	basePath := filepath.Join(tempDir, "base.yaml")
	baseContent := `database_url: postgres://base
cache_url: redis://base`
	err := os.WriteFile(basePath, []byte(baseContent), 0644)
	assert.NoError(t, err)

	localPath := filepath.Join(tempDir, "local.yaml")
	localContent := `cache_url: redis://local`
	err = os.WriteFile(localPath, []byte(localContent), 0644)
	assert.NoError(t, err)

	cfg := &testConcreteConfig{}
	err = Config(cfg,
		FileLoader(basePath),
		OptionalFileLoader(localPath),
		OptionalFileLoader(filepath.Join(tempDir, "nonexistent.yaml")),
	)
	assert.NoError(t, err)

	assert.Equal(t, "postgres://base", cfg.DatabaseURL) // from base
	assert.Equal(t, "redis://local", cfg.CacheURL)      // overridden by local
}

func TestConfigUnsupportedExtension(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.txt")

	err := os.WriteFile(configPath, []byte("test"), 0644)
	assert.NoError(t, err)

	cfg := &testConcreteConfig{}
	err = Config(cfg, FileLoader(configPath))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config extension")
}

// test nil pointer handling
type testNilApp struct {
	Present *testNilComponent
	Nil     *testNilComponent
}

type testNilComponent struct {
	started bool
}

func (c *testNilComponent) Start() error {
	c.started = true
	return nil
}

func TestNilPointerSkipped(t *testing.T) {
	app := &testNilApp{
		Present: &testNilComponent{},
		Nil:     nil,
	}

	err := Start(app)
	assert.NoError(t, err)
	assert.True(t, app.Present.started)
}

// test top-level slice traversal
type testTopLevelSliceComponent struct {
	name    string
	started bool
	stopped bool
}

func (c *testTopLevelSliceComponent) Start() error {
	c.started = true
	return nil
}

func (c *testTopLevelSliceComponent) Stop() error {
	c.stopped = true
	return nil
}

func TestStartSlice(t *testing.T) {
	workers := []*testTopLevelSliceComponent{
		{name: "worker1"},
		{name: "worker2"},
		{name: "worker3"},
	}

	err := Start(&workers)
	assert.NoError(t, err)

	for _, w := range workers {
		assert.True(t, w.started, "worker %s should be started", w.name)
	}
}

func TestStopSlice(t *testing.T) {
	workers := []*testTopLevelSliceComponent{
		{name: "worker1"},
		{name: "worker2"},
		{name: "worker3"},
	}

	err := Stop(&workers)
	assert.NoError(t, err)

	for _, w := range workers {
		assert.True(t, w.stopped, "worker %s should be stopped", w.name)
	}
}

// test Wire with slice
type testWireSliceComponent struct {
	name  string
	wired bool
}

func (c *testWireSliceComponent) Wire(components *[]*testWireSliceComponent) error {
	c.wired = true
	return nil
}

func TestWireSlice(t *testing.T) {
	components := []*testWireSliceComponent{
		{name: "first"},
		{name: "second"},
		{name: "third"},
	}

	err := Wire(&components)
	assert.NoError(t, err)

	for _, c := range components {
		assert.True(t, c.wired, "component %s should be wired", c.name)
	}
}

// test top-level map traversal
type testTopLevelMapComponent struct {
	name    string
	started bool
	stopped bool
}

func (c *testTopLevelMapComponent) Start() error {
	c.started = true
	return nil
}

func (c *testTopLevelMapComponent) Stop() error {
	c.stopped = true
	return nil
}

func TestStartMap(t *testing.T) {
	handlers := map[string]*testTopLevelMapComponent{
		"get":    {name: "get"},
		"post":   {name: "post"},
		"delete": {name: "delete"},
	}

	err := Start(&handlers)
	assert.NoError(t, err)

	for _, h := range handlers {
		assert.True(t, h.started, "handler %s should be started", h.name)
	}
}

func TestStopMap(t *testing.T) {
	handlers := map[string]*testTopLevelMapComponent{
		"get":    {name: "get"},
		"post":   {name: "post"},
		"delete": {name: "delete"},
	}

	err := Stop(&handlers)
	assert.NoError(t, err)

	for _, h := range handlers {
		assert.True(t, h.stopped, "handler %s should be stopped", h.name)
	}
}

// test interface slice traversal
type testInterfaceComponent struct {
	name    string
	started bool
	stopped bool
}

func (c *testInterfaceComponent) Start() error {
	c.started = true
	return nil
}

func (c *testInterfaceComponent) Stop() error {
	c.stopped = true
	return nil
}

func TestStartInterfaceSlice(t *testing.T) {
	c1 := &testInterfaceComponent{name: "first"}
	c2 := &testInterfaceComponent{name: "second"}
	c3 := &testInterfaceComponent{name: "third"}

	// slice of interfaces
	components := []Startable{c1, c2, c3}

	err := Start(&components)
	assert.NoError(t, err)

	assert.True(t, c1.started, "first should be started")
	assert.True(t, c2.started, "second should be started")
	assert.True(t, c3.started, "third should be started")
}

func TestStopInterfaceSlice(t *testing.T) {
	c1 := &testInterfaceComponent{name: "first"}
	c2 := &testInterfaceComponent{name: "second"}
	c3 := &testInterfaceComponent{name: "third"}

	// slice of interfaces
	components := []Stoppable{c1, c2, c3}

	err := Stop(&components)
	assert.NoError(t, err)

	assert.True(t, c1.stopped, "first should be stopped")
	assert.True(t, c2.stopped, "second should be stopped")
	assert.True(t, c3.stopped, "third should be stopped")
}

func TestStartInterfaceMap(t *testing.T) {
	c1 := &testInterfaceComponent{name: "first"}
	c2 := &testInterfaceComponent{name: "second"}

	// map of interfaces
	components := map[string]Startable{
		"first":  c1,
		"second": c2,
	}

	err := Start(&components)
	assert.NoError(t, err)

	assert.True(t, c1.started, "first should be started")
	assert.True(t, c2.started, "second should be started")
}

// test interface field on struct
type testAppWithInterfaceField struct {
	Worker Startable
}

func TestStartStructWithInterfaceField(t *testing.T) {
	c := &testInterfaceComponent{name: "worker"}
	app := &testAppWithInterfaceField{
		Worker: c,
	}

	err := Start(app)
	assert.NoError(t, err)

	assert.True(t, c.started, "worker should be started")
}

// test struct with interface slice field
type testAppWithInterfaceSliceField struct {
	Workers []Startable
}

func TestStartStructWithInterfaceSliceField(t *testing.T) {
	c1 := &testInterfaceComponent{name: "first"}
	c2 := &testInterfaceComponent{name: "second"}

	app := &testAppWithInterfaceSliceField{
		Workers: []Startable{c1, c2},
	}

	err := Start(app)
	assert.NoError(t, err)

	assert.True(t, c1.started, "first should be started")
	assert.True(t, c2.started, "second should be started")
}

// test struct with interface map field
type testAppWithInterfaceMapField struct {
	Handlers map[string]Stoppable
}

func TestStopStructWithInterfaceMapField(t *testing.T) {
	c1 := &testInterfaceComponent{name: "get"}
	c2 := &testInterfaceComponent{name: "post"}

	app := &testAppWithInterfaceMapField{
		Handlers: map[string]Stoppable{
			"get":  c1,
			"post": c2,
		},
	}

	err := Stop(app)
	assert.NoError(t, err)

	assert.True(t, c1.stopped, "get should be stopped")
	assert.True(t, c2.stopped, "post should be stopped")
}
