package registry

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/docker/distribution/configuration"
	_ "github.com/docker/distribution/registry/storage/driver/inmemory"
	"github.com/stretchr/testify/require"
)

// Tests to ensure nextProtos returns the correct protocols when:
// * config.HTTP.HTTP2.Disabled is not explicitly set => [h2 http/1.1]
// * config.HTTP.HTTP2.Disabled is explicitly set to false [h2 http/1.1]
// * config.HTTP.HTTP2.Disabled is explicitly set to true [http/1.1]
func TestNextProtos(t *testing.T) {
	config := &configuration.Configuration{}
	protos := nextProtos(config)
	if !reflect.DeepEqual(protos, []string{"h2", "http/1.1"}) {
		t.Fatalf("expected protos to equal [h2 http/1.1], got %s", protos)
	}
	config.HTTP.HTTP2.Disabled = false
	protos = nextProtos(config)
	if !reflect.DeepEqual(protos, []string{"h2", "http/1.1"}) {
		t.Fatalf("expected protos to equal [h2 http/1.1], got %s", protos)
	}
	config.HTTP.HTTP2.Disabled = true
	protos = nextProtos(config)
	if !reflect.DeepEqual(protos, []string{"http/1.1"}) {
		t.Fatalf("expected protos to equal [http/1.1], got %s", protos)
	}
}

func setupRegistry() (*Registry, error) {
	config := &configuration.Configuration{}
	configuration.ApplyDefaults(config)
	// probe free port where the server can listen
	ln, err := net.Listen("tcp", ":")
	if err != nil {
		return nil, err
	}
	defer ln.Close()
	config.HTTP.Addr = ln.Addr().String()
	config.HTTP.DrainTimeout = time.Duration(10) * time.Second
	config.Storage = map[string]configuration.Parameters{"inmemory": map[string]interface{}{}}
	return NewRegistry(context.Background(), config)
}

func TestGracefulShutdown(t *testing.T) {
	registry, err := setupRegistry()
	if err != nil {
		t.Fatal(err)
	}

	// run registry server
	var errchan chan error
	go func() {
		errchan <- registry.ListenAndServe()
	}()
	select {
	case err = <-errchan:
		t.Fatalf("Error listening: %v", err)
	default:
	}

	// Wait for some unknown random time for server to start listening
	time.Sleep(3 * time.Second)

	// send incomplete request
	conn, err := net.Dial("tcp", registry.config.HTTP.Addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(conn, "GET /v2/ ")

	// send stop signal
	quit <- os.Interrupt
	time.Sleep(100 * time.Millisecond)

	// try connecting again. it shouldn't
	_, err = net.Dial("tcp", registry.config.HTTP.Addr)
	if err == nil {
		t.Fatal("Managed to connect after stopping.")
	}

	// make sure earlier request is not disconnected and response can be received
	fmt.Fprintf(conn, "HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "200 OK" {
		t.Error("response status is not 200 OK: ", resp.Status)
	}
	if body, err := ioutil.ReadAll(resp.Body); err != nil || string(body) != "{}" {
		t.Error("Body is not {}; ", string(body))
	}
}

func requireEnvNotSet(t *testing.T, names ...string) {
	t.Helper()

	for _, name := range names {
		_, ok := os.LookupEnv(name)
		require.False(t, ok)
	}
}

func requireEnvSet(t *testing.T, name, value string) {
	t.Helper()

	require.Equal(t, value, os.Getenv(name))
}

func TestConfigureStackDriver_Disabled(t *testing.T) {
	config := &configuration.Configuration{}

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "GITLAB_CONTINUOUS_PROFILING")
	require.NoError(t, configureStackdriver(config))
	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "GITLAB_CONTINUOUS_PROFILING")
}

func TestConfigureStackDriver_Enabled(t *testing.T) {
	config := &configuration.Configuration{
		Monitoring: configuration.Monitoring{
			Stackdriver: configuration.StackdriverProfiler{
				Enabled: true,
			},
		},
	}

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "GITLAB_CONTINUOUS_PROFILING")
	require.NoError(t, configureStackdriver(config))
	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS")
	requireEnvSet(t, "GITLAB_CONTINUOUS_PROFILING", "stackdriver")
	require.NoError(t, os.Unsetenv("GITLAB_CONTINUOUS_PROFILING"))
}

func TestConfigureStackDriver_WithParams(t *testing.T) {
	config := &configuration.Configuration{
		Monitoring: configuration.Monitoring{
			Stackdriver: configuration.StackdriverProfiler{
				Enabled:        true,
				Service:        "registry",
				ServiceVersion: "2.9.1",
				ProjectID:      "internal",
			},
		},
	}

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "GITLAB_CONTINUOUS_PROFILING")
	require.NoError(t, configureStackdriver(config))
	defer os.Unsetenv("GITLAB_CONTINUOUS_PROFILING")

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS")
	requireEnvSet(t, "GITLAB_CONTINUOUS_PROFILING", "stackdriver?project_id=internal&service=registry&service_version=2.9.1")

}

func TestConfigureStackDriver_WithKeyFile(t *testing.T) {
	config := &configuration.Configuration{
		Monitoring: configuration.Monitoring{
			Stackdriver: configuration.StackdriverProfiler{
				Enabled: true,
				KeyFile: "/path/to/credentials.json",
			},
		},
	}

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "GITLAB_CONTINUOUS_PROFILING")
	require.NoError(t, configureStackdriver(config))
	defer os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	defer os.Unsetenv("GITLAB_CONTINUOUS_PROFILING")

	requireEnvSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "/path/to/credentials.json")
	requireEnvSet(t, "GITLAB_CONTINUOUS_PROFILING", "stackdriver")

}

func TestConfigureStackDriver_DoesNotOverrideGoogleApplicationCredentialsEnvVar(t *testing.T) {
	require.NoError(t, os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/foo/bar.json"))

	config := &configuration.Configuration{
		Monitoring: configuration.Monitoring{
			Stackdriver: configuration.StackdriverProfiler{
				Enabled: true,
				KeyFile: "/path/to/credentials.json",
			},
		},
	}

	requireEnvNotSet(t, "GITLAB_CONTINUOUS_PROFILING")
	require.NoError(t, configureStackdriver(config))
	defer os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	defer os.Unsetenv("GITLAB_CONTINUOUS_PROFILING")

	requireEnvSet(t, "GOOGLE_APPLICATION_CREDENTIALS", "/foo/bar.json")
	requireEnvSet(t, "GITLAB_CONTINUOUS_PROFILING", "stackdriver")
}

func TestConfigureStackDriver_DoesNotOverrideGitlabContinuousProfilingEnvVar(t *testing.T) {
	value := "stackdriver?project_id=foo&service=bar&service_version=1"
	require.NoError(t, os.Setenv("GITLAB_CONTINUOUS_PROFILING", value))

	config := &configuration.Configuration{
		Monitoring: configuration.Monitoring{
			Stackdriver: configuration.StackdriverProfiler{
				Enabled:        true,
				Service:        "registry",
				ServiceVersion: "2.9.1",
				ProjectID:      "internal",
			},
		},
	}

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS")
	require.NoError(t, configureStackdriver(config))
	defer os.Unsetenv("GITLAB_CONTINUOUS_PROFILING")

	requireEnvNotSet(t, "GOOGLE_APPLICATION_CREDENTIALS")
	requireEnvSet(t, "GITLAB_CONTINUOUS_PROFILING", value)
}
