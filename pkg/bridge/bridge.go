package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	client "github.com/utiiz/go-hue/pkg/http_client"
	"github.com/utiiz/go-hue/pkg/user"
)

var (
	discoverURL = "https://discovery.meethue.com"
)

type IBridge interface {
	String() string
	URL() string
	GetUser() (*user.User, error)
	SetUser(user *user.User)
	GetLights(id string)
}

type Bridge struct {
	Host   string `json:"internalipaddress"`
	User   *user.User
	Client client.HTTPClient
}

func NewBridge(host string) *Bridge {
	return &Bridge{
		Host: host,
		User: nil,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (b *Bridge) String() string {
	return fmt.Sprintf("%s", b.Host)
}

func (b *Bridge) URL() string {
	if b.User == nil {
		return fmt.Sprintf("http://%s/api", b.Host)
	}
	return fmt.Sprintf("http://%s/api/%s", b.Host, b.User.Username)
}

func Discover() (*[]Bridge, error) {
	resp, err := http.Get(discoverURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bridges []Bridge
	err = json.Unmarshal(bodyBytes, &bridges)
	if err != nil {
		return nil, err
	}

	return &bridges, nil
}

func (b *Bridge) GetUser() (*user.User, error) {
	// URL	https://<bridge ip address>/api
	// Body	{"devicetype":"app_name#instance_name", "generateclientkey":true}
	// Method	POST

	inputData := map[string]any{
		"devicetype": "go-home#go-home",
	}
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", b.URL(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var outputData []map[string]interface{}
	err = json.Unmarshal(bodyBytes, &outputData)
	if err != nil {
		return nil, err
	}

	if len(outputData) > 0 {
		if success, ok := outputData[0]["success"].(map[string]interface{}); ok {
			username, _ := success["username"].(string)

			b.User = user.NewUser(username)
			return b.User, nil
		} else {
			return nil, fmt.Errorf("success key not found or not a map")
		}
	} else {
		return nil, fmt.Errorf("no data found in the response")
	}
}

func (b *Bridge) SetUser(user *user.User) {
	b.User = user
}

func (b *Bridge) GetLights(id string) {
	// URL	https://<bridge ip address>/api/<username>/lights
	// Body	{}
	// Method	GET
}
