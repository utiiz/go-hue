package bridge

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utiiz/go-hue/pkg/user"
)

func TestNewBridge(t *testing.T) {
	host := "192.168.1.100"
	bridge := NewBridge(host)

	assert.Equal(t, host, bridge.Host, "Bridge host should match")
	assert.Nil(t, bridge.User, "User should be nil")
	assert.NotNil(t, bridge.Client, "Client should not be nil")
	assert.Equal(t, 5*time.Second, bridge.Client.Timeout, "Client timeout should be 5 seconds")
}

func TestBridgeString(t *testing.T) {
	bridge := NewBridge("192.168.1.100")
	assert.Equal(t, "192.168.1.100", bridge.String(), "Bridge String() should return the host")
}

func TestBridgeURL(t *testing.T) {
	bridge := NewBridge("192.168.1.100")
	assert.Equal(t, "http://192.168.1.100/api", bridge.URL(), "Bridge URL without user should be correct")

	bridge.User = user.NewUser("testuser")
	assert.Equal(t, "http://192.168.1.100/api/testuser", bridge.URL(), "Bridge URL with user should be correct")
}

func TestDiscover(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bridges := []Bridge{
			{Host: "192.168.1.100"},
			{Host: "192.168.1.101"},
		}
		json.NewEncoder(w).Encode(bridges)
	}))
	defer server.Close()

	originalDiscoverURL := discoverURL
	discoverURL = server.URL
	defer func() { discoverURL = originalDiscoverURL }()

	bridges, err := Discover()

	assert.NoError(t, err, "Discover() should not return an error")
	assert.NotNil(t, bridges, "Bridges should not be nil")
	assert.Len(t, *bridges, 2, "Should discover 2 bridges")

	expectedHosts := []string{"192.168.1.100", "192.168.1.101"}
	for i, bridge := range *bridges {
		assert.Equal(t, expectedHosts[i], bridge.Host, "Bridge host should match")
	}
}

func TestGetUser(t *testing.T) {
	bridge := NewBridge("192.168.1.100")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []map[string]interface{}{
			{
				"success": map[string]interface{}{
					"username": "testuser",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	bridge.Client = server.Client()
	bridge.Host = server.URL[7:] // Remove "http://" prefix

	user, err := bridge.GetUser()

	assert.NoError(t, err, "GetUser() should not return an error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "testuser", user.Username, "Username should match")
	assert.Equal(t, user, bridge.User, "Bridge User should be set correctly")
}

func TestSetUser(t *testing.T) {
	bridge := NewBridge("192.168.1.100")
	user := user.NewUser("testuser")

	bridge.SetUser(user)

	assert.Equal(t, user, bridge.User, "SetUser() should set the User correctly")
}
