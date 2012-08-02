package response

import "time"

//TestResult is an entity type that describes the sucessful running of a test
//and the output it generated.
type TestResult struct {
	ImportPath string    //import path of the test
	Revision   string    //revision of the source code
	RevDate    time.Time //when the revision was commit
	When       time.Time //when the test was run
	Output     string    //the output of the test
	Passed     bool      //the status of the test (pass/fail)
}

//WorkResult is an entity type that represents the result of the work item being
//run through the queue. It records any build failures or other errors in
//generating the test results.
type WorkResult struct {
	Success bool      //all the work items completed sucessfully
	When    time.Time //the time the work response was recorded
	Error   string    //a general error before any code could be built
}

//BuildFailure represents a code compilation error when trying to make a test
//binary.
type BuildFailure struct {
	ImportPath string    //the import path of the package
	Revision   string    //the revision of the build failure
	RevDate    time.Time //the date of the revision
	When       time.Time //when the build failure was created
	Output     string    //the ouput of the compiler error
}
