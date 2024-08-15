package bridge

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	httpclient "github.com/utiiz/go-hue/pkg/http_client"
)

func TestNewBridge(t *testing.T) {
	bridge := NewBridge("237.84.2.178")
	assert.Equal(t, "237.84.2.178", bridge.String())
	assert.Equal(t, "http://237.84.2.178/api", bridge.URL())
}

func TestString(t *testing.T) {
	bridge := NewBridge("237.84.2.178")
	assert.Equal(t, "237.84.2.178", bridge.String())
}

func TestURL(t *testing.T) {
	bridge := NewBridge("237.84.2.178")
	assert.Equal(t, "http://237.84.2.178/api", bridge.URL())
}

func TestDiscover(t *testing.T) {
	t.Run("Successful discovery", func(t *testing.T) {
		// Mock the http response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"internalipaddress":"237.84.2.178","id":"123","port":8080}]`))
		}))

		defer server.Close()
		discoverURL = server.URL

		bridges, err := Discover()

		expected := &[]Bridge{
			{
				Host: "237.84.2.178",
				User: nil,
			},
		}

		assert.Nil(t, err)
		assert.Equal(t, expected, bridges)
	})

	t.Run("Failed discovery", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))

		defer server.Close()
		discoverURL = server.URL

		bridges, err := Discover()

		assert.Nil(t, bridges)
		assert.NotNil(t, err)
	})
}

func TestGetUser(t *testing.T) {
	// Create a new bridge
	bridge := NewBridge("237.84.2.178")

	t.Run("Successfully retrieved a user", func(t *testing.T) {
		// Create a mock http client
		mockClient := &httpclient.MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte(`[{"success":{"username":"6glgT6ofIRwOs4SIKt32zdDfBKrRGE8JT7Min5xi"}}]`))),
				}, nil
			},
		}

		bridge.Client = mockClient

		user, err := bridge.GetUser()

		assert.Nil(t, err)
		assert.Equal(t, "6glgT6ofIRwOs4SIKt32zdDfBKrRGE8JT7Min5xi", user.Username)
	})

	t.Run("Failed to retrieve a user - Success key not found", func(t *testing.T) {
		// Create a mock http client
		mockClient := &httpclient.MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte(`[{"error":{"address":"237.84.2.178","description":"Success key not found"}}]`))),
				}, nil
			},
		}

		bridge.Client = mockClient

		user, err := bridge.GetUser()

		assert.Nil(t, user)
		assert.Equal(t, "success key not found or not a map", err.Error())
	})

	t.Run("Failed to retrieve a user - No data found", func(t *testing.T) {
		// Create a mock http client
		mockClient := &httpclient.MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte(`[]`))),
				}, nil
			},
		}

		bridge.Client = mockClient

		user, err := bridge.GetUser()

		assert.Nil(t, user)
		assert.Equal(t, "no data found in the response", err.Error())
	})

	t.Run("Failed to retrieve a user - Unmarshal error", func(t *testing.T) {
		// Create a mock http client
		mockClient := &httpclient.MockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer([]byte(`{}`))),
				}, nil
			},
		}

		bridge.Client = mockClient

		user, err := bridge.GetUser()

		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "cannot unmarshal")
	})
}
