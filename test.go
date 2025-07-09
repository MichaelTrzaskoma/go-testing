package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

// var templateHost string = "localhost"
// var templatePort string = "19000"
// var templateProtocol string = "http"
// var templateWG string = "default"

// var baseUrl = templateProtocol + "://" + templateHost + ":" + templatePort

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
	allowedCharsRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+(?:\.csv)?$`)
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
func retryHttp(client *http.Client, req *http.Request, retryCount int) (*http.Response, error) {
	var (
		retries int = retryCount
		resp    *http.Response
		err     error
	)

	for retries > 0 {
		resp, err = client.Do(req)

		if err != nil {
			retries -= 1
		} else {
			break
		}
	}

	return resp, err
}

func tokenApiCall(baseUrl string, username string, password string) (string, error) {
	url := baseUrl + "/api/v1/auth/login"
	authBody := map[string]string{"username": username, "password": password}
	authBodyJson, _ := json.Marshal(authBody)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(authBodyJson))
	req.Header = http.Header{"content-type": {"application/json"}}

	var (
		maxRetries int = 5
		resp       *http.Response
		httpErr    error
	)

	resp, httpErr = retryHttp(client, req, maxRetries)

	if resp != nil && httpErr == nil {
		defer resp.Body.Close()

		responseData, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("unable to properly read response body %w", readErr)
		}

		// Struct for Response, caring about token
		type Token struct {
			Token string `json:"token"`
		}
		var tok Token

		unMarshErr := json.Unmarshal(responseData, &tok)
		if unMarshErr != nil {
			return "", fmt.Errorf("unable to extract token value from respones body: %w", unMarshErr)
		}

		return "Bearer " + tok.Token, nil
	} else {
		return "", fmt.Errorf("token unable to be retrieved from url %s : %w Attempted (%d) time(s)", url, httpErr, maxRetries)
	}
	// Handling response error
}

func getWorkerGroups(baseUrl string, token string) ([]byte, error) {
	url := baseUrl + "/api/v1/master/groups"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"Authorization": {token}}

	var (
		maxRetries int = 5
		resp       *http.Response
		httpErr    error
	)

	resp, httpErr = retryHttp(client, req, maxRetries)

	if resp != nil && httpErr == nil {
		defer resp.Body.Close()

		responseData, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("unable to properly read response body %w", readErr)
		} else {
			return responseData, nil
		}
	} else {
		return nil, fmt.Errorf("worker groups unable to be retrieved from url %s : %w Attempted (%d) time(s)", url, httpErr, maxRetries)
	}
}

func getLookupContent(baseApiUrl string, workerGroup string, token string, lookupId string) ([]byte, error) {
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/lookups/" + lookupId + "/content?raw=0"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}

	var (
		maxRetries int = 5
		resp       *http.Response
		httpErr    error
	)

	resp, httpErr = retryHttp(client, req, maxRetries)

	if resp != nil && httpErr == nil {
		defer resp.Body.Close()

		responseData, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("unable to properly read response body %w", readErr)
		}

		return responseData, nil
	} else {
		return nil, fmt.Errorf("lookup content for %s unable to be retrieved from url %s : %w Attempted (%d) time(s)", lookupId, url, httpErr, maxRetries)
	}

}

func uploadLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, lookupContent []byte) ([]byte, error) {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/lookups/?filename=" + lookup_id
	//objectConfigBytes, _ := json.Marshal(responseData)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(lookupContent))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"text/csv"}}

	var (
		maxRetries int = 5
		resp       *http.Response
		httpErr    error
	)

	resp, httpErr = retryHttp(client, req, maxRetries)

	if resp != nil && httpErr == nil {
		responseData, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("unable to properly read response body %w", readErr)
		}

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

		unMarshErr := json.Unmarshal([]byte(responseData), &lookupPayloadFileinfo)
		if unMarshErr != nil {
			return nil, fmt.Errorf("unable to extract lookup upload details from respones body: %w", unMarshErr)
		}

		lookupApiPayload := LookupPayload{
			Id: lookup_id,
			FileInfo: fileInfo{
				Filename: lookupPayloadFileinfo.FileName,
			},
		}

		lookUpPayloadJson, marshErr := json.Marshal(lookupApiPayload)
		if marshErr != nil {
			return nil, fmt.Errorf("unable to format upload payload to be used for patching: %w", marshErr)
		}

		return lookUpPayloadJson, nil
	} else {
		return nil, fmt.Errorf("uploading lookup failed when trying url %s : %w Attempted (%d) time(s)", url, httpErr, maxRetries)
	}

}

func patchLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, patchPayload []byte) error {
	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + "/system/lookups/" + lookup_id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(patchPayload))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	var (
		maxRetries int = 5
		// resp       *http.Response
		httpErr error
	)

	_, httpErr = retryHttp(client, req, maxRetries)
	if httpErr != nil {
		return fmt.Errorf("patching lookup failed when trying url %s : %w Attempted (%d) time(s)", url, httpErr, maxRetries)
	} else {
		return nil
	}
}

func getDataObj(baseApiUrl string, workerGroup string, token string, id string, objType string) ([]byte, error) {

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

	var (
		maxRetries int = 5
		resp       *http.Response
		httpErr    error
	)

	resp, httpErr = retryHttp(client, req, maxRetries)

	if resp != nil && httpErr == nil {
		defer resp.Body.Close()

		responseData, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("unable to properly read response body %w", readErr)
		}

		var response struct {
			Items []CribConfig `json:"items"`
			Count int          `json:"count"`
		}

		unMarshErr := json.Unmarshal([]byte(responseData), &response)
		if unMarshErr != nil {
			return nil, fmt.Errorf("unable to extract token value from respones body: %w", unMarshErr)
		}

		if len(response.Items) > 0 {
			delete(response.Items[0], "status")
			delete(response.Items[0], "notifications")

			objectConfig, marshErr := json.Marshal(response.Items[0])
			if marshErr != nil {
				return nil, fmt.Errorf("unable to format get %s payload to be used for patching: %w", objType, marshErr)
			}
			return objectConfig, nil

		} else {
			return nil, fmt.Errorf("%s content for %s returned empty from url %s", objType, id, url)
		}

	} else {
		return nil, fmt.Errorf("%s content for %s unable to be retrieved from url %s : %w Attempted (%d) time(s)", objType, id, url, httpErr, maxRetries)
	}
}

func updateDataObj(baseApiUrl string, workerGroup string, token string, id string, objConfig []byte, objType string) error {
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
		log.Fatalf("invalid Object Type provided: %s. Valid options are: Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup", objType)
	}

	client := &http.Client{}
	url := baseApiUrl + "/api/v1/m/" + workerGroup + objEndpoint + "/" + id
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(objConfig))
	req.Header = http.Header{"Authorization": {token}, "content-type": {"application/json"}}
	var (
		maxRetries int = 5
		//resp       *http.Response
		httpErr error
	)

	_, httpErr = retryHttp(client, req, maxRetries)

	if httpErr != nil {
		return fmt.Errorf("patching lookup failed when trying url %s : %w Attempted (%d) time(s)", url, httpErr, maxRetries)
	} else {
		return nil
	}
}

func relicateConfigPatch(origBaseApiUrl string, origWorkerGroup string, origToken string, targetBaseApiUrl string, targetWorkerGroups []string, targetToken string, objType string, objId string) {

	switch strings.ToLower(objType) {
	case "source", "destination", "pipeline", "globalvariable":
		objectConfigBytes, getDataErr := getDataObj(origBaseApiUrl, origWorkerGroup, origToken, objId, objType)
		if getDataErr != nil {
			log.Fatalf("Fatal error encountered with initial GET for %s '%s': %v", objType, objId, getDataErr)
		}
		for _, workerGroup := range targetWorkerGroups {
			updateErr := updateDataObj(targetBaseApiUrl, workerGroup, targetToken, objId, objectConfigBytes, objType)
			if updateErr != nil {
				log.Printf("Skipped updating %s '%s' on worker group '%s' due to following error during patching: %v", objType, objId, workerGroup, updateErr)
			} else {
				log.Printf("Successfully updated %s '%s' on worker group %s", objType, objId, workerGroup)
			}
		}
	case "lookup":
		if strings.HasSuffix(objId, ".csv") {
			objectContent, getLookupErr := getLookupContent(origBaseApiUrl, origWorkerGroup, origToken, objId)
			if getLookupErr != nil {
				log.Fatalf("Fatal error encountered with initial GET for %s '%s': %v", objType, objId, getLookupErr)
			}
			for _, workerGroup := range targetWorkerGroups {
				objectUpload, uploadErr := uploadLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectContent)

				if uploadErr != nil {
					log.Printf("Skipped updating %s '%s' on worker group '%s' due to following error during PUT: %v", objType, objId, workerGroup, uploadErr)
					continue
				}

				patchErr := patchLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectUpload)
				if patchErr != nil {
					log.Printf("Skipped updating %s '%s' on worker group '%s' due to following error during PATCH: %v", objType, objId, workerGroup, patchErr)
				} else {
					log.Printf("Successfully updated %s '%s' on worker group '%s'", objType, objId, workerGroup)
				}
			}
		} else {
			fmt.Println("Error: Expected object Id for lookup to end with '.csv', invalid lookup submitted")
		}

	default:
		log.Fatalf("(%s) not valid object type, ignored", objType)
	}
}

func main() {
	var templateProtocol, templateHost, templatePort, templateWorkerGroup, templateUser, templatePass, targetProtocol, targetHost, targetPort, targetUser, targetPass, templateUrl, targetUrl string

	// Global Var Loading
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, relying on environment variables alone")
	}
	templateProtocol = os.Getenv("TEMPLATE_API_PROTOCOL")
	templateHost = os.Getenv("TEMPLATE_HOST")
	templatePort = os.Getenv("TEMPLATE_PORT")
	templateWorkerGroup = os.Getenv("TEMPLATE_WORKER_GROUP")
	templateUser = os.Getenv("TEMPLATE_API_USERNAME")
	templatePass = os.Getenv("TEMPLATE_API_PASSWORD")

	flag.Var(&env, "env", "Set the env (Uat or Prod)")

	if strings.ToLower(string(env)) == "prod" {
		targetProtocol = os.Getenv("PROD_API_PROTOCOL")
		targetHost = os.Getenv("PROD_HOST")
		targetPort = os.Getenv("PROD_PORT")
		targetUser = os.Getenv("PROD_API_USERNAME")
		targetPass = os.Getenv("PROD_API_PASSWORD")
	} else {
		targetProtocol = os.Getenv("UAT_API_PROTOCOL")
		targetHost = os.Getenv("UAT_HOST")
		targetPort = os.Getenv("UAT_PORT")
		targetUser = os.Getenv("UAT_API_USERNAME")
		targetPass = os.Getenv("UAT_API_PASSWORD")
	}

	templateUrl = templateProtocol + "://" + templateHost
	if len(strings.TrimSpace(templatePort)) != 0 {
		templateUrl = templateUrl + ":" + templatePort
	}

	targetUrl = targetProtocol + "://" + targetHost
	if len(strings.TrimSpace(targetPort)) != 0 {
		targetUrl = targetUrl + ":" + targetPort
	}

	flag.Var(&action, "action", "Set the action (Create, Update, or Delete)")
	flag.Var(&objType, "objType", "Defines the object type of the configuration item that we are targeting (Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup)")
	flag.Var(&id, "id", "Set the id for configuration item you're looking to target")
	flag.Var(&wgList, "wgList", "List of worker groups to target")

	flag.Parse()
	log.Println("Selected action:", action)
	log.Println("Selected env:", env)
	log.Println("Selected object type:", objType)
	log.Println("Selected object id:", id)
	log.Println("Selected worker group(s) are:", wgList)

	templateToken, tempTokenErr := tokenApiCall(templateUrl, templateUser, templatePass)

	if tempTokenErr != nil {
		log.Fatal("Fatal error encountered: ", tempTokenErr)
	}

	targetToken, targetTokenErr := tokenApiCall(targetUrl, targetUser, targetPass)
	if targetTokenErr != nil {
		log.Fatal("Fatal error encountered: ", targetTokenErr)
	}
	//fmt.Println("Token here:", val)

	// getWorkerGroups(token)
	relicateConfigPatch(templateUrl, templateWorkerGroup, templateToken, targetUrl, wgList, targetToken, string(objType), string(id))
	// getLookup(baseUrl, "default", token, "test.csv")
	// getSourceConfig := getSource(token)
	// fmt.Print(getSourceConfig)

	// getSourceConfig["id"] = "test"
	// getSourceConfig["port"] = 1943
	// sourceConfigBytes, _ := json.Marshal(getSourceConfig)
	// //createSource(baseUrl, templateWG, token, sourceConfigBytes)
	// updateSource(baseUrl, templateWG, token, sourceConfigBytes)

}
