package telemetry

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	segmentio "github.com/segmentio/analytics-go/v3"
	"github.com/sirupsen/logrus"

	"github.com/tensorchord/openmodelz/mdz/pkg/version"
)

type TelemetryField func(*segmentio.Properties)

type Telemetry interface {
	Record(command string, args ...TelemetryField)
}

type defaultTelemetry struct {
	client  segmentio.Client
	uid     string
	enabled bool
}

const telemetryToken = "65WHA9bxCNX74K3HjgplMOmsio9LkYSI"

var (
	once                sync.Once
	telemetry           *defaultTelemetry
	telemetryConfigFile string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	telemetryConfigFile = filepath.Join(home, ".config", "openmodelz", "telemetry")
}

func GetTelemetry() Telemetry {
	return telemetry
}

func Initialize(enabled bool) error {
	once.Do(func() {
		client, err := segmentio.NewWithConfig(telemetryToken, segmentio.Config{
			BatchSize: 1,
		})
		if err != nil {
			panic(err)
		}
		telemetry = &defaultTelemetry{
			client:  client,
			enabled: enabled,
		}
	})
	return telemetry.init()
}

func (t *defaultTelemetry) init() error {
	if !t.enabled {
		return nil
	}
	// detect if the config file already exists
	_, err := os.Stat(telemetryConfigFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to stat telemetry config file")
		}
		t.uid = uuid.New().String()
		return t.dumpConfig()
	}
	if err = t.loadConfig(); err != nil {
		return errors.Wrap(err, "failed to load telemetry config")
	}
	t.Idnetify()
	return nil
}

func (t *defaultTelemetry) dumpConfig() error {
	if err := os.MkdirAll(filepath.Dir(telemetryConfigFile), os.ModeDir|0700); err != nil {
		return errors.Wrap(err, "failed to create telemetry config directory")
	}
	file, err := os.Create(telemetryConfigFile)
	if err != nil {
		return errors.Wrap(err, "failed to create telemetry config file")
	}
	defer file.Close()
	_, err = file.WriteString(t.uid)
	if err != nil {
		return errors.Wrap(err, "failed to write telemetry config file")
	}
	return nil
}

func (t *defaultTelemetry) loadConfig() error {
	file, err := os.Open(telemetryConfigFile)
	if err != nil {
		return errors.Wrap(err, "failed to open telemetry config file")
	}
	defer file.Close()
	uid, err := io.ReadAll(file)
	if err != nil {
		return errors.Wrap(err, "failed to read telemetry config file")
	}
	t.uid = string(uid)
	return nil
}

func (t *defaultTelemetry) Idnetify() {
	if !t.enabled {
		return
	}
	v := version.GetOpenModelzVersion()
	if err := t.client.Enqueue(segmentio.Identify{
		AnonymousId: t.uid,
		Context: &segmentio.Context{
			OS: segmentio.OSInfo{
				Name:    runtime.GOOS,
				Version: runtime.GOARCH,
			},
			App: segmentio.AppInfo{
				Name:    "openmodelz",
				Version: v,
			},
		},
		Timestamp: time.Now(),
		Traits:    segmentio.NewTraits(),
	}); err != nil {
		logrus.WithError(err).Debug("failed to identify user")
		return
	}
}

func AddField(name string, value interface{}) TelemetryField {
	return func(p *segmentio.Properties) {
		p.Set(name, value)
	}
}

func (t *defaultTelemetry) Record(command string, fields ...TelemetryField) {
	if !t.enabled {
		return
	}
	logrus.WithField("UID", t.uid).WithField("command", command).Debug("send telemetry")
	track := segmentio.Track{
		AnonymousId: t.uid,
		Event:       command,
		Properties:  segmentio.NewProperties(),
	}
	for _, field := range fields {
		field(&track.Properties)
	}
	if err := t.client.Enqueue(track); err != nil {
		logrus.WithError(err).Debug("failed to send telemetry")
	}
	// make sure the msg can be sent out
	t.client.Close()
}
