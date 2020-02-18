package plugin

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
)

// The TF_DISABLE_PLUGIN_TLS environment variable is intended only for use by
// the plugin SDK test framework. We do not recommend Terraform CLI end-users
// set this variable.
var enableAutoMTLS = os.Getenv("TF_DISABLE_PLUGIN_TLS") == ""

// ClientConfig returns a configuration object that can be used to instantiate
// a client for the plugin described by the given metadata.
func ClientConfig(m discovery.PluginMeta) *plugin.ClientConfig {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Trace,
		Output: os.Stderr,
	})

	var cmdPath string
	var cmdArgs []string
	clientTimeout := 60 * time.Second

	// Plugin debugging using DELVE
	// Env variables available to configure the debugger:
	// - TF_PLUGIN_DELVE file path to the delve executable
	// - TF_PLUGIN_DELVE_LISTEN_ADDR address delve will be listening (default: "")
	// - TF_PLUGIN_DELVE_LISTEN_PORT port delve will be listening (default: 2345)
	// - TF_PLUGIN_DELVE_API_VERSION api version to use for remote debugging (default: 2)
	tfPluginDelve := os.Getenv("TF_PLUGIN_DELVE")
	if len(tfPluginDelve) > 0 {
		info, err := os.Stat(tfPluginDelve)

		if os.IsNotExist(err) {
			log.Fatalf("TF_PLUGIN_DELVE=%s does not exist", tfPluginDelve)
		} else if info.Mode()&0111 == 0 {
			log.Fatalf("TF_PLUGIN_DELVE=%s is not executable", tfPluginDelve)
		} else {
			pluginDelveAddr := "127.0.0.1"
			pluginDelvePort := 2345
			pluginDelveAPIVersion := 2
			// pluginDelveLogDest := "/dev/console"

			envPluginDelveAddr := os.Getenv("TF_PLUGIN_DELVE_LISTEN_ADDR")
			envPluginDelvePort := os.Getenv("TF_PLUGIN_DELVE_LISTEN_PORT")
			envPluginDelveAPIVersion := os.Getenv("TF_PLUGIN_DELVE_API_VERSION")
			// envPluginDelveLogDest := os.Getenv("TF_PLUGIN_DELVE_LOG_DEST")

			if len(envPluginDelveAddr) > 0 {
				pluginDelveAddr = envPluginDelveAddr
			}

			if len(envPluginDelvePort) > 0 {
				port, err := strconv.Atoi(envPluginDelvePort)
				if err == nil {
					pluginDelvePort = port
				} else {
					log.Fatalf("TF_PLUGIN_DELVE_LISTEN_PORT=%v is invalid", envPluginDelvePort)
				}
			}

			if len(envPluginDelveAPIVersion) > 0 {
				apiVersion, err := strconv.Atoi(envPluginDelveAPIVersion)
				if err == nil && apiVersion >= 1 && apiVersion <= 2 {
					pluginDelveAPIVersion = apiVersion
				} else {
					log.Fatalf("TF_PLUGIN_DELVE_API_VERSION=%v is invalid (must be 1 or 2)", envPluginDelveAPIVersion)
				}
			}

			// if len(envPluginDelveLogDest) > 0 {
			// 	pluginDelveLogDest = envPluginDelveLogDest
			// }

			clientTimeout = 10 * time.Minute
			cmdPath = tfPluginDelve
			cmdArgs = []string{
				"exec",
				"--headless",
				"--listen=" + fmt.Sprintf("%s:%d", pluginDelveAddr, pluginDelvePort),
				"--api-version=" + fmt.Sprintf("%d", pluginDelveAPIVersion),
				"--",
				m.Path,
			}
		}
	} else {
		cmdPath = m.Path
	}

	return &plugin.ClientConfig{
		Cmd:              exec.Command(cmdPath, cmdArgs...),
		HandshakeConfig:  Handshake,
		VersionedPlugins: VersionedPlugins,
		Managed:          true,
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         enableAutoMTLS,
		StartTimeout:     clientTimeout,
	}
}

// Client returns a plugin client for the plugin described by the given metadata.
func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
