package mr

import "fmt"
import "log"
import "net/rpc"
import "hash/fnv"
import "os"
import "io/ioutil"
// import "sort"
import "strconv"
import "encoding/json"
import "time"
//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}
// for sorting by key.
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}


//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	args := RequireTaskArgs{}
	reply := RequireTaskReply{}
	//periodically call "Coordinator.AssignTask"
	call("Coordinator.AssignTask", &args, &reply)
	for reply.NeedWait == true{
		time.Sleep(time.Second)

		reply.NeedWait = false
		call("Coordinator.AssignTask", &args, &reply)
	}

	if reply.Exit==true{
		return
	}
	
	if reply.Task=="map"{
		err:=callMap(mapf, reply.Filename, reply.NReduce)
		if err != nil{
			fmt.Println("callMap fail")
			return
		}

		mapFinishArgs := MapFinishArgs{reply.Filename}
		mapFinishReply := MapFinishReply{}
		ok := call("Coordinator.MapFinish", &mapFinishArgs, &mapFinishReply)
		if !ok{
			fmt.Println("call Coordinator.MapFinish fail")
		}
	} else if reply.Task=="reduce"{
		callReduce(reducef, reply.Filename)
	}
	
	

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}
type TmpOneFile map[string][]string
type TmpAllFile map[string]TmpOneFile
// call map function
func callMap(mapf func(string, string) []KeyValue, filename string, nReduce int) error {

	map_task_id := ihash(filename)
	
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)

	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()
	kva := mapf(filename, string(content))

	var tmpAllFile TmpAllFile = make(TmpAllFile)
	for _, kv := range kva{
		reduce_task_id := ihash(kv.Key) % nReduce
		tmpFilename := "mr-"+strconv.Itoa(map_task_id)+"-"+strconv.Itoa(reduce_task_id)+".json"
		
		if _,ok:=tmpAllFile[tmpFilename];!ok{
			tmpAllFile[tmpFilename] = make(TmpOneFile)
		}
		if _,ok:=tmpAllFile[tmpFilename][kv.Key];!ok{
			tmpAllFile[tmpFilename][kv.Key] = []string{}
		}

		tmpAllFile[tmpFilename][kv.Key] = append(tmpAllFile[tmpFilename][kv.Key], kv.Value)
		
	}
	for fn, ct := range tmpAllFile{
		imFile, err := os.Create(fn)
		if err != nil{
			fmt.Println("map task "+strconv.Itoa(map_task_id)+": fail to create intermediate file")
			return err
		}

		encoder := json.NewEncoder(imFile)

    	err = encoder.Encode(ct)
    	if err != nil {
        	fmt.Println("fail to encode json file", err.Error())
    	} else {
        	fmt.Println("map task "+strconv.Itoa(map_task_id)+": successfully encode an imtermediate file")
    	}	

		imFile.Close()
	}
	
	return nil
    
	
	
}
//call reduce function
func callReduce(reducef func(string, []string) string, reduce_task_id string) error {

	// call a reduce funciton
	return nil
}
//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()

	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
