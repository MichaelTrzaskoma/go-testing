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
	case "source", "destination", "pipeline", "pack", "globalvariable", "lookup":
		*e = ObjType(s)
		return nil
	default:
		return fmt.Errorf("invalid Object Type provided: %s. Valid options are: Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup", s)
	}

}

func (e *ObjType) String() string {
	return string(*e)
}

// Id of the data object we are targetting
type Id string

var id Id

func (e *Id) Set(s string) error {
	allowedCharsRegex := regexp.MustCompile("^[a-zA-Z0-9_-]+(?:\\.csv)?$")
	if allowedCharsRegex.MatchString(s) {
		*e = Id(s)
		return nil
	} else {
		return fmt.Errorf("invalid Id String: %s. String must be alphanumeric with special characters '-' and '_' allowed or if a lookup, ensure ending is .csv", s)
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

func tokenApiCall(username string, password string) string {
	url := baseUrl + "/api/v1/auth/login"
	authBody := map[string]string{"username": username, "password": password}
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

func getSource(baseApiUrl string, workerGroup string, token string, id string) CribConfig {
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/inputs/" + id

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

		var response struct {
			Items []CribConfig `json:"items"`
			Count int          `json:"count"`
		}

		_ = json.Unmarshal([]byte(responseData), &response)

		fmt.Println(string(responseData))

		delete(response.Items[0], "status")
		delete(response.Items[0], "notifications")

		return response.Items[0]

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

func updateSource(baseApiUrl string, workerGroup string, token string, id string, sourceConfig []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/inputs/" + id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(sourceConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData_2, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData_2))
}

func getDest(baseApiUrl string, workerGroup string, token string, id string) CribConfig {
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/outputs/" + id

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

		var response struct {
			Items []CribConfig `json:"items"`
			Count int          `json:"count"`
		}

		_ = json.Unmarshal([]byte(responseData), &response)

		fmt.Println(string(responseData))

		delete(response.Items[0], "status")
		delete(response.Items[0], "notifications")

		return response.Items[0]

	}
}

func createDest(baseApiUrl string, workerGroup string, token string, destConfig []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/outputs"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(destConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData))
}

func updateDest(baseApiUrl string, workerGroup string, token string, id string, destConfig []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/outputs/" + id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(destConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData))
}

func getLookupContent(baseApiUrl string, workerGroup string, token string, lookup_id string) []byte {
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/lookups/" + lookup_id + "/content?raw=0"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
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

		}

		// var response struct {
		// 	Items []CribConfig `json:"items"`
		// 	Count int          `json:"count"`
		// }

		fmt.Println(string(responseData))
		return responseData
	}

}

func uploadLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, lookupContent []byte) []byte {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/lookups/?filename=" + lookup_id
	//objectConfigBytes, _ := json.Marshal(responseData)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(lookupContent))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"text/csv"}}
	resp, _ := client.Do(req)
	responseData, _ := io.ReadAll(resp.Body)

	fmt.Println(string(responseData))

	var lookupPayloadFileinfo struct {
		FileName string `json:"filename"`
	}

	type fileInfo struct {
		Filename string `json:"filename"`
	}
	type LookupPayload struct {
		Id       string   `json:"id"`
		FileInfo fileInfo `json:"fileInfo"`
	}
	_ = json.Unmarshal([]byte(responseData), &lookupPayloadFileinfo)

	lookupApiPayload := LookupPayload{
		Id: lookup_id,
		FileInfo: fileInfo{
			Filename: lookupPayloadFileinfo.FileName,
		},
	}

	lookUpPayloadJson, _ := json.Marshal(lookupApiPayload)
	return lookUpPayloadJson
}

func patchLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, patchPayload []byte) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + "GroupA" + "/system/lookups/" + lookup_id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(patchPayload))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData, _ := io.ReadAll(resp.Body)
	print(responseData)

}

func getDataObj(baseApiUrl string, workerGroup string, token string, id string, objType string) CribConfig {

	var objEndpoint string

	switch strings.ToLower(objType) {
	case "source":
		objEndpoint = "/system/inputs"
	case "destination":
		objEndpoint = "/system/outputs"
	case "pipeline":
		objEndpoint = "/pipelines"
	case "globalvariable":
		objEndpoint = "/lib/vars"
	default:
		fmt.Printf("invalid Object Type provided: %s. Valid options are: Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup", objType)
	}

	url := baseApiUrl + "/api/v1/m/" + workerGroup + objEndpoint + "/" + id

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

		var response struct {
			Items []CribConfig `json:"items"`
			Count int          `json:"count"`
		}

		_ = json.Unmarshal([]byte(responseData), &response)

		fmt.Println(string(responseData))

		delete(response.Items[0], "status")
		delete(response.Items[0], "notifications")

		return response.Items[0]

	}
}

func updateDataObj(baseApiUrl string, workerGroup string, token string, id string, objConfig []byte, objType string) {
	var objEndpoint string

	switch strings.ToLower(objType) {
	case "source":
		objEndpoint = "/system/inputs"
	case "destination":
		objEndpoint = "/system/outputs"
	case "pipeline":
		objEndpoint = "/pipelines"
	case "globalvariable":
		objEndpoint = "/lib/vars"
	default:
		fmt.Printf("invalid Object Type provided: %s. Valid options are: Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup", objType)
	}

	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + objEndpoint + "/" + id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(objConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	resp, _ := client.Do(req)
	responseData, _ := io.ReadAll(resp.Body)
	fmt.Println(string(responseData))
}

func relicateConfigPatch(origBaseApiUrl string, origWorkerGroup string, origToken string, targetBaseApiUrl string, targetWorkerGroups []string, targetToken string, objType string, objId string) {

	switch strings.ToLower(objType) {
	case "source", "destination", "pipeline", "globalvariable":
		objectConfig := getDataObj(origBaseApiUrl, origWorkerGroup, origToken, objId, objType)
		objectConfig["port"] = 1953 // testing
		objectConfigBytes, _ := json.Marshal(objectConfig)
		for _, workerGroup := range targetWorkerGroups {
			updateDataObj(targetBaseApiUrl, workerGroup, targetToken, objId, objectConfigBytes, objType)
		}
	case "lookup":
		if strings.HasSuffix(objId, ".csv") {
			objectContent := getLookupContent(origBaseApiUrl, origWorkerGroup, origToken, objId)
			for _, workerGroup := range targetWorkerGroups {
				objectUpload := uploadLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectContent)
				patchLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectUpload)

			}
		} else {
			fmt.Println("Error: Expected object Id for lookup to end with '.csv', invalid lookup submitted")
		}

	default:
		fmt.Println("Not valid, ignored")
	}
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
	fmt.Println("Selected object type:", objType)
	fmt.Println("Selected id:", id)
	fmt.Println("Selected worker group(s) are:", wgList)

	token := tokenApiCall("admin", "Test1234")
	//fmt.Println("Token here:", val)

	// getWorkerGroups(token)
	relicateConfigPatch(baseUrl, "default", token, baseUrl, wgList, token, string(objType), string(id))
	// getLookup(baseUrl, "default", token, "test.csv")
	// getSourceConfig := getSource(token)
	// fmt.Print(getSourceConfig)

	// getSourceConfig["id"] = "test"
	// getSourceConfig["port"] = 1943
	// sourceConfigBytes, _ := json.Marshal(getSourceConfig)
	// //createSource(baseUrl, templateWG, token, sourceConfigBytes)
	// updateSource(baseUrl, templateWG, token, sourceConfigBytes)

}
