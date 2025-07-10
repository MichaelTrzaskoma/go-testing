// Action
package vars

import (
	"fmt"
	"regexp"
	"strings"
)

type Action string

var InputAction Action

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

var InputEnv Env

func (e *Env) Set(s string) error {
	// switch strings.ToLower(s) {
	// case "uat", "prod":
	// 	*e = Env(s)
	// 	return nil
	// default:
	// 	return fmt.Errorf("invalid env: %s. Valid options are: Uat or Prod", s)
	// }
	return nil
}

func (e *Env) String() string {
	return string(*e)
}

// Object type for what we are targeting
type ObjType string

var InputObjType ObjType

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

var InputId Id

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

type WorkerGroupList []string

var InputWgList WorkerGroupList

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
