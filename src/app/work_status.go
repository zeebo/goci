package main

import "launchpad.net/mgo"

type WorkStatus int

const (
	WorkStatusPassed = iota
	WorkStatusFailed
	WorkStatusWary
)

func (w WorkStatus) String() (r string) {
	switch w {
	case WorkStatusPassed:
		r = "Passed"
	case WorkStatusFailed:
		r = "Failed"
	case WorkStatusWary:
		r = "Be Wary"
	default:
		panic("unknown work status")
	}
	return
}

var work_status_job = mgo.MapReduce{
	Map: `function() {
	      	this.builds.forEach(function(build) {
	      		build.tests.forEach(function(test) {
	      			emit(build._id, test.passed);
	      		});
	      	});
	      }`,

	Reduce: `function(key, values) {
	         	var result = true;
	         	values.forEach(function(value) {
	         		result = result && value;
	         	});
	         	return result;
	         }`,
}

func work_status(db *mgo.Database, id string) (status WorkStatus, err error) {
	status = WorkStatusWary

	var res []struct {
		ID    string `bson:"_id"`
		Value bool
	}
	_, err = db.C(collection).Find(d{"_id": id}).MapReduce(work_status_job, &res)
	if err != nil {
		return
	}

	var passed, failed = true, true
	for _, r := range res {
		passed = passed && r.Value
		failed = failed && !r.Value
	}
	switch {
	case passed && !failed:
		status = WorkStatusPassed
	case !passed && failed:
		status = WorkStatusFailed
	}
	return
}
