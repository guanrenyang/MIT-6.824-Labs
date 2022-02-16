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
- [ ] `AssignTask`处理没有任务再分配的情况
- [ ] 没有任务的Worker周期性请求任务

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

**TODO**

- [ ] use TempFile to deal with the failing if workers
- [ ] store `mapTask2state` in disk in case of coordinator failing

**Possible Improvement**

* If the worker called `MapFinish` but didn't receive reply, some problem of the network could occur. In this case, the worker just need to call `MapFinish` again instead of doing the map task again.
