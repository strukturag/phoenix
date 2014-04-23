package phoenix_test

import (
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"testing"

	phoenix "."
)

var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
})

func simpleTLSRunFunc(runtime phoenix.Runtime) error {
	runtime.DefaultHTTPSHandler(okHandler)
	return runtime.Start()
}

func Test_Server_DefaultTLSServer_IsConnectableAfterStartup(t *testing.T) {
	// NOTE(lcooper): simply making requests to localhost is far easier
	// than generating a cert with an IP SAN. If this changes, feel free
	// to use 127.0.0.1 instead.
	baseURL := "localhost:62001"
	server := phoenix.NewServer("integration-test", "unreleased").
		OverrideOption("https", "listen", baseURL).
		OverrideOption("https", "certificate", "testdata/server.crt").
		OverrideOption("https", "key", "testdata/key.pem")
	withTestServer(t, server, simpleTLSRunFunc, func() {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		resp, err := client.Get("https://"+baseURL+"/")
		if err != nil {
			t.Errorf("unexpected error making HTTP request: %v", err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("unexpected error reading response body: %v", err)
		}

		if string(data) != "ok" {
			t.Errorf("Expected data to be 'ok', but was '%s'", data)
		}
	})
}
