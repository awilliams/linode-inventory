package api

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
)

const (
  API_URL = "https://api.linode.com/"
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
  apiUrl, err := url.Parse(API_URL)
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

func (self apiRequest) URL() string {
  params := make(url.Values)
  params.Set("api_key", self.apiKey)

  if len(self.actions) == 1 {
    for key, value := range self.actions[0].values() {
      params.Set(key, value)
    }
  } else if len(self.actions) > 1 {
    params.Set("api_action", "batch")
    var requestArray []queryParams
    for _, action := range self.actions {
      requestArray = append(requestArray, action.values())
    }
    b, err := json.Marshal(requestArray)
    if err != nil {
      log.Fatal(err)
    }
    params.Set("api_requestArray", string(b))
  }
  self.baseUrl.RawQuery = params.Encode()
  return self.baseUrl.String()
}

func (self apiRequest) GetJson(data interface{}) error {
  resp, err := http.Get(self.URL())
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  
  // the linode API does not use HTTP status codes to indicate errors,
  // rather it embeds in the JSON document the errors. When there is an error
  // the foramt of the `DATA` element changes as well, which would cause the json decode to fail.
  //
  // Here we first parse the json to see if it contains errors, then re-parse with the provided
  // json structure
  if linodeErr := checkForLinodeError(bytes.NewReader(body)); linodeErr != nil {
    return linodeErr
  }

  decoder := json.NewDecoder(bytes.NewReader(body))
  err = decoder.Decode(data)
  if err != nil {
    return err
  }
  
  return nil
}

type errorJson struct {
  Errors []struct {
    Code    int    `json:"ERRORCODE"`
    Message string `json:"ERRORMESSAGE"`
  } `json:"ERRORARRAY,omitempty"`
}

func checkForLinodeError(body *bytes.Reader) error {
  data := new(errorJson)
  decoder := json.NewDecoder(body)
  err := decoder.Decode(&data)
  if err != nil {
    // this is not actually an error, since there is not always an error present in the JSON
    return nil
  }
  if len(data.Errors) > 0 {
    var buf bytes.Buffer
    buf.WriteString("Api Error!\n")
    for _, e := range data.Errors {
      buf.WriteString(fmt.Sprintf("[Code: %d] %s\n", e.Code, e.Message))
    }
    return fmt.Errorf(buf.String())
  }
  return nil
}

func (self *apiRequest) GoString() string {
  s, err := url.QueryUnescape(self.URL())
  if err != nil {
    return ""
  }
  return s
}
