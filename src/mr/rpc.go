package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

// arguments for workers requiring tasks
type RequireTaskArgs struct{}
type RequireTaskReply struct{
	Filename string
	Task string //"map" or "reduce"
	NReduce int
	NeedWait bool
	Exit bool
}
//
// arguments for finishing map task
type MapFinishArgs struct{
	Filename string
}
type MapFinishReply struct{}
//
// arguments for finishing reduce task
type ReduceFinishArgs struct{
	Filename string
}
type ReduceFinishReply struct{}

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
