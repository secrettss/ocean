package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os/exec"
	"strings"
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/require"
)

var _ = suite("registry/get", func(t *testing.T, when spec.G, it spec.S) {
	var (
		expect         *require.Assertions
		server         *httptest.Server
		expectedRegion string
	)

	it.Before(func() {
		expect = require.New(t)

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Add("content-type", "application/json")

			switch req.URL.Path {
			case "/v2/registry":
				auth := req.Header.Get("Authorization")
				if auth != "Bearer some-magic-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if req.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}

				w.Write([]byte(registryGetResponse))
			default:
				dump, err := httputil.DumpRequest(req, true)
				if err != nil {
					t.Fatal("failed to dump request")
				}

				t.Fatalf("received unknown request: %s", dump)
			}
		}))
	})

	it("returns my account's registry", func() {
		cmd := exec.Command(builtBinaryPath,
			"-t", "some-magic-token",
			"-u", server.URL,
			"registry",
			"get",
		)
		expectedRegion = "r1"

		output, err := cmd.CombinedOutput()
		expect.NoError(err)

		expect.Equal(strings.TrimSpace(fmt.Sprintf(registryGetOutput, expectedRegion)), strings.TrimSpace(string(output)))
	})
})

const (
	registryGetResponse = `
{
	"registry": {
		"name": "my-registry",
		"region": "r1"
	}
}`
	// note: used by tests in registry_create_test.go as well
	registryGetOutput = `
Name           Endpoint                                 Region slug
my-registry    registry.digitalocean.com/my-registry    %s
`
)
