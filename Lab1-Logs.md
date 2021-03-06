# MIT-6.824-Lab1
### Requirements
* workers may crash but coornaditor won't
### 2022.2.7

`mr/worker.go`

Worker向Coordinator发送rpc来请求map任务

`mr/rpc.go`

定义了用于分配map任务的args和reply

`mr/coordinator.go`

1. 添加类型，用以保存 **文件名→状态** 的映射。注意需要lock
    
    ```go
    type TaskStateMap struct{
    	m map[string]int
    	mu sync.Mutex
    }
    ```
    
2. 实现函数 `AssignTask(args *RequireTaskArgs, reply *RequireTaskReply)` 。*到目前为止：如果有未开始的工作就分配工作，**如果没有未开始的工作没有处理。**

**TODO**
- [x] `AssignTask`处理没有任务再分配的情况

### 2022.2.8

`mr/worker.go`

add function `callMap` to call a mapf function. **But I haven't store intermediate k/v in disk yet**

`mr/rpc.go`

change struct `RequireTaskReply`, add seqments `NReduce`, `Map_task_id`, and `NReduce`

`mr/coordinator.go`

modify function `AssignTask` in reponse to the change in `mr.rpc.go`

**TODO**

- [x]  Think how
to modify `Worker()` to store intermediate key-value pairs in disk.

use iHash(word)%NReduce to compute the id of reduce task

I currently add some simple code in worker.go, but they need big modifications.

### 2022.2.10

`mr/worker.go`

尝试将中间结果保存使用 `json` 格式保存在磁盘上。 *现在还在学习json格式保存的方法*

**TODO**

- [x] 将中间key/value保存在磁盘上。将变成一个字典，然后将字典编码为json。

### 2022.2.14

将中间结果保存在内存中————使用两级map

### 2022.2.16

store intermediate kvs in disk

delete `Map_task_id` in struct `RequireTaskReply` and use `ihash(filename)` instead. Thus the map task id will no longer rely on the sequence of reading input files, and **we could know the id even if the coordiantor fails.**

add struct `MapDone` in coordinator to indicate whether all the map tasks are finished

**TODO**

- [x] use TempFile to deal with the failing if workers
- [ ] store `mapTask2state` in disk in case of coordinator failing
- [ ] refactor the code of assigning tasks in one function
- [ ] refactor the if-branch of `globalState` to switch-branch. 
- [x] **finish `callReduce` function and corresponding code in `Worker()`**

**Possible Improvements**

* If the worker called `MapFinish` but didn't receive reply, some problem of the network could occur. In this case, the worker just need to call `MapFinish` again instead of doing the map task again.

* I use the same logic to assgin `map` and `redcue` tasks. Thus the two segments of code are similar and should be in one function.

* Improve time efficiency of `Coordinator.MapFinish`

### 2022.2.17

**Fix a bug:** reduce task ids might not fill 0-nReduce-1, so it must be determined by map task instead of nReduce. 
Because `nReduce` is much smaller than `nMap`, I take the strategy that the reducer return `done` to the coordinator if it receives a empty reduce task.

**`callReduce` Function in `/mr/Worker.go`:** 
1. read intermediates kvs in a multi-level map `tmpAllFile`
2. merge intermediate kvs in a single-level map `intermediate`
3. call `reducef` and store the out file on disk

**TODO**

- [x] steal code from `mrsequential.go` and finish `callReduce`
- [x] send a *Reduce Task Finish* signal to the coordinator when the reduce task not existing. This funtion could be in `ReduceFinish` or in another function.
- [x] add a `ReduceFinish` function in `mr/coordinator.go`
- [x] **没有任务的Worker周期性请求任务. `test-mr.sh` only creates three worker threads.**

**Possible Improvements**

* reducer merging multiple intermediate files: current verson is reading all file in memory and then merge them, so the total memory usage is `2*all kvs`. It could be imporved by merging while reading files.

### 2022.2.18

* Workers call the coordinator for a job unless that the coordinator asks it to exit or that the coordinator has exited.
* **add function `checkAlive`.**: if `checkAlive[i]==0`, which means the ith worker has failed-- after 10 seconds it hasn't finish the job, the coordinator will mask the job as `unstarted`--`m[i]==0` and assign the job to another worker; when assgin task i, which means setting `m[i]=1`, `checkAlive[i]` should be set to `10`, which means the coordiantor will think the worker as crached if the task could be done in 10 seconds; when the job is done, which means setting `m[i]=2`, `checkAlive[i]` should be set to `0`, which means if task i is done then there is no need to check it.
* `go cheakAlive` in `coordinator.go` to run the correctness of workers in separate goroutines
* workers create temporary files initially. After calling `MapFinish` or `ReduceFinish` and getting the reponse, workers rename the files to `mr-X-Y` format for map tasks and `mr-out-X` format for reduce tasks.

## PASSED ALL TESTS
## 2022.2.7-2022.2.18

