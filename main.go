package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"code.google.com/p/gcfg"
	"github.com/awilliams/linode-inventory/api"
)

const ConfigPath = "linode-inventory.ini"

type Configuration struct {
	APIKey       string `gcfg:"api-key"`
	DisplayGroup string `gcfg:"display-group"`
}

func getConfig() (*Configuration, error) {
	// first check directory where the executable is located
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	path := dir + "/" + ConfigPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// fallback to PWD. This is usefull when using `go run`
		path = ConfigPath
	}

	var config struct {
		Linode Configuration
	}

	err = gcfg.ReadFileInto(&config, path)
	if err != nil {
		return nil, err
	}

	return &config.Linode, nil
}

func getLinodes(config *Configuration) ([]*api.Linode, error) {
	nodeMap, err := api.FetchLinodesWithIps(config.APIKey)
	if err != nil {
		return nil, err
	}

	var linodes []*api.Linode
	// function to add nodes to the slice
	addNodes := func(nodes []*api.Linode) {
		for _, node := range nodes {
			// Status 1 == running
			if node.Status == 1 {
				linodes = append(linodes, node)
			}
		}
	}

	if config.DisplayGroup != "" {
		nodes, ok := nodeMap[config.DisplayGroup]
		if !ok {
			return nil, fmt.Errorf("display group '%s' not found", config.DisplayGroup)
		}
		addNodes(nodes)
	} else {
		for _, nodes := range nodeMap {
			addNodes(nodes)
		}
	}

	return linodes, nil
}

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	// --list and --host are called from Ansible
	// see: http://docs.ansible.com/developing_inventory.html
	if len(os.Args) > 1 && os.Args[1][0:2] == "--" {
		switch os.Args[1] {
		case "--list":
			linodes, err := getLinodes(config)
			if err != nil {
				log.Fatal(err)
			}

			inventory := makeInventory(linodes)
			inventoryJSON, err := inventory.toJSON()
			if err != nil {
				log.Fatal(err)
			}
			os.Stdout.Write(inventoryJSON)
		case "--host":
			// empty hash
			fmt.Fprint(os.Stdout, "{}")
		default:
			fmt.Fprintf(os.Stderr, "Unrecognized flag: %v\nUsage: linode-inventory --list or --host\n", os.Args[1])
		}
	} else {
		fmt.Fprint(os.Stderr, "Usage: linode-inventory --list or --host\n")
	}
}
