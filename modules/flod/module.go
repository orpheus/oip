package flod

import (
	"context"
	"github.com/azer/logger"
	"github.com/bitspill/flod/rpcclient"
	"github.com/spf13/viper"
	"time"
)

const ModuleId = "flod-default"

var FlodModule = &module{
	id: ModuleId,
}

type module struct {
	id     string
	client *rpcclient.Client
	active bool
	config interface{}
	ctx    context.Context
}

// Get the Module ID
func (m *module) ID() string {
	return m.id
}

// Wait for node be found & add rpc client
func (m *module) ConnectToNode(ctx context.Context) {
	// add context to node
	m.ctx = ctx
	log.Info("Added Context", logger.Attrs{"module-context": m.ctx})

	// create timeout context
	tenMinuteCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// get flod rpc config
	host := viper.GetString("flod.host")
	user := viper.GetString("flod.user")
	pass := viper.GetString("flod.pass")
	tls := viper.GetBool("flod.tls")

	// wait to connect to flod and add rpc client
	client, err := WaitForFlod(tenMinuteCtx, host, user, pass, tls)
	if err != nil {
		log.Error("Unable to connect to Flod", logger.Attrs{"host": host, "err": err})
		log.Error("Shutting down...")
		return
	}

	log.Info("Connected Module", logger.Attrs{"module": m.ID()})

	m.client = client

	// set is active
	m.active = true
}

// Disconnect rpc client
func (m *module) DisconnectNode() {
	// disconnect node
	m.client.Disconnect()
	// set is not active
	m.active = false
	log.Info("Disconnected", logger.Attrs{"module": m.ID()})
}

// Initialize side effects
func (m *module) Initialize() {}

// Is the rpc connection current
func (m module) Active() bool {
	return m.active
}

func (m *module) SetActive(active bool) {
	m.active = active
}