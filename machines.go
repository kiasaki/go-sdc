package sdc

import (
	"time"
)

// Machine represents an SDC instance.
type Machine struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	State   string `json:"state"`
	Dataset string `json:"dataset"`

	Memory int `json:"memory"`
	Disk   int `json:"disk"`

	Ips      []string          `json:"ips"`
	Metadata map[string]string `json:"metadata"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`

	Package     string `json:"package"`
	Image       string `json:"image"`
	Credentials bool   `json:"credentials"`
}

func (c *Client) ListMachines() ([]*Machine, error) {
	response := []*Machine{}
	_, err := c.Get("/my/machines", &response)
	return response, err
}
