//package notifications dispatches notifications about build status changes
package notifications

import (
	"github.com/zeebo/goci/app/entities"
	"github.com/zeebo/goci/app/httputil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"labix.org/v2/mgo/txn"
	"net/http"
	"strings"
	"time"
)

const handleUrl = "/notifications/dispatch"

func init() {
	http.Handle(handleUrl, httputil.Handler(dispatchNotifications))
}

const (
	attemptTime = 1 * time.Minute
	maxAttempts = 2
)

func dispatchNotifications(w http.ResponseWriter, req *http.Request, ctx httputil.Context) (e *httputil.Error) {
	//find all documents that are waiting or (processing and their attempt is
	//taking too long)
	type L []interface{}
	selector := bson.M{
		"$or": L{
			bson.M{"status": entities.NotifStatusWaiting},
			bson.M{
				"status":            entities.NotifStatusProcessing,
				"attemptlog.0.when": bson.M{"$lt": time.Now().Add(-1 * attemptTime)},
			},
		},
	}
	iter := ctx.DB.C("Notification").Find(selector).Iter()

	var n entities.Notification
	for iter.Next(&n) {
		//if it's processing with too may attempts then just give up
		if len(n.AttemptLog) >= maxAttempts {
			ctx.Infof("Notification %s had too many attempts", n.ID)

			ops := []txn.Op{{
				C:  "Notification",
				Id: n.ID,
				Assert: bson.M{
					"status":   entities.NotifStatusProcessing,
					"revision": n.Revision,
				},
				Update: bson.M{
					"$set": bson.M{"status": entities.NotifStatusError},
					"$inc": bson.M{"revision": 1},
				},
			}}

			//try to update the notification
			err := ctx.R.Run(ops, bson.NewObjectId(), nil)
			if err == txn.ErrAborted {
				ctx.Infof("Lost race updating notification %s", n.ID)
				err = nil
			}
			if err != nil {
				ctx.Errorf("Error updating notification %s: %s", n.ID, err)
			}

			continue
		}

		err := dispatchNotificationItem(ctx, &n)
		if err != nil {
			ctx.Errorf("Error processing notification %s: %s", n.ID, err)
			continue
		}

		//update the thing as being done
		ops := []txn.Op{{
			C:  "Notification",
			Id: n.ID,
			Assert: bson.M{
				"revision": n.Revision,
			},
			Update: bson.M{
				"$inc": bson.M{"revision": 1},
				"$set": bson.M{"status": entities.NotifStatusCompleted},
			},
		}}

		err = ctx.R.Run(ops, bson.NewObjectId(), nil)
		if err == txn.ErrAborted {
			ctx.Infof("Lost the race setting the notification %s to complete", n.ID)
			err = nil
		}
		if err != nil {
			ctx.Errorf("Error setting notification %s to complete: %s", n.ID, err)
		}
	}

	//check for errors in the iteration
	if err := iter.Err(); err != nil {
		ctx.Errorf("Error iterating over notifications: %s", err)
		e = httputil.Errorf(err, "Error iterating over notifications")
		return
	}

	return
}

//multiError holds multiple errors
type multiError []error

//Error makes a multiError an error.
func (m multiError) Error() string {
	var buf []string
	for _, e := range m {
		if e == nil {
			continue
		}
		buf = append(buf, e.Error())
	}
	return strings.Join(buf, "\n===\n")
}

//isNil returns true if every error in the multiError is nil
func (m multiError) isNil() bool {
	v := true
	for _, e := range m {
		v = v && e == nil
	}
	return v
}

func dispatchNotificationItem(ctx httputil.Context, n *entities.Notification) (err error) {
	//create an attempt for this notification
	a := entities.NotifAttempt{
		When: time.Now(),
		ID:   bson.NewObjectId(),
	}

	//push it to the start
	log := append([]entities.NotifAttempt{a}, n.AttemptLog...)

	//transactionally acquire ownership of this notification
	ops := []txn.Op{{
		C:  "Notification",
		Id: n.ID,
		Assert: bson.M{
			"revision": n.Revision,
		},
		Update: bson.M{
			"$inc": bson.M{"revision": 1},
			"$set": bson.M{
				"attemptlog": log,
				"status":     entities.NotifStatusProcessing,
			},
		},
	}}

	err = ctx.R.Run(ops, bson.NewObjectId(), nil)
	if err == txn.ErrAborted {
		ctx.Infof("Lost the race dispatching a notification")
		err = nil
	}
	if err != nil {
		return
	}

	//inc the revision locally
	n.Revision++

	//try to load up the last two test results to see if there was a delta.
	var test entities.TestResult
	err = ctx.DB.C("TestResult").FindId(n.Test).One(&test)
	if err != nil {
		return
	}

	//attempt to grab the test previous to it
	var oneResult bool
	var prev entities.TestResult
	query := bson.M{
		"revdate": bson.M{"$lt": test.RevDate},
	}
	err = ctx.DB.C("TestResult").Find(query).Sort("-revdate").One(&prev)
	if err == mgo.ErrNotFound {
		err = nil
		oneResult = true
	}
	if err != nil {
		return
	}

	//figure out if we meet the conditions to notify
	var perform bool
	switch strings.ToLower(n.Config.NotifyOn) {
	case "pass":
		perform = test.Status == "Pass"
	case "fail":
		perform = test.Status == "Fail"
	case "error":
		perform = test.Status == "Error"
	case "wontbuild":
		perform = test.Status == "WontBuild"
	case "problem":
		perform = false ||
			test.Status == "Fail" ||
			test.Status == "Error" ||
			test.Status == "WontBuild"
	case "always":
		perform = true
	case "change":
		perform = !oneResult && test.Status != prev.Status
	}

	//if we have nothing to perform, we're done
	if !perform {
		return
	}

	//do the url and jabber concurrently
	errs := make(chan error)
	go func() { errs <- sendUrlNotification(ctx, n.Config.NotifyURL, test) }()
	go func() { errs <- sendJabberNotification(ctx, n.Config.NotifyJabber, test) }()

	//store the errors from it
	var me multiError
	for i := 0; i < 2; i++ {
		me = append(me, <-errs)
	}
	if !me.isNil() {
		err = me
	}

	return
}

func sendUrlNotification(ctx httputil.Context, u string, test entities.TestResult) (err error) {
	if u == "" {
		return
	}

	ctx.Infof("Send url notification:", u)
	return
}

func sendJabberNotification(ctx httputil.Context, u string, test entities.TestResult) (err error) {
	if u == "" {
		return
	}

	ctx.Infof("Send jabber notification:", u)
	return
}
