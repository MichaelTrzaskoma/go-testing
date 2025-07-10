package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type CribConfig map[string]interface{}

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

func TokenApiCall(baseUrl string, username string, password string) (string, error) {
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

func GetWorkerGroups(baseUrl string, token string) ([]byte, error) {
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

func GetLookupContent(baseApiUrl string, workerGroup string, token string, lookupId string) ([]byte, error) {
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

func UploadLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, lookupContent []byte) ([]byte, error) {
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

func PatchLookup(baseApiUrl string, workerGroup string, token string, lookup_id string, patchPayload []byte) error {
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

func GetDataObj(baseApiUrl string, workerGroup string, token string, id string, objType string) ([]byte, error) {

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

func UpdateDataObj(baseApiUrl string, workerGroup string, token string, id string, objConfig []byte, objType string) error {
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
