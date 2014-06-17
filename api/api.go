package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	ApiUrl           = "https://api.linode.com/"
	MaxBatchRequests = 25
)

type queryParams map[string]string

type apiAction struct {
	method string
	params queryParams
}

func (self *apiAction) Set(key, value string) {
	self.params[key] = value
}

func (self apiAction) values() queryParams {
	self.params["api_action"] = self.method
	return self.params
}

type apiRequest struct {
	apiKey  string
	baseUrl *url.URL
	actions []*apiAction
}

func NewApiRequest(apiKey string) (*apiRequest, error) {
	apiUrl, err := url.Parse(ApiUrl)
	if err != nil {
		return nil, err
	}
	var actions []*apiAction
	return &apiRequest{apiKey: apiKey, baseUrl: apiUrl, actions: actions}, nil
}

func (self *apiRequest) AddAction(method string) *apiAction {
	action := &apiAction{method: method, params: make(queryParams)}
	self.actions = append(self.actions, action)
	return action
}

func (self apiRequest) URL() []string {
	actionBatches := make([][]*apiAction, (len(self.actions)/MaxBatchRequests)+1)
	for j, action := range self.actions {
		i := j / MaxBatchRequests
		actionBatches[i] = append(actionBatches[i], action)
	}

	var urls []string
	for _, actions := range actionBatches {
		params := make(url.Values)
		params.Set("api_key", self.apiKey)
		params.Set("api_action", "batch")
		var requestArray []queryParams
		for _, action := range actions {
			requestArray = append(requestArray, action.values())
		}
		b, _ := json.Marshal(requestArray)
		params.Set("api_requestArray", string(b))
		u := self.baseUrl
		u.RawQuery = params.Encode()
		urls = append(urls, u.String())
	}
	return urls
}

type apiResponse struct {
	Action string `json:"ACTION"`
	Errors []struct {
		Code    int    `json:"ERRORCODE"`
		Message string `json:"ERRORMESSAGE"`
	} `json:"ERRORARRAY,omitempty"`
	Data json.RawMessage `json:"DATA,omitempty"`
}

func (self apiRequest) GetJson() ([]json.RawMessage, []error) {
	var datas []json.RawMessage
	var errs []error

	for _, url := range self.URL() {
		resp, err := http.Get(url)
		if err != nil {
			return nil, []error{err}
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)

		var apiResponses []apiResponse
		err = decoder.Decode(&apiResponses)
		if err != nil {
			return nil, []error{err}
		}

		for _, apiResp := range apiResponses {
			// Check for 'ERROR' attribute for any values, which would indicate an error
			if len(apiResp.Errors) > 0 {
				for _, e := range apiResp.Errors {
					errs = append(errs, errors.New(fmt.Sprintf("[Code: %d] %s\n", e.Code, e.Message)))
				}
				continue
			}
			datas = append(datas, apiResp.Data)
		}
	}

	return datas, errs
}
