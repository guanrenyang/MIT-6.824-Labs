package mr


import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"

type TaskStateMap struct{
	m map[string]int
	mu sync.Mutex
}
type Coordinator struct {
	// Your definitions here.
	task2state TaskStateMap //0-unstarted, 1-processing, 2-finished
}

// Your code here -- RPC handlers for the worker to call.

// return an as-yet-unstarted filename
func (c *Coordinator) AssignTask(args *RequireTaskArgs, reply *RequireTaskReply) error{
	c.task2state.mu.Lock()
	for key := range c.task2state.m{
		if c.task2state.m[key] == 0{
			c.task2state.m[key] = 1
			c.task2state.mu.Unlock()
			reply.Filename  = key
			return nil
		}
	}
	c.task2state.mu.Unlock()
	reply.Filename = "not file"
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}


//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	ret := false

	// Your code here.


	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}

	// Your code here.
	c.task2state.m = make(map[string]int)
	// Set initial state
	c.task2state.mu.Lock()
	for _,filename := range files{
		c.task2state.m[filename] = 0
	}
	c.task2state.mu.Unlock()

	c.server()
	return &c
}
