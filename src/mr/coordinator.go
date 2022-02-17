package mr


import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"
import "sync"
import "strconv"
import "fmt"
type MapTaskState struct{
	globalState int // 0-has map task to assign, 1-all the map tasks are precessing or finished, 2-all done
	m map[string]int //0-unstarted, 1-processing, 2-finished
	mu sync.Mutex
}
type ReduceTaskState struct{
	globalState int
	m map[string]int	// 0 - unstarted, 1-processing, 2-finished
	mu sync.Mutex
}
type Coordinator struct {
	// Your definitions here.
	mapTask2state MapTaskState 
	reduceTask2State ReduceTaskState
	nReduce int
}

// Your code here -- RPC handlers for the worker to call.

// return an as-yet-unstarted filename
func (c *Coordinator) AssignTask(args *RequireTaskArgs, reply *RequireTaskReply) error{
	
	c.mapTask2state.mu.Lock()
	defer c.mapTask2state.mu.Unlock()
	fmt.Println(c.mapTask2state.m)
	if c.mapTask2state.globalState==0{
		// if there is some map task to assign, the function will return in the `for` loop
		for key := range c.mapTask2state.m{
			if c.mapTask2state.m[key] == 0{
				c.mapTask2state.m[key] = 1
				reply.Filename  = key
				reply.Task = "map"
				reply.NReduce = c.nReduce
				return nil
			}
		}
		// reach here iff globalState == 1 or 2
		// after the following segement, `c.mapTask2State` is correctly set
		c.mapTask2state.globalState = 2
		for _, singleState := range c.mapTask2state.m{
			if singleState == 1{
				c.mapTask2state.globalState = 1
				break
			}
		}
		
	}
	
	
	if c.mapTask2state.globalState == 1{ // must not be `else if`
		reply.NeedWait = true
		return nil
	}
	
	// here `c.mapTask2state.globalState ==2`
	// deal with assigning Reduce Task	
	c.reduceTask2State.mu.Lock()
	defer c.reduceTask2State.mu.Unlock()
	
	// DEBUG
	fmt.Println(c.reduceTask2State.m)

	if c.reduceTask2State.globalState==0{
		// if there is some map task to assign, the function will return in the `for` loop
		for key := range c.reduceTask2State.m{
			if c.reduceTask2State.m[key] == 0{
				c.reduceTask2State.m[key] = 1
				reply.Filename  = key
				reply.Task = "reduce"
				reply.NReduce = c.nReduce
				return nil
			}
		}
		// reach here iff globalState == 1 or 2
		// after the following segement, `c.mapTask2State` is correctly set
		c.reduceTask2State.globalState = 2
		for _, singleState := range c.reduceTask2State.m{
			if singleState == 1{
				c.reduceTask2State.globalState = 1
				break
			}
		}
	}
	if c.reduceTask2State.globalState==1 {  // must not be `else if`
		reply.NeedWait = true
		return nil
	}
	// when reaching here, all the tasks are done
	reply.Exit = true
	return nil
}
func (c *Coordinator) MapFinish(args *MapFinishArgs, reply *MapFinishArgs) error {
	filename := args.Filename
	c.mapTask2state.mu.Lock()
	

	c.mapTask2state.m[filename] = 2
	
	//check whether all the map tasks are done
	c.mapTask2state.globalState = 2
	for _, singleState := range c.mapTask2state.m{
		if singleState == 1{
			c.mapTask2state.globalState = 1
			break
		}
	}
	for _, singleState := range c.mapTask2state.m{
		if singleState == 0{
			c.mapTask2state.globalState = 0
			break
		}
	}
	c.mapTask2state.mu.Unlock()
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

	c.mapTask2state.mu.Lock()
	// Your code here.
	c.mapTask2state.m = make(map[string]int)

	// Set initial state of map task
	for _,filename := range files{
		c.mapTask2state.m[filename] = 0 // 0 means `unstated`
	}
	c.mapTask2state.mu.Unlock()
	
	// Set initial state of reduce task
	c.reduceTask2State.mu.Lock()
	c.reduceTask2State.m = make(map[string]int)
	for i:=0;i<nReduce;i++{
		c.reduceTask2State.m[strconv.Itoa(i)] = 0
	}
	c.reduceTask2State.mu.Unlock()

	c.server()
	return &c
}
