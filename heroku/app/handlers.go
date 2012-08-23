package main

import (
	_ "github.com/zeebo/goci/app/response"  //handle responses from workers
	_ "github.com/zeebo/goci/app/test"      //simple test handlers
	_ "github.com/zeebo/goci/app/tracker"   //handle tracking workers
	_ "github.com/zeebo/goci/app/workqueue" //handle queuing/dispatching work

	_ "net/http/pprof" //add pprof support
)
