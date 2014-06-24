package main

// Provides output for use as an Ansible inventory plugin

import (
	"encoding/json"

	"github.com/awilliams/linode"
)

func newInventory(nodes map[int]*linodeWithIPs) *inventory {
	meta := make(map[string]map[string]map[string]string)
	hostvars := make(map[string]map[string]string, len(nodes))
	meta["hostvars"] = hostvars

	inv := inventory{Meta: meta, Hosts: make([]string, len(nodes))}
	i := 0
	for _, n := range nodes {
		inv.Hosts[i] = n.node.Label
		i++
		publicIP, privateIP := publicPrivateIP(n.ips)
		hostvars[n.node.Label] = map[string]string{
			"ansible_ssh_host":   publicIP,
			"host_label":         n.node.Label,
			"host_display_group": n.node.DisplayGroup,
			"host_private_ip":    privateIP,
			"host_public_ip":     publicIP,
		}
	}
	return &inv
}

type inventory struct {
	Meta  map[string]map[string]map[string]string `json:"_meta"`
	Hosts []string                                `json:"hosts"`
}

func (i *inventory) toJSON() ([]byte, error) {
	return json.MarshalIndent(i, " ", "  ")
}

func publicPrivateIP(ips []linode.LinodeIP) (string, string) {
	var pub, prv string
	for _, ip := range ips {
		if ip.IsPublic() {
			pub = ip.IP
		} else {
			prv = ip.IP
		}
		if pub != "" && prv != "" {
			break
		}
	}
	return pub, prv
}
