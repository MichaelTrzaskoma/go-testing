package main

import (
	"criblPatching/functions"
	"criblPatching/vars"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// var templateHost string = "localhost"
// var templatePort string = "19000"
// var templateProtocol string = "http"
// var templateWG string = "default"

// var baseUrl = templateProtocol + "://" + templateHost + ":" + templatePort

// Worker Group List for what is being targetted

func relicateConfigPatch(origBaseApiUrl string, origWorkerGroup string, origToken string, targetBaseApiUrl string, targetWorkerGroups []string, targetToken string, objType string, objId string) {

	switch strings.ToLower(objType) {
	case "source", "destination", "pipeline", "globalvariable":
		objectConfigBytes, getDataErr := functions.GetDataObj(origBaseApiUrl, origWorkerGroup, origToken, objId, objType)
		if getDataErr != nil {
			log.Fatalf("Fatal error encountered with initial GET for %s '%s': %v", objType, objId, getDataErr)
		}
		for _, workerGroup := range targetWorkerGroups {
			updateErr := functions.UpdateDataObj(targetBaseApiUrl, workerGroup, targetToken, objId, objectConfigBytes, objType)
			if updateErr != nil {
				log.Printf("Skipped updating %s '%s' on worker group '%s' due to following error during patching: %v", objType, objId, workerGroup, updateErr)
			} else {
				log.Printf("Successfully updated %s '%s' on worker group %s", objType, objId, workerGroup)
			}
		}
	case "lookup":
		if strings.HasSuffix(objId, ".csv") {
			objectContent, getLookupErr := functions.GetLookupContent(origBaseApiUrl, origWorkerGroup, origToken, objId)
			if getLookupErr != nil {
				log.Fatalf("Fatal error encountered with initial GET for %s '%s': %v", objType, objId, getLookupErr)
			}
			for _, workerGroup := range targetWorkerGroups {
				objectUpload, uploadErr := functions.UploadLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectContent)

				if uploadErr != nil {
					log.Printf("Skipped updating %s '%s' on worker group '%s' due to following error during PUT: %v", objType, objId, workerGroup, uploadErr)
					continue
				}

				patchErr := functions.PatchLookup(targetBaseApiUrl, workerGroup, targetToken, objId, objectUpload)
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

	flag.Var(&vars.InputEnv, "env", "Set the env (Uat or Prod)")

	if strings.ToLower(string(vars.InputEnv)) == "prod" {
		targetProtocol = os.Getenv("PROD_API_PROTOCOL")
		targetHost = os.Getenv("PROD_HOST")
		targetPort = os.Getenv("PROD_PORT")
		targetUser = os.Getenv("PROD_API_USERNAME")
		targetPass = os.Getenv("PROD_API_PASSWORD")
	} else {
		if strings.ToLower(string(vars.InputEnv)) != "uat" {
			vars.InputEnv = "UAT (Defaulted)"
		}
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

	// flag.Var(&action, "action", "Set the action (Create, Update, or Delete)")
	flag.Var(&vars.InputObjType, "objType", "Defines the object type of the configuration item that we are targeting (Source, Destination, Pipeline, Pack, GlobalVariable, or Lookup)")
	flag.Var(&vars.InputId, "id", "Set the id for configuration item you're looking to target")
	flag.Var(&vars.InputWgList, "wgList", "List of worker groups to target")

	flag.Parse()
	log.Print("Running tool with the following settings:")
	log.Printf("Environment: (%s) | Action: (Update) | Object Type: (%s) | Object Id: (%s) | Target Worker Group(s): (%s)", vars.InputEnv, vars.InputObjType, vars.InputId, vars.InputWgList)

	templateToken, tempTokenErr := functions.TokenApiCall(templateUrl, templateUser, templatePass)

	if tempTokenErr != nil {
		log.Fatal("Fatal error encountered: ", tempTokenErr)
	}

	targetToken, targetTokenErr := functions.TokenApiCall(targetUrl, targetUser, targetPass)
	if targetTokenErr != nil {
		log.Fatal("Fatal error encountered: ", targetTokenErr)
	}
	//fmt.Println("Token here:", val)

	// getWorkerGroups(token)
	relicateConfigPatch(templateUrl, templateWorkerGroup, templateToken, targetUrl, vars.InputWgList, targetToken, string(vars.InputObjType), string(vars.InputId))
	// getLookup(baseUrl, "default", token, "test.csv")
	// getSourceConfig := getSource(token)
	// fmt.Print(getSourceConfig)

	// getSourceConfig["id"] = "test"
	// getSourceConfig["port"] = 1943
	// sourceConfigBytes, _ := json.Marshal(getSourceConfig)
	// //createSource(baseUrl, templateWG, token, sourceConfigBytes)
	// updateSource(baseUrl, templateWG, token, sourceConfigBytes)

}
