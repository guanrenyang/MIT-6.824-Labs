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
	mapTask2state TaskStateMap //0-unstarted, 1-processing, 2-finished
	mapTask2id map[string]int
	nReduce int
}

// Your code here -- RPC handlers for the worker to call.

// return an as-yet-unstarted filename
func (c *Coordinator) AssignTask(args *RequireTaskArgs, reply *RequireTaskReply) error{
	c.mapTask2state.mu.Lock()
	for key := range c.mapTask2state.m{
		if c.mapTask2state.m[key] == 0{
			c.mapTask2state.m[key] = 1
			c.mapTask2state.mu.Unlock()
			reply.Filename  = key
			reply.Task = "map"
			reply.Map_task_id = c.mapTask2id[key]
			reply.NReduce = c.nReduce
			return nil
		}
	}
	c.mapTask2state.mu.Unlock()
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

	c.nReduce = nReduce

	// Your code here.
	c.mapTask2state.m = make(map[string]int)
	c.mapTask2id = make(map[string]int)
	// Set initial state
	c.mapTask2state.mu.Lock()
	for i,filename := range files{
		c.mapTask2state.m[filename] = 0 // 0 means `unstated`
		c.mapTask2id[filename] = i
	}
	c.mapTask2state.mu.Unlock()

	c.server()
	return &c
}
