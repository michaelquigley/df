package df

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServiceConfig struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

type testServiceDatabase struct {
	connected bool
	linked    bool
	started   bool
	stopped   bool
}

func (d *testServiceDatabase) Link(r *Registry) error {
	d.linked = true
	return nil
}

func (d *testServiceDatabase) Start() error {
	d.started = true
	return nil
}

func (d *testServiceDatabase) Stop() error {
	d.stopped = true
	return nil
}

type testWebServer struct {
	db      *testServiceDatabase
	linked  bool
	started bool
	stopped bool
}

func (w *testWebServer) Link(r *Registry) error {
	if db, found := Get[*testServiceDatabase](r); found {
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

type testServiceDatabaseFactory struct{}

func (f *testServiceDatabaseFactory) Build(s *Service[testConfig]) error {
	db := &testServiceDatabase{connected: true}
	SetAs[*testServiceDatabase](s.R, db)
	return nil
}

type testWebServerFactory struct{}

func (f *testWebServerFactory) Build(s *Service[testConfig]) error {
	server := &testWebServer{}
	SetAs[*testWebServer](s.R, server)
	return nil
}

type errorFactory struct{}

func (f *errorFactory) Build(s *Service[testConfig]) error {
	return errors.New("build failed")
}

type errorLinkable struct{}

func (e *errorLinkable) Link(r *Registry) error {
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

func TestNewService(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.Cfg)
	assert.NotNil(t, service.R)
	assert.Empty(t, service.Factories)

	// configuration should be registered in registry
	retrievedCfg, found := Get[testConfig](service.R)
	assert.True(t, found)
	assert.Equal(t, cfg, retrievedCfg)
}

func TestWithFactory(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	factory1 := &testServiceDatabaseFactory{}
	factory2 := &testWebServerFactory{}

	// test fluent api
	result := WithFactory(service, factory1)
	assert.Equal(t, service, result) // should return same service
	assert.Len(t, service.Factories, 1)
	assert.Equal(t, factory1, service.Factories[0])

	// add second factory
	WithFactory(service, factory2)
	assert.Len(t, service.Factories, 2)
	assert.Equal(t, factory1, service.Factories[0])
	assert.Equal(t, factory2, service.Factories[1])
}

func TestService_Build(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	err := service.Build()
	assert.NoError(t, err)

	// verify objects were created and registered
	db, found := Get[*testServiceDatabase](service.R)
	assert.True(t, found)
	assert.True(t, db.connected)
	assert.False(t, db.linked) // not linked yet

	server, found := Get[*testWebServer](service.R)
	assert.True(t, found)
	assert.False(t, server.linked) // not linked yet
}

func TestService_Link(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	err := service.Build()
	assert.NoError(t, err)

	err = service.Link()
	assert.NoError(t, err)

	// verify objects were linked
	db, _ := Get[*testServiceDatabase](service.R)
	assert.True(t, db.linked)

	server, _ := Get[*testWebServer](service.R)
	assert.True(t, server.linked)
	assert.Equal(t, db, server.db) // dependency injection worked
}

func TestService_Start(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	err := service.Build()
	assert.NoError(t, err)

	err = service.Link()
	assert.NoError(t, err)

	err = service.Start()
	assert.NoError(t, err)

	// verify objects were started
	db, _ := Get[*testServiceDatabase](service.R)
	assert.True(t, db.started)

	server, _ := Get[*testWebServer](service.R)
	assert.True(t, server.started)
}

func TestService_FullLifecycle(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	// full lifecycle
	err := service.Build()
	assert.NoError(t, err)

	err = service.Link()
	assert.NoError(t, err)

	err = service.Start()
	assert.NoError(t, err)

	// verify final state
	db, _ := Get[*testServiceDatabase](service.R)
	assert.True(t, db.connected)
	assert.True(t, db.linked)
	assert.True(t, db.started)

	server, _ := Get[*testWebServer](service.R)
	assert.True(t, server.linked)
	assert.True(t, server.started)
	assert.Equal(t, db, server.db)
}

func TestService_Initialize(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	// initialize should do Build + Link in one call
	err := service.Initialize()
	assert.NoError(t, err)

	// verify objects were created and linked
	db, found := Get[*testServiceDatabase](service.R)
	assert.True(t, found)
	assert.True(t, db.connected)
	assert.True(t, db.linked)

	server, found := Get[*testWebServer](service.R)
	assert.True(t, found)
	assert.True(t, server.linked)
	assert.Equal(t, db, server.db)

	// objects should not be started yet
	assert.False(t, db.started)
	assert.False(t, server.started)
}

func TestService_Stop(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &testWebServerFactory{})

	// full startup
	err := service.Initialize()
	assert.NoError(t, err)

	err = service.Start()
	assert.NoError(t, err)

	// shutdown
	err = service.Stop()
	assert.NoError(t, err)

	// verify objects were stopped in reverse order
	db, _ := Get[*testServiceDatabase](service.R)
	assert.True(t, db.stopped)

	server, _ := Get[*testWebServer](service.R)
	assert.True(t, server.stopped)
}

func TestService_StopEmpty(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	// stop empty service should not error
	err := service.Stop()
	assert.NoError(t, err)
}

func TestService_BuildError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &testServiceDatabaseFactory{})
	WithFactory(service, &errorFactory{})

	err := service.Build()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build failed")

	// first factory should have succeeded
	_, found := Get[*testServiceDatabase](service.R)
	assert.True(t, found)
}

func TestService_LinkError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	// add an object that fails to link
	errorLinkable := &errorLinkable{}
	service.R.Set(errorLinkable)

	err := service.Link()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "link failed")
}

func TestService_StartError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	// add an object that fails to start
	errorStartable := &errorStartable{}
	service.R.Set(errorStartable)

	err := service.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start failed")
}

func TestService_StopError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)

	// add objects - one that fails to stop, one that succeeds
	errorStoppable := &errorStoppable{}
	db := &testServiceDatabase{}
	service.R.Set(errorStoppable)
	SetAs[*testServiceDatabase](service.R, db)

	err := service.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stop failed")

	// but successful object should still be stopped
	assert.True(t, db.stopped)
}

func TestService_InitializeError(t *testing.T) {
	cfg := testConfig{Name: "test", Port: 8080}
	service := NewService(cfg)
	WithFactory(service, &errorFactory{})

	err := service.Initialize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "build failed")
}
