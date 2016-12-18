package mqtt_test

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/mqhub.go/mqhub"
	_ "github.com/robotalks/mqhub.go/mqtt"
)

type EnvBuilder interface {
	Setup() error
	TearDown() error
	NewConnector(id string) (mqhub.Connector, error)
}

type RemoteEnvBuilder struct {
	serverURL string
}

func (b *RemoteEnvBuilder) Setup() error {
	b.serverURL = os.Getenv("MQTT_URL")
	if b.serverURL == "" {
		b.serverURL = "tcp://iot.eclipse.org:1883"
	}
	return nil
}

func (b *RemoteEnvBuilder) TearDown() error {
	return nil
}

func (b *RemoteEnvBuilder) NewConnector(id string) (mqhub.Connector, error) {
	return mqhub.NewConnector("mqtt+" + b.serverURL + "?client-id=" + url.QueryEscape(id))
}

var (
	TestEnv EnvBuilder
)

func buildEnv() error {
	envBuilders := map[string]EnvBuilder{
		"remote": &RemoteEnvBuilder{},
	}
	envName := os.Getenv("TEST_ENV")
	if envName == "" {
		envName = "remote"
	}

	if TestEnv = envBuilders[envName]; TestEnv == nil {
		return fmt.Errorf("invalid test env: %s", envName)
	}
	return TestEnv.Setup()
}

func TestMain(m *testing.M) {
	flag.Parse()
	if err := buildEnv(); err != nil {
		log.Fatalln(err)
	}
	paho.ERROR = log.New(os.Stderr, "paho:ERR ", log.LstdFlags)
	paho.CRITICAL = log.New(os.Stderr, "paho:CRI ", log.LstdFlags)
	paho.WARN = log.New(os.Stderr, "paho:WRN ", log.LstdFlags)
	paho.DEBUG = log.New(os.Stderr, "paho:DBG ", log.LstdFlags)
	code := m.Run()
	TestEnv.TearDown()
	os.Exit(code)
}
