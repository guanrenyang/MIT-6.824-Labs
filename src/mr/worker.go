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
	ok := call("Coordinator.AssignTask", &args, &reply)
	if !ok{
		fmt.Printf("Assign Task failed!")
	}
	
	if reply.Task=="map"{
		callMap(mapf, reply.Filename, reply.Map_task_id, reply.NReduce)
	}
	
	

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}
type TmpOneFile map[int][]string
type TmpAllFile map[string]TmpOneFile
// call map function
func callMap(mapf func(string, string) []KeyValue, filename string, map_task_id int, nReduce int)  {

	
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
		if _,ok:=tmpAllFile[tmpFilename][reduce_task_id];!ok{
			tmpAllFile[tmpFilename][reduce_task_id] = []string{}
		}

		tmpAllFile[tmpFilename][reduce_task_id] = append(tmpAllFile[tmpFilename][reduce_task_id], kv.Value)
		
	}
	for fn, ct := range tmpAllFile{
		imFile, err := os.Create(fn)
		if err != nil{
			fmt.Println("map task "+strconv.Itoa(map_task_id)+": fail to create intermediate file")
			return
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
