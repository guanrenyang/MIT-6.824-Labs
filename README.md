# MIT-6.824-Lab1

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

- [ ] Modify `Worker()` to store intermediate key-value pairs in disk. 

use iHash(word)%NReduce to compute the id of reduce task

I currently add some simple code in worker.go, but they need big modifications.