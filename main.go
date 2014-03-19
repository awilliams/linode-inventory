package main

import (
  "code.google.com/p/gcfg"
  "fmt"
  "github.com/awilliams/linode-inventory/api"
  "github.com/mgutz/ansi"
  "log"
  "os"
  "sort"
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

type sortedLinodes []*api.Linode

func (self sortedLinodes) Len() int {
  return len(self)
}
func (self sortedLinodes) Swap(i, j int) {
  self[i], self[j] = self[j], self[i]
}
func (self sortedLinodes) Less(i, j int) bool {
  return self[i].Label < self[j].Label
}

func printLinodes(linodes api.Linodes) {
  grouped := make(map[string]sortedLinodes)
  for _, linode := range linodes {
    grouped[linode.DisplayGroup] = append(grouped[linode.DisplayGroup], linode)
  }
  for displayGroup, linodes := range grouped {
    sort.Sort(linodes)
    fmt.Printf("[%s]\n\n", ansi.Color(displayGroup, "green"))
    for _, linode := range linodes {
      labelColor := "magenta"
      if linode.Status != 1 {
        labelColor = "blue"
      }
      fmt.Printf(" * %-25s\tRunning=%v, Ram=%d, LinodeId=%d\n", ansi.Color(linode.Label, labelColor), linode.Status == 1, linode.Ram, linode.Id)
      linode.SortIps()
      for _, ip := range linode.Ips {
        var ipType string
        if ip.Public == 1 {
          ipType = "Public"
        } else {
          ipType = "Private"
        }
        fmt.Printf("   %-15s\t%s\n", ip.Ip, ipType)
      }
      fmt.Println("")
    }
  }
}

func main() {
  config, err := getConfig()
  if err != nil {
    log.Fatal(err)
  }

  linodes, err := api.LinodeListWithIps(config.ApiKey)
  if err != nil {
    log.Fatal(err)
  }

  // --list and --host are called from Ansible
  // see: http://docs.ansible.com/developing_inventory.html
  if len(os.Args) > 1 && os.Args[1][0:2] == "--" {
    // 1 == running
    linodes = linodes.FilterByStatus(1)
    // only apply DisplayGroup filter when using ansible feature
    if config.DisplayGroup != "" {
      linodes = linodes.FilterByDisplayGroup(config.DisplayGroup)
    }
    switch os.Args[1] {
    case "--list":
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
      fmt.Errorf("Unrecognized flag: %v\nAvailable flags: --list or --host\n", os.Args[1])
    }
  } else {
    // non-ansible case, just print linodes
    // optionally using first arg as display group filter
    if len(os.Args) > 1 {
      linodes = linodes.FilterByDisplayGroup(os.Args[1])
    }    
    printLinodes(linodes)
  }
}
