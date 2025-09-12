package da

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name     string `df:"app_name"`
	Port     int
	Secret   string `df:"api_key,+secret"`
	Timeout  time.Duration
	Enabled  bool
	Database *testDB
	Services []testService
}

type testDB struct {
	Host     string
	Username string
	Password string `df:"+secret"`
	Port     int
}

type testService struct {
	Name string
	URL  string `df:"url"`
}

type testApplicationDatabase struct {
	connected bool
	linked    bool
	started   bool
	stopped   bool
}

func (d *testApplicationDatabase) Link(c *Container) error {
	d.linked = true
	return nil
}

func (d *testApplicationDatabase) Start() error {
	d.started = true
	return nil
}

func (d *testApplicationDatabase) Stop() error {
	d.stopped = true
	return nil
}

type testWebServer struct {
	db      *testApplicationDatabase
	linked  bool
	started bool
	stopped bool
}

func (w *testWebServer) Link(c *Container) error {
	if db, found := Get[*testApplicationDatabase](c); found {
		w.db = db
		w.linked = true
		return nil
	}
	return errors.New("database not found")
}

func (w *testWebServer) Start() error {
	w.started = true
	return nil
}

func (w *testWebServer) Stop() error {
	w.stopped = true
	return nil
}

type testApplicationDatabaseFactory struct{}

func (f *testApplicationDatabaseFactory) Build(a *Application[testConfig]) error {
	db := &testApplicationDatabase{connected: true}
	SetAs[*testApplicationDatabase](a.C, db)
	return nil
}

type testWebServerFactory struct{}

func (f *testWebServerFactory) Build(a *Application[testConfig]) error {
	server := &testWebServer{}
	SetAs[*testWebServer](a.C, server)
	return nil
}

type errorFactory struct{}

func (f *errorFactory) Build(a *Application[testConfig]) error {
	return errors.New("build failed")
}

type errorLinkable struct{}

func (e *errorLinkable) Link(c *Container) error {
	return errors.New("link failed")
}

type errorStartable struct{}

func (e *errorStartable) Start() error {
	return errors.New("start failed")
}

type errorStoppable struct{}

func (e *errorStoppable) Stop() error {
	return errors.New("stop failed")
}

func TestNewApplication(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	assert.NotNil(t, app)
	assert.Equal(t, cfg, app.Cfg)
	assert.NotNil(t, app.C)
	assert.Empty(t, app.Factories)

	// configuration should be registered in container
	retrievedCfg, found := Get[testConfig](app.C)
	assert.True(t, found)
	assert.Equal(t, cfg, retrievedCfg)
}

func TestWithFactory(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	factory1 := &testApplicationDatabaseFactory{}
	factory2 := &testWebServerFactory{}

	// test fluent api
	result := WithFactory(app, factory1)
	assert.Equal(t, app, result) // should return same app
	assert.Len(t, app.Factories, 1)
	assert.Equal(t, factory1, app.Factories[0])

	// add second factory
	WithFactory(app, factory2)
	assert.Len(t, app.Factories, 2)
	assert.Equal(t, factory1, app.Factories[0])
	assert.Equal(t, factory2, app.Factories[1])
}

func TestApplication_Build(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	err := app.Build()
	assert.NoError(t, err)

	// verify objects were created and registered
	db, found := Get[*testApplicationDatabase](app.C)
	assert.True(t, found)
	assert.True(t, db.connected)
	assert.False(t, db.linked) // not linked yet

	server, found := Get[*testWebServer](app.C)
	assert.True(t, found)
	assert.False(t, server.linked) // not linked yet
}

func TestApplication_Link(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	err := app.Build()
	assert.NoError(t, err)

	err = app.Link()
	assert.NoError(t, err)

	// verify objects were linked
	db, _ := Get[*testApplicationDatabase](app.C)
	assert.True(t, db.linked)

	server, _ := Get[*testWebServer](app.C)
	assert.True(t, server.linked)
	assert.Equal(t, db, server.db) // dependency injection worked
}

func TestApplication_Start(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	err := app.Build()
	assert.NoError(t, err)

	err = app.Link()
	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	// verify objects were started
	db, _ := Get[*testApplicationDatabase](app.C)
	assert.True(t, db.started)

	server, _ := Get[*testWebServer](app.C)
	assert.True(t, server.started)
}

func TestApplication_FullLifecycle(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	// full lifecycle
	err := app.Build()
	assert.NoError(t, err)

	err = app.Link()
	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	// verify final state
	db, _ := Get[*testApplicationDatabase](app.C)
	assert.True(t, db.connected)
	assert.True(t, db.linked)
	assert.True(t, db.started)

	server, _ := Get[*testWebServer](app.C)
	assert.True(t, server.linked)
	assert.True(t, server.started)
	assert.Equal(t, db, server.db)
}

func TestApplication_Initialize(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	// initialize should do Build + Link in one call
	err := app.Initialize()
	assert.NoError(t, err)

	// verify objects were created and linked
	db, found := Get[*testApplicationDatabase](app.C)
	assert.True(t, found)
	assert.True(t, db.connected)
	assert.True(t, db.linked)

	server, found := Get[*testWebServer](app.C)
	assert.True(t, found)
	assert.True(t, server.linked)
	assert.Equal(t, db, server.db)

	// objects should not be started yet
	assert.False(t, db.started)
	assert.False(t, server.started)
}

func TestApplication_Stop(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &testWebServerFactory{})

	// full startup
	err := app.Initialize()
	assert.NoError(t, err)

	err = app.Start()
	assert.NoError(t, err)

	// shutdown
	err = app.Stop()
	assert.NoError(t, err)

	// verify objects were stopped in reverse order
	db, _ := Get[*testApplicationDatabase](app.C)
	assert.True(t, db.stopped)

	server, _ := Get[*testWebServer](app.C)
	assert.True(t, server.stopped)
}

func TestApplication_StopEmpty(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// stop empty app should not error
	err := app.Stop()
	assert.NoError(t, err)
}

func TestApplication_BuildError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &testApplicationDatabaseFactory{})
	WithFactory(app, &errorFactory{})

	err := app.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build failed")

	// first factory should have succeeded
	_, found := Get[*testApplicationDatabase](app.C)
	assert.True(t, found)
}

func TestApplication_LinkError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// add an object that fails to link
	errorLinkable := &errorLinkable{}
	Set(app.C, errorLinkable)

	err := app.Link()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "link failed")
}

func TestApplication_StartError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// add an object that fails to start
	errorStartable := &errorStartable{}
	Set(app.C, errorStartable)

	err := app.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start failed")
}

func TestApplication_StopError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// add objects - one that fails to stop, one that succeeds
	errorStoppable := &errorStoppable{}
	db := &testApplicationDatabase{}
	Set(app.C, errorStoppable)
	SetAs[*testApplicationDatabase](app.C, db)

	err := app.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stop failed")

	// but successful object should still be stopped
	assert.True(t, db.stopped)
}

func TestApplication_InitializeError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)
	WithFactory(app, &errorFactory{})

	err := app.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build failed")
}

func TestWithFactoryFunc(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// test with function factory
	result := WithFactoryFunc(app, func(a *Application[testConfig]) error {
		db := &testApplicationDatabase{connected: true}
		SetAs[*testApplicationDatabase](a.C, db)
		return nil
	})

	assert.Equal(t, app, result) // should return same app
	assert.Len(t, app.Factories, 1)

	// build should work
	err := app.Build()
	assert.NoError(t, err)

	// verify object was created and registered
	db, found := Get[*testApplicationDatabase](app.C)
	assert.True(t, found)
	assert.True(t, db.connected)
}

func TestFactoryFunc_Build(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// create function factory directly
	factoryFunc := FactoryFunc[testConfig](func(a *Application[testConfig]) error {
		server := &testWebServer{}
		SetAs[*testWebServer](a.C, server)
		return nil
	})

	WithFactory(app, factoryFunc)

	err := app.Build()
	assert.NoError(t, err)

	// verify object was created
	server, found := Get[*testWebServer](app.C)
	assert.True(t, found)
	assert.NotNil(t, server)
}

func TestFactoryFunc_Error(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	app := NewApplication(cfg)

	// test error handling in function factory
	WithFactoryFunc(app, func(a *Application[testConfig]) error {
		return errors.New("function factory failed")
	})

	err := app.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "function factory failed")
}
