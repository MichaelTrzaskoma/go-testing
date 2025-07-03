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
	url := "http://localhost:19000/api/v1/auth/login"
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
	url := "http://localhost:19000/api/v1/master/groups"

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

	val := tokenApiCall()
	fmt.Println("Token here:", val)

	getWorkerGroups(val)

}
