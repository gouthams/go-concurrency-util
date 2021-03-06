# Overview of actions library utils
Action utils has two simple functions:
1) Add Action - Accepts a json serialized string and keeps track of the average time for each distinct action.
2) Statistics - Returns an array of serialized jsons added using the add action function.

This util uses go-routine, go-channels, read write mutex and wait groups to make the library thread-safe 
and also to test its efficiency.
    
## Assumptions:

1) Actions are treated in a case insensitive manner

2) Library uses a map to store and process data and it is shared, so any caller/importer of the utils will see the same
    instance results . Assumption is since the concurrent users need to access the same copy to validate the concurrency.
    In general library should avoid using share memory or go-routines if possible. In context of this library inside 
    an application, singleton pattern could have been used to ensure there is only one instance exists.
    
    Use of RWMutex for the map to share data. 
    Default sync.Map use case recommends only to use it when its read heavy or key are processed in disjoint manner.
    https://golang.org/pkg/sync/. Using the RWMutex has advantage on multiple read scenario, so RWMutex is preferred.
     
3) Logging is limit to standard logger. In practice need a logger with logging levels supported for debug/trace logs.
   An example third party logrus logger could be used https://github.com/Sirupsen/logrus

4) Unit test leverages max logical CPUs to run tests simultaneously for a large data set operation. 

### Install and Build
Requires Golang installed. Please follow the instruction from here https://golang.org/doc/install

This library is developed with go version 1.14.4

Download the library from https://github.com/gouthams/go-concurrency-util

Need access to github or github mirror repo to resolve dependencies.

To run a sample example, do the following
```shell script
cd example
go build
./example or example.exe (windows)
```

To execute unit tests, do the following
```shell script
cd actions
go test
```

To execute the unit test with the coverage profile, do the following
```shell script
cd actions
go test -coverprofile cp.out
go tool cover -html=cp.out
```

To execute the unit test with mutex profiler via interactive shell, do the following
```shell script
cd actions
go test -mutexprofile=mutex.out
go tool pprof actions.test mutex.out
top5
quit
```

### Unit test
For assertion in unit test, this library is used https://github.com/stretchr/testify. 
This go module dependency should be resolved during the build time.  

Executing unit test could be longer in few systems. 
Especially TestAddActionMultipleEntryConcurrentAndParallel test uses maximum available logical cores
to run the 2 million concurrent adds and a million concurrent get stats.

It takes about 30 seconds to run the tests on 12 logical core machine with 2.6 GHz 6-Core Intel Core i7 CPU.

### Future Consideration:
   1) Adding a remove/delete action function.
   2) Parse the json as bulk that could support multiple payload entry in as single add operation
   
