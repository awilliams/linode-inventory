package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func FetchLinodesWithIps(apiKey string) (Linodes, error) {
	api, err := NewApiRequest(apiKey)
	if err != nil {
		return nil, err
	}

	linodes, err := FetchLinodeList(*api)
	if err != nil {
		return nil, err
	}

	linodeIps, err := FetchLinodeIpList(*api, linodes.Ids())
	if err != nil {
		return nil, err
	}

	// associate ips with linodes
	for _, linodeDisplayGroup := range linodes {
		for _, linode := range linodeDisplayGroup {
			if ips, ok := linodeIps[linode.Id]; ok {
				sortLinodeIps(ips)
				linode.Ips = ips
			}
		}
		sortLinodes(linodeDisplayGroup)
	}

	return linodes, nil
}

// map of Linodes by their display group
type Linodes map[string][]*Linode

func (nodes Linodes) Ids() []int {
	ids := []int{}
	for _, linodeDisplayGroup := range nodes {
		for _, linode := range linodeDisplayGroup {
			ids = append(ids, linode.Id)
		}
	}
	return ids
}

func (nodes Linodes) Size() int {
	s := 0
	for _, grp := range nodes {
		s += len(grp)
	}
	return s
}

type Linode struct {
	Id           int    `json:"LINODEID"`
	Status       int    `json:"STATUS"`
	Label        string `json:"LABEL"`
	DisplayGroup string `json:"LPM_DISPLAYGROUP"`
	Ram          int    `json:"TOTALRAM"`
	Ips          []*LinodeIp
}

func (node *Linode) PublicIp() string {
	var ip string
	for _, linodeIp := range node.Ips {
		if linodeIp.Public == 1 {
			ip = linodeIp.Ip
			break
		}
	}
	return ip
}

func (node *Linode) PrivateIp() string {
	var ip string
	for _, linodeIp := range node.Ips {
		if linodeIp.Public == 0 {
			ip = linodeIp.Ip
			break
		}
	}
	return ip
}

func (node *Linode) IsRunning() bool {
	return node.Status == 1
}

func FetchLinodeList(api apiRequest) (Linodes, error) {
	api.AddAction("linode.list")

	datas, errs := api.GetJson()
	if len(errs) > 0 {
		errMsg := make([]string, len(errs))
		for i, err := range errs {
			errMsg[i] = err.Error()
		}
		return nil, errors.New(strings.Join(errMsg, "\n"))
	}

	var err error
	if len(datas) != 1 {
		return nil, fmt.Errorf("unexpected numbers of results")
	}
	var ls []Linode
	err = json.Unmarshal(datas[0], &ls)
	if err != nil {
		return nil, err
	}
	linodes := make(Linodes, len(ls))
	for _, linode := range ls {
		l := linode
		linodes[linode.DisplayGroup] = append(linodes[linode.DisplayGroup], &l)
	}

	return linodes, nil
}

// map of LinodeIps by their Linode.Id
type LinodeIps map[int][]*LinodeIp

type LinodeIp struct {
	LinodeId int    `json:"LINODEID"`
	Public   int    `json:"ISPUBLIC"`
	Ip       string `json:"IPADDRESS"`
}

func FetchLinodeIpList(api apiRequest, linodeIds []int) (LinodeIps, error) {
	apiMethod := "linode.ip.list"
	// one batch request for all linode Ids
	for _, linodeId := range linodeIds {
		action := api.AddAction(apiMethod)
		action.Set("LinodeID", strconv.Itoa(linodeId))
	}

	datas, errs := api.GetJson()
	if len(errs) > 0 {
		errMsg := make([]string, len(errs))
		for i, err := range errs {
			errMsg[i] = err.Error()
		}
		return nil, errors.New(strings.Join(errMsg, "\n"))
	}

	var err error
	linodeIps := make(LinodeIps, len(datas))
	for _, rawJson := range datas {
		var ipList []LinodeIp
		err = json.Unmarshal(rawJson, &ipList)
		if err != nil {
			return nil, err
		}
		for _, linodeIp := range ipList {
			i := linodeIp
			linodeIps[linodeIp.LinodeId] = append(linodeIps[linodeIp.LinodeId], &i)
		}
	}

	return linodeIps, nil
}

// Sort functions

type sortedLinodeIps []*LinodeIp

func (sorted sortedLinodeIps) Len() int {
	return len(sorted)
}
func (sorted sortedLinodeIps) Swap(i, j int) {
	sorted[i], sorted[j] = sorted[j], sorted[i]
}

// Sort by public IPs first
func (sorted sortedLinodeIps) Less(i, j int) bool {
	return sorted[i].Public > sorted[j].Public
}
func sortLinodeIps(ips []*LinodeIp) {
	sort.Sort(sortedLinodeIps(ips))
}

type sortedLinodes []*Linode

func (sorted sortedLinodes) Len() int {
	return len(sorted)
}
func (sorted sortedLinodes) Swap(i, j int) {
	sorted[i], sorted[j] = sorted[j], sorted[i]
}

// Sort by Label
func (sorted sortedLinodes) Less(i, j int) bool {
	return sorted[i].Label < sorted[j].Label
}
func sortLinodes(linodes []*Linode) {
	sort.Sort(sortedLinodes(linodes))
}
