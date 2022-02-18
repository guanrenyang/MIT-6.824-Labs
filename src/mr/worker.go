package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// import "sort"

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
	reply := RequireTaskReply{NeedWait: true}
	//periodically call "Coordinator.AssignTask"
	// call("Coordinator.AssignTask", &args, &reply)
	for reply.NeedWait == true {
		time.Sleep(time.Second)

		reply.NeedWait = false
		ok := call("Coordinator.AssignTask", &args, &reply)
		if !ok { // if the coordinator has exited, the worker should alse exit
			return
		}
		if reply.NeedWait == true { // if the worker need wait, just skip to the next requiring
			continue
		}

		// if the worker doesn't need to wait, do the job
		if reply.Exit == true {
			return
		}
		//DEBUG
		// fmt.Println(reply.Filename)

		if reply.Task == "map" {
			err := callMap(mapf, reply.Filename, reply.NReduce)
			if err != nil {
				fmt.Println("callMap fail")
				return
			}

		} else if reply.Task == "reduce" {
			err := callReduce(reducef, reply.Filename)
			if err != nil {
				fmt.Println("callReduce fail")
				return
			}
		}
		//DEBUG
		// fmt.Println(reply.Task, reply.Filename)
		// set `NeedWait=true` to require another job
		reply.NeedWait = true
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}

type TmpOneFile map[string][]string
type TmpAllFile map[string]TmpOneFile

// // do a job
// func doJob()

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
	for _, kv := range kva {
		reduce_task_id := ihash(kv.Key) % nReduce
		tmpFilename := "mr-" + strconv.Itoa(map_task_id) + "-" + strconv.Itoa(reduce_task_id) + ".json"

		if _, ok := tmpAllFile[tmpFilename]; !ok {
			tmpAllFile[tmpFilename] = make(TmpOneFile)
		}
		if _, ok := tmpAllFile[tmpFilename][kv.Key]; !ok {
			tmpAllFile[tmpFilename][kv.Key] = []string{}
		}

		tmpAllFile[tmpFilename][kv.Key] = append(tmpAllFile[tmpFilename][kv.Key], kv.Value)

	}

	tmp2perm := make(map[string]string) // temp file name -> permernanent file name
	pwd, _ := os.Getwd()
	for fn, ct := range tmpAllFile { //fn->filename, ct->contents
		tmpImFile, err := ioutil.TempFile(pwd, "tmp")
		tmp2perm[tmpImFile.Name()] = fn
		if err != nil {
			fmt.Println("map task " + strconv.Itoa(map_task_id) + ": fail to create intermediate file")
			return err
		}

		encoder := json.NewEncoder(tmpImFile)

		err = encoder.Encode(ct)
		if err != nil {
			fmt.Println("fail to encode json file", err.Error())
		} else {
			// fmt.Println("map task " + strconv.Itoa(map_task_id) + ": successfully encode an imtermediate file")
		}

		tmpImFile.Close()
	}

	mapFinishArgs := MapFinishArgs{filename}
	mapFinishReply := MapFinishReply{}
	ok := call("Coordinator.MapFinish", &mapFinishArgs, &mapFinishReply)
	if !ok {
		fmt.Println("call Coordinator.MapFinish fail")
	} else {
		for tmpName, permName := range tmp2perm {
			os.Rename(tmpName, permName)
		}
	}
	return nil

}

//call reduce function
func callReduce(reducef func(string, []string) string, reduce_task_id string) error {

	// load all intermediate files
	pwd, _ := os.Getwd()

	filepathNames, err := filepath.Glob(filepath.Join(pwd, "mr-*-"+reduce_task_id+".json"))
	if err != nil {
		log.Fatal(err)
	}

	if len(filepathNames) == 0 {
		reduceFinishArgs := ReduceFinishArgs{reduce_task_id}
		reduceFinishReply := ReduceFinishReply{}
		call("Coordinator.ReduceFinish", &reduceFinishArgs, &reduceFinishReply)
	}

	var tmpAllFile TmpAllFile = make(TmpAllFile)
	for _, filename := range filepathNames {

		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}

		decoder := json.NewDecoder(file)

		tmpOneFile := make(TmpOneFile)

		err = decoder.Decode(&tmpOneFile)
		if err != nil {
			fmt.Println("reduce task " + reduce_task_id + ": fail to decode a json file")
		} else {
			// fmt.Println("reduce task " + reduce_task_id + ": successfully decode a json file")
		}

		tmpAllFile[filename] = tmpOneFile

		file.Close()
	}
	//DEBUG
	// fmt.Println("here")

	// merge multiple maps
	intermediate := make(map[string][]string)
	for _, singleFileMap := range tmpAllFile {
		for key := range singleFileMap {
			intermediate[key] = append(intermediate[key], singleFileMap[key]...)
		}
	}

	oname := "mr-out-" + reduce_task_id
	tmpOFile, _ := ioutil.TempFile(pwd, "tmp")

	//
	// call Reduce on each distinct key in intermediate[],
	// and print the result to mr-out-reduce_task_id.
	//
	for key, values := range intermediate {

		output := reducef(key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(tmpOFile, "%v %v\n", key, output)
	}

	reduceFinishArgs := ReduceFinishArgs{reduce_task_id}
	reduceFinishReply := ReduceFinishReply{}
	ok := call("Coordinator.ReduceFinish", &reduceFinishArgs, &reduceFinishReply)
	if !ok {
		fmt.Println("call Coordinator.ReduceFinish fail")
	} else {
		tmpOFile.Close()
		os.Rename(tmpOFile.Name(), oname)
		// fmt.Println("reduce task " + reduce_task_id + ": successfully reduce ")
	}
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
