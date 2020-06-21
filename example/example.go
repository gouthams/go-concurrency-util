package main

import (
	"github.com/gouthams/awesomeProject1/actions"
)

func main() {
	// Sample demo of the actions utils
	var action actions.ActionUtilsInt
	action = actions.ActionUtil{}

	// Look at the actionsutils_test TestAddActionMultipleEntryConcurrentAndParallel for concurrent calls using go-routines
	action1 := `{"action":"jump", "time":100}`
	action2 := `{"action":"run", "time":75}`
	action3 := `{"action":"jump", "time":200}`
	action4 := `{"action":"jump", "time":300}`

	action.AddAction(action1)
	action.AddAction(action2)

	action.GetStats()

	action.AddAction(action3)
	action.AddAction(action4)

	action.GetStats()
}
