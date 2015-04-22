package sdc

import (
	"time"
)

// Machine represents an SDC instance.
type Machine struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	// Can either be "virtual" or "smart"
	Type    string `json:"type"`
	State   string `json:"state"`
	Dataset string `json:"dataset"`

	Memory int `json:"memory"`
	Disk   int `json:"disk"`

	Ips      []string          `json:"ips"`
	Metadata map[string]string `json:"metadata"`

	// Formated in ISO 8601
	Created time.Time `json:"created"`
	// Formated in ISO 8601
	Updated time.Time `json:"updated"`

	Package     string `json:"package"`
	Image       string `json:"image"`
	Credentials bool   `json:"credentials"`
}

// ListMachines fetches all your machines from the SDC API.
func (c *Client) ListMachines() ([]*Machine, error) {
	response := []*Machine{}
	_, err := c.Get("/my/machines", &response)
	return response, err
}

// GetMachine fetches a specific machine from the SDC api.
func (c *Client) GetMachine(id string) (*Machine, error) {
	response := *Machine{}
	_, err := c.Get("/my/machines/"+id, &response)
	return response, err
}

// CreateMachineRequest represents the parameter needed for calling CreateMachine.
type CreateMachineRequest struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Package string `json:"package"`

	Networks        []string `json:"networks"`
	DefaultNetworks []string `json:"default_networks"`

	Metadata map[string]string `json:"metadata"`
	Tags     map[string]string `json:"tags"`

	FirewallEnabled bool `json:"firewall_enabled"`
}

// CreateMachine creates a new machine in SDC. The only two required parameters
// are `image` and `package`.
//
// The returned machine will be incomplete and you will need to
// poll the GetMachine method utils the `state` is equal to "running"
// for logging in and getting ips.
func (c *Client) CreateMachine(request CreateMachineRequest) (*Machine, error) {
	response := *Machine{}
	_, err := c.Post("/my/machines", &request, &response)
	return response, err
}
