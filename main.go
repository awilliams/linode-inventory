package main

import (
  "code.google.com/p/gcfg"
  "fmt"
  "github.com/awilliams/linode-inventory/api"
  "log"
  "os"
  "path/filepath"
)

const CONFIG_PATH = "linode-inventory.ini"

type Configuration struct {
  ApiKey       string `gcfg:"api-key"`
  DisplayGroup string `gcfg:"display-group"`
}

func getConfig() (*Configuration, error) {
  // first check directory where the executable is located
  dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
  if err != nil {
    return nil, err
  }
  path := dir + "/" + CONFIG_PATH
  if _, err := os.Stat(path); os.IsNotExist(err) {
    // fallback to PWD. This is usefull when using `go run`
    path = CONFIG_PATH
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

func getLinodes(config *Configuration) (api.Linodes, error) {
  linodes, err := api.LinodeListWithIps(config.ApiKey)
  if err != nil {
    return nil, err
  }
  // 1 == running
  linodes = linodes.FilterByStatus(1)
  // only apply DisplayGroup filter when using ansible feature
  if config.DisplayGroup != "" {
    linodes = linodes.FilterByDisplayGroup(config.DisplayGroup)
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
      inventoryJson, err := inventory.toJson()
      if err != nil {
        log.Fatal(err)
      }
      os.Stdout.Write(inventoryJson)
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
