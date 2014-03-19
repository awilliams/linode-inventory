package api

import (
  "strconv"
  "sort"
)

type Linode struct {
  Id           int      `json:"LINODEID"`
  Status       int      `json:"STATUS"`
  Label        string   `json:"LABEL"`
  DisplayGroup string   `json:"LPM_DISPLAYGROUP"`
  Ram          int      `json:"TOTALRAM"`
  Ips          []LinodeIp 
}

func (self Linode) PublicIp() string {
  var ip string
  for _, linodeIp := range self.Ips {
    if linodeIp.Public == 1 {
      ip = linodeIp.Ip
      break
    }
  }
  return ip
}

func (self Linode) PrivateIp() string {
  var ip string
  for _, linodeIp := range self.Ips {
    if linodeIp.Public == 0 {
      ip = linodeIp.Ip
      break
    }
  }
  return ip
}

type sortedLinodeIps []LinodeIp

func (self sortedLinodeIps) Len() int {
  return len(self)
}
func (self sortedLinodeIps) Swap(i, j int) {
  self[i], self[j] = self[j], self[i]
}
// Public first
func (self sortedLinodeIps) Less(i, j int) bool {
  return self[i].Public > self[j].Public
}

func (self Linode) SortIps() {
  sort.Sort(sortedLinodeIps(self.Ips))
}

type Linodes map[int]*Linode

func (self Linodes) FilterByDisplayGroup(group string) Linodes {
  for id, linode := range self {
    if linode.Status != 1 || (linode.DisplayGroup != "" && linode.DisplayGroup != group) {
      delete(self, id)
    }
  }
  return self
}

func (self Linodes) FilterByStatus(status int) Linodes {
  for id, linode := range self {
    if linode.Status != status {
      delete(self, id)
    }
  }
  return self
}

func LinodeList(apiKey string) (Linodes, error) {
  method := "linode.list"
  apiRequest, err := NewApiRequest(apiKey)
  if err != nil {
    return nil, err
  }
  apiRequest.AddAction(method)

  var data struct {
    Linodes []Linode `json:"DATA,omitempty"`
  }
  err = apiRequest.GetJson(&data)
  if err != nil {
    return nil, err
  }

  linodes := make(Linodes)
  for _, linode := range data.Linodes {
    l := linode
    linodes[linode.Id] = &l
  }

  return linodes, nil
}

type LinodeIp struct {
  LinodeId int    `json:"LINODEID"`
  Ip       string `json:"IPADDRESS"`
  Public   int    `json:"ISPUBLIC"`
}

// first fetch the list of linodes, 
// then use a batch request to list all the ips associated with those linodes
func LinodeListWithIps(apiKey string) (Linodes, error) {
  linodes, err := LinodeList(apiKey)
  if err != nil {
    return nil, err
  }

  method := "linode.ip.list"
  apiRequest, err := NewApiRequest(apiKey)
  if err != nil {
    return nil, err
  }
  for _, linode := range linodes {
    action := apiRequest.AddAction(method)
    action.Set("LinodeID", strconv.Itoa(linode.Id))
  }
  
  var data []struct {
    LinodeIps []LinodeIp `json:"DATA"`
  }
  err = apiRequest.GetJson(&data)
  if err != nil {
    return nil, err
  }

  for _, ipList := range data {
    for _, linodeIp := range ipList.LinodeIps {
      if linode, ok := linodes[linodeIp.LinodeId]; ok {
        linode.Ips = append(linode.Ips, linodeIp)        
      }
    }    
  }

  return linodes, nil
}
