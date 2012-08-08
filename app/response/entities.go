package response

import "time"

//TestResult is an entity type that describes the sucessful running of a test
//and the output it generated.
type TestResult struct {
	ImportPath string    //import path of the test
	Revision   string    //revision of the source code
	RevDate    time.Time //when the revision was commit
	When       time.Time //when the test was run
	Output     []byte    //the output of the test
	Status     string    //the status of the build: (Pass/Fail/WontBuild/Error)
}

//WorkResult is an entity type that represents the result of the work item being
//run through the queue. It records any build failures or other errors in
//generating the test results.
type WorkResult struct {
	Success  bool      //all the work items completed sucessfully
	Revision string    //the revision of the build (if known)
	RevDate  time.Time //when the revision was commit (if known)
	When     time.Time //the time the work response was recorded
	Error    []byte    //a general error before any code could be built
}
