package actions

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"runtime"
	"sync"
	"testing"
)

//WaitGroup for go-routine to access AddAction and GetStats
var waitGrp sync.WaitGroup

var action ActionUtilsInt

func init() {
	action = ActionUtil{}
}

type MockAction struct {
	mock.Mock
}

func TestAddActionSingleEntry(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	action1 := `{"action":"jump", "time":100}`
	expectedResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, action.GetStats())
}

func TestAddActionMultipleEntry(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	action1 := `{"action":"jump", "time":100}`
	action2 := `{"action":"run", "time":75}`
	action3 := `{"action":"jump", "time":200}`
	action4 := `{"action":"jump", "time":300}`
	expectedResult1 := `{"action":"jump","avg":200}`
	expectedResult2 := `{"action":"run","avg":75}`

	err1 := action.AddAction(action1)
	err2 := action.AddAction(action2)
	err3 := action.AddAction(action3)
	err4 := action.AddAction(action4)

	assert.Nil(t, err1, err2, err3, err4)
	assert.Contains(t, action.GetStats(), expectedResult1)
	assert.Contains(t, action.GetStats(), expectedResult2)
}

func TestAddActionMultipleEntryConcurrent(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	// Sum of 1+2+3+.....+100 = 50.5
	expectedResult1 := `{"action":"jump","avg":50}`
	expectedResult2 := `{"action":"run","avg":10}`

	times := 100
	//Wait unit two for loop go-routines are done
	waitGrp.Add(2 * times)

	// Calls 100 times to insert the jump with i values
	for i := 1; i <= times; i++ {
		action := fmt.Sprintf(`{"action":"jump", "time":%d}`, i)

		go wrapAddAction(action)
	}

	// Calls 100 times to insert run with 10
	for i := 1; i <= times; i++ {
		action := `{"action":"run", "time":10}`

		go wrapAddAction(action)
	}

	//wait until all go-routines are done
	waitGrp.Wait()

	assert.Contains(t, action.GetStats(), expectedResult1)
	assert.Contains(t, action.GetStats(), expectedResult2)
}

func TestAddActionMultipleEntryConcurrentAndParallel(t *testing.T) {
	//Clean up after this test
	defer deleteActionsMap()

	//This allows go-routines to run parallel on multiple logical cores available
	log.Printf("CPUs to use for parallel processing: %d", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Sum of (1000000 * 1000001) / 2 = 500000.5
	expectedResult1 := `{"action":"jump","avg":500000}`
	expectedResult2 := `{"action":"run","avg":10}`

	times := 1000 * 1000
	//Wait unit 3 loops of go-routines are done
	waitGrp.Add(3 * times)

	// Calls number of times to insert the jump with i values
	for i := 1; i <= times; i++ {
		action := fmt.Sprintf(`{"action":"jump", "time":%d}`, i)

		go wrapAddAction(action)
	}

	// Calls number of times to insert run with 10
	for i := 1; i <= times; i++ {
		action := `{"action":"run", "time":10}`

		go wrapAddAction(action)
	}

	// Call getstats specified number of times
	for i := 1; i <= times; i++ {
		go wrapGetStats()
	}

	//wait until all go-routines are done
	waitGrp.Wait()

	assert.Contains(t, action.GetStats(), expectedResult1)
	assert.Contains(t, action.GetStats(), expectedResult2)
}

func TestJsonExtraAndDuplicateData(t *testing.T) {
	action1 := `{"action":"jump", "fox":"ran", "5":34, "time":150, "Valid": "string", "time":200}`
	expectedResult := `[{"action":"jump","avg":200}]`

	err := action.AddAction(action1)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, action.GetStats())
}

// Negative test cases
func TestInvalidJsonAction(t *testing.T) {

	action1 := `{"abcdef":"jump", "time":100}`
	expectedResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)
	assert.NotContains(t, action.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild action:")
	}
}

func TestInvalidJsonActionEmpty(t *testing.T) {

	action1 := `{"action":""}`
	unknownResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)
	assert.NotContains(t, action.GetStats(), unknownResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild action:")
	}
}

func TestInvalidJsonTime(t *testing.T) {

	action1 := `{"action":"jump", "time":-100}`
	expectedResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)
	assert.NotContains(t, action.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "Invaild time value:")
	}
}

func TestInvalidJsonTimeLimit(t *testing.T) {

	action1 := `{"action":"jump", "time":922337203685477580900}`
	expectedResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)
	assert.NotContains(t, action.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "json: cannot unmarshal number 922337203685477580900")
	}
}

func TestInvalidJsonTimeDouble(t *testing.T) {

	action1 := `{"action":"jump", "time":1.123}`
	expectedResult := `[{"action":"jump","avg":100}]`

	err := action.AddAction(action1)
	assert.NotContains(t, action.GetStats(), expectedResult)
	if err != nil {
		assert.Contains(t, err.Error(), "json: cannot unmarshal number 1.123")
	}
}

func (mc *MockAction) GetStats() string {
	args := mc.Called()
	return args.String(0)
}

func TestMockGetStats(t *testing.T) {
	mc := new(MockAction)
	mc.On("GetStats").Return(`{"action":"jump","avg":100}`)
	stat := mc.GetStats()

	assert.Equal(t, `{"action":"jump","avg":100}`, stat)
	mc.AssertExpectations(t)
}

// Internal private helper functions for unit tests

func wrapAddAction(actionStr string) error {
	//Indicate the completion via deferred done
	defer waitGrp.Done()

	return action.AddAction(actionStr)
}

func wrapGetStats() string {
	//Indicate the completion via deferred done
	defer waitGrp.Done()

	return action.GetStats()
}

func deleteActionsMap() {
	for entry := range actionMap.data {
		delete(actionMap.data, entry)
	}
}
