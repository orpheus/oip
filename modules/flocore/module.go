package flocore

import (
	"context"
	"github.com/azer/logger"
	"github.com/bitspill/flod/rpcclient"
)

const ModuleId = "flocore-default"

var FlocoreModule = &module{
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

	//host := viper.GetString("flocore.host")
	//user := viper.GetString("flocore.user")
	//pass := viper.GetString("flocore.pass")
	//tls := viper.GetBool("flocore.tls")

	client, err := AddFloCore("127.0.0.1:7313", "user", "pass", true)
	if (err != nil) {
		log.Error("Unable to connect to Flocore", logger.Attrs{"host": "127.0.0.1:7313", "err": err})
		return
	}

	log.Info("Connected", logger.Attrs{"module": m.ID()})

	// set client
	m.client = client
	// set active
	m.active = true
}

// Disconnect rpc client
func (m *module) DisconnectNode() {
	// disconnect node
	m.client.Disconnect()
	// set is not active
	m.active = false
	//log.Info("Disconnected", logger.Attrs{"module": m.ID()})
}

// Initialize side effects
func (m *module) Initialize() {}

// Is the rpc connection current
func (m module) Active() bool {
	return m.active
}
