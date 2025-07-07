package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var templateHost string = "localhost"
var templatePort string = "19000"
var templateProtocol string = "http"
var templateWG string = "default"

var baseUrl = templateProtocol + "://" + templateHost + ":" + templatePort

type CribConfig map[string]interface{}

// Action
type Action string

var action Action

func (a *Action) Set(s string) error {
	switch strings.ToLower(s) {
	case "uat", "prod":
		*a = Action(s)
		return nil
	default:
		return fmt.Errorf("invalid action: %s. Valid options are: Create, Update, or Delete", s)
	}
}

func (a *Action) String() string {
	return string(*a)
}

// Env
type Env string

var env Env

func (e *Env) Set(s string) error {
	switch strings.ToLower(s) {
	case "uat", "prod":
		*e = Env(s)
		return nil
	default:
		return fmt.Errorf("invalid env: %s. Valid options are: Uat or Prod", s)
	}
}

func (e *Env) String() string {
	return string(*e)
}

// Object type for what we are targeting
type ObjType string

var objType ObjType

func (e *ObjType) Set(s string) error {
	switch strings.ToLower(s) {
	case "Source", "Destination", "Pipeline", "Pack", "GlobalVariable", "Lookup":
		*e = ObjType(s)
		return nil
	default:
		return fmt.Errorf("invalid env: %s. Valid options are: Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup", s)
	}

}

func (e *ObjType) String() string {
	return string(*e)
}

// Id of the data object we are targetting
type Id string

var id Id

func (e *Id) Set(s string) error {
	allowedCharsRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	if allowedCharsRegex.MatchString(s) {
		*e = Id(s)
		return nil
	} else {
		return fmt.Errorf("invalid Id String: %s. String must be alphanumeric with special characters '-' and '_' allowed", s)
	}

}

func (e *Id) String() string {
	return string(*e)
}

// Worker Group List for what is being targetted
type WorkerGroupList []string

var wgList WorkerGroupList

func (e *WorkerGroupList) Set(s string) error {
	var validWgList []string
	allowedCharsRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+$")
	for listVal := range strings.SplitSeq(s, ",") {
		if allowedCharsRegex.MatchString(listVal) {
			validWgList = append(validWgList, listVal)
		} else {
			fmt.Println("Skipping invalid worker group:", listVal)
		}
	}

	*e = validWgList

	return nil
}

func (e WorkerGroupList) String() string {
	return strings.Join(e, ", ")
}

func tokenApiCall() string {
	url := baseUrl + "/api/v1/auth/login"
	authBody := map[string]string{"username": "admin", "password": "Test1234"}
	authBodyJson, _ := json.Marshal(authBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(authBodyJson))
	req.Header = http.Header{"content-type": {"application/json"}}
	resp, err := client.Do(req)
	// Handling response error
	if err != nil {
		fmt.Println(err.Error())
		return ""
	} else {
		// Read Body
		responseData, err := io.ReadAll(resp.Body)
		//Handling Read error
		if err != nil {
			fmt.Println(err.Error())
		}

		// Struct for Response, caring about token
		type Token struct {
			Token string `json:"token"`
		}

		var t Token
		json.Unmarshal(responseData, &t)

		return "Bearer " + t.Token
	}
}

func getWorkerGroups(token string) {
	url := baseUrl + "/api/v1/master/groups"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"Authorization": {token}}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())

	} else {
		// Read Body
		responseData, err := io.ReadAll(resp.Body)
		//Handling Read error
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(string(responseData))

	}
}

func getSource(token string) CribConfig {
	url := baseUrl + "/api/v1/m/" + templateWG + "/system/inputs/splunk_uf"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"Authorization": {token}}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
		return nil

	} else {
		// Read Body
		responseData, err := io.ReadAll(resp.Body)
		//Handling Read error
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		//var result map[string]interface{}

		// _ = json.Unmarshal([]byte(responseData), &result)
		var response struct {
			Items []CribConfig `json:"items"`
			Count int          `json:"count"`
		}

		_ = json.Unmarshal([]byte(responseData), &response)

		fmt.Println(string(responseData))

		delete(response.Items[0], "status")
		delete(response.Items[0], "notifications")
		// fmt.Println(response.Items[0])

		return response.Items[0]

		// response.Items[0]["id"] = "test"
		// response.Items[0]["port"] = 1923

		// jsonBytes, _ := json.Marshal(response.Items[0])
		// jsonString := string(jsonBytes)

		// fmt.Println(jsonString)
		// // fmt.Println((result["items"]))

		// url_2 := base_url + "/api/v1/m/" + template_wg + "/system/inputs"

		// sourceBodyJson, _ := json.Marshal(response.Items[0])
		// req, _ := http.NewRequest("POST", url_2, bytes.NewBuffer(sourceBodyJson))
		// req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
		// resp, _ := client.Do(req)
		// responseData_2, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(responseData_2))

	}
}

func createSource(baseApiUrl string, workerGroup string, token string, sourceConfig []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/inputs"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(sourceConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData_2, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData_2))
}

func updateSource(baseApiUrl string, workerGroup string, token string, sourceConfig []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/inputs/test"
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(sourceConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData_2, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData_2))
}

func main() {

	flag.Var(&action, "action", "Set the action (Create, Update, or Delete)")
	flag.Var(&env, "env", "Set the env (Uat or Prod)")
	flag.Var(&objType, "objType", "Defines the object type of the configuration item that we are targeting (Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup)")
	flag.Var(&id, "id", "Set the id for configuration item you're looking to target")
	flag.Var(&wgList, "wgList", "List of worker groups to target")
	flag.Parse()
	fmt.Println("Selected action:", action)
	fmt.Println("Selected env:", env)
	fmt.Println("Selected object type", objType)
	fmt.Println("Selected id:", id)
	fmt.Println("Selected worker group(s) are:", wgList)

	token := tokenApiCall()
	//fmt.Println("Token here:", val)

	getWorkerGroups(token)
	getSourceConfig := getSource(token)
	fmt.Print(getSourceConfig)

	getSourceConfig["id"] = "test"
	getSourceConfig["port"] = 1943
	sourceConfigBytes, _ := json.Marshal(getSourceConfig)
	//createSource(baseUrl, templateWG, token, sourceConfigBytes)
	updateSource(baseUrl, templateWG, token, sourceConfigBytes)

}
