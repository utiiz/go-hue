package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/utiiz/go-hue/internal/types"
	"github.com/utiiz/go-hue/pkg/light"
	"github.com/utiiz/go-hue/pkg/user"
)

var (
	discoverURL = "https://discovery.meethue.com"
)

type Bridge struct {
	IP     string `json:"internalipaddress"`
	User   *user.User
	Client *http.Client
}

func NewBridge(ip string) *Bridge {
	return &Bridge{
		IP:   ip,
		User: nil,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (b *Bridge) String() string {
	return fmt.Sprintf("%s", b.IP)
}

func (b *Bridge) UnmarshalJSON(data []byte) error {
	var rawMap map[string]json.RawMessage
	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	var ip string
	if ipRaw, ok := rawMap["internalipaddress"]; ok {
		err = json.Unmarshal(ipRaw, &ip)
		if err != nil {
			return err
		}
	}

	*b = *NewBridge(ip)

	return nil
}

func (b *Bridge) URL() string {
	if b.User == nil {
		return fmt.Sprintf("http://%s/api", b.IP)
	}
	return fmt.Sprintf("http://%s/api/%s", b.IP, b.User.Username)
}

func (b *Bridge) GetClient() *http.Client {
	return b.Client
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

// ------------------------------------
// ------         Users          ------
// ------------------------------------

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

	resp, err := b.Client.Post(b.URL(), "application/json", bytes.NewBuffer(jsonData))
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

// ------------------------------------
// ------        Lights          ------
// ------------------------------------

func (b *Bridge) GetLights() (*[]light.Light, error) {
	// URL	https://<bridge ip address>/api/<username>/lights
	// Body	{}
	// Method	GET

	url := fmt.Sprintf("%s/lights", b.URL())

	resp, err := b.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lightsMap []light.Light
	err = json.Unmarshal(bodyBytes, &lightsMap)
	if err != nil {
		return nil, err
	}

	var lights []light.Light
	for _, light := range lightsMap {
		light.Bridge = b
		lights = append(lights, light)
	}

	return &lights, nil
}

func (b *Bridge) SetLightOn(id string) error {
	// URL	https://<bridge ip address>/api/<username>/lights/<light id>/state
	// Body	{"on":true}
	// Method	PUT

	url := fmt.Sprintf("%s/lights/%s/state", b.URL(), id)

	inputData := map[string]any{
		"on": true,
	}
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var outputData []map[string]interface{}
	err = json.Unmarshal(bodyBytes, &outputData)
	if err != nil {
		return err
	}

	if len(outputData) > 0 {
		if _, ok := outputData[0]["success"].(map[string]interface{}); ok {
			return nil
		} else {
			return fmt.Errorf("success key not found or not a map")
		}
	} else {
		return fmt.Errorf("no data found in the response")
	}
}

// Ensure Bridge implements Bridge interface
var _ types.Bridge = (*Bridge)(nil)
