package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

// Interface for the ActionUtil
type ActionUtilsInt interface {
	AddAction(str string) error
	GetStats() string
}

// A wrapper struct for the action utils
type ActionUtil struct {
}

//json hints helps while de-serialization
type Action struct {
	Action string `json:"action"`
	Time   int    `json:"time"`
}

type AvgAction struct {
	Action string `json:"action"`
	Avg    int    `json:"avg"`
}

type AvgActions []AvgAction

//Use a map with the RWMutex to optimize for the multiple read and writes
// This is in global scope like static in other languages
var actionMap = struct {
	sync.RWMutex
	data map[string][]int
}{data: make(map[string][]int)}

//AddAction taken a json serialized string. Keep track of the actions with the map.
//This thread-safe map is used to calculate the average of the actions
func (action ActionUtil) AddAction(input string) error {
	incomingAction := Action{}
	err := json.Unmarshal([]byte(input), &incomingAction)
	if err != nil {
		log.Printf("Invalid json. Failed to unmarshal! %s", err.Error())
		return err
	}

	validationError := validateAction(incomingAction)
	if validationError != nil {
		return validationError
	}

	//Use a boolean channel as a locking mechanism to avoid race condition on write data.
	//This channel is used to wait till a signal is sent back from the go-routine when it completes.
	//NOTE: Here the unbuffered boolean channel is uses as a synchronous block only to process the writes on the map.
	//       Concurrent reads are fine. This enables to process the action concurrent to the caller.
	processChannel := make(chan bool)
	go processAction(incomingAction, processChannel)
	isComplete := <-processChannel

	//If the channel is done close it
	if isComplete {
		close(processChannel)
	}

	//Processed successfully, return nil
	return nil
}

//Helper function to process the actions
func processAction(action Action, channel chan bool) {
	//Sanitize the incoming data to lowercase
	actionKey := strings.ToLower(action.Action)

	//Check if the incoming action exists, if so sum it up and add the counts
	// else add the time value as new entry
	//Takes a write lock
	actionMap.Lock()
	if val, ok := actionMap.data[actionKey]; ok {
		actionMap.data[actionKey][0] = val[0] + action.Time
		actionMap.data[actionKey][1] += 1
	} else {
		actionMap.data[actionKey] = []int{action.Time, 1}
	}
	actionMap.Unlock()

	// Done processing the actions to map, send signal via channel
	channel <- true

}

func validateAction(incomingAction Action) error {
	if incomingAction.Action == "" {
		log.Printf("Invaild action: %s", incomingAction.Action)
		return errors.New(fmt.Sprintf("Invaild action: %s", incomingAction.Action))
	}

	if incomingAction.Time < 0 {
		log.Printf("Invaild time value: %d", incomingAction.Time)
		return errors.New(fmt.Sprintf("Invaild time value: %d", incomingAction.Time))
	}

	//If no validation issue, return nil
	return nil
}

func (action ActionUtil) GetStats() string {
	//Collect the map data into array of AvgAction
	avgStructs := AvgActions{}

	//Takes a read Lock, multiple read locks can be used simultaneously
	actionMap.RLock()
	for key, val := range actionMap.data {
		avgStructs = append(avgStructs, AvgAction{key, val[0] / val[1]})
	}
	actionMap.RUnlock()

	//Marshal the data to structured json format
	avgJson, err := json.Marshal(avgStructs)
	if err != nil {
		log.Println("Marshalling error", err)
	}

	log.Printf("Stats: %s", avgJson)
	//Send the string value from the marshalled data
	return string(avgJson)
}
