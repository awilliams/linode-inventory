package main

import (
  "encoding/json"
  "github.com/awilliams/linode-inventory/api"
)

type HostMeta map[string]string

type Inventory struct {
  Meta  map[string]map[string]HostMeta `json:"_meta"`
  Hosts []string                       `json:"hosts"`
}

func (self Inventory) toJson() ([]byte, error) {
  return json.MarshalIndent(self, " ", "  ")
}

func makeInventory(linodes api.Linodes) Inventory {
  meta := make(map[string]map[string]HostMeta)
  hostvars := make(map[string]HostMeta)
  meta["hostvars"] = hostvars

  inventory := Inventory{Hosts: []string{}, Meta: meta}
  for _, linode := range linodes {
    inventory.Hosts = append(inventory.Hosts, linode.Label)
    hostmeta := make(HostMeta)
    hostmeta["ansible_ssh_host"] = linode.PublicIp()
    hostmeta["host_label"] = linode.Label
    hostmeta["host_display_group"] = linode.DisplayGroup
    hostmeta["host_private_ip"] = linode.PrivateIp()
    hostvars[linode.Label] = hostmeta
  }
  return inventory
}