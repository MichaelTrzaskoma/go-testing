package main

import (
	"flag"
	"fmt"
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

//

func main() {

	flag.Var(&action, "action", "Set the action (Create, Update, or Delete)")
	flag.Var(&env, "env", "Set the env (Uat or Prod)")
	flag.Parse()
	fmt.Println("Selected action:", action)
	fmt.Println("Selected env:", env)
}
