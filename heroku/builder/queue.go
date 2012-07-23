package builder

import "github.com/zeebo/goci/app/rpc"

//create our local queues.
var (
	builderQueue = rpc.NewBuilderQueue()
	runnerQueue  = rpc.NewRunnerQueue()
)

//register our queues
func init() {
	if err := rpcServer.RegisterService(runnerQueue, ""); err != nil {
		bail(err)
	}
	if err := rpcServer.RegisterService(builderQueue, ""); err != nil {
		bail(err)
	}
}
