DISCLAIMER: All the code here is written by me and not vibe coded. So expect issues with optimization and stuff. Also there might be design that might not be the best approach but I chose whatever I felt would be good. I'd like to apologize for the grammar and the lack of professionality in this readme file. Finally, I used claude to review the code because I don't know any senior or mentor in this space, so you can say this is vibe reviewed. If you actually look at this and find some issue, feel free to convey. Thanks...

# Concurrent Task Execution Engine
A concurrent task execution engine built in Go to explore and understand the concurrency of Go. Or more lke wanting to check out how good concurrency in Go is.
## What it does
This engine concurrently executes tasks using a number of workers, retries failed tasks based on failure classification, tracks metrics, and manages task state throughout the lifecycle. The current task types are simple arithmetic operations. But this architecture is designed to handle any task type without modification to the core engine (You might wanna change the retry limit and task timeout. IF you are considering on using this which I highly doubt it).
## Architecture
### Task Lifecycle
Submission — Tasks are submitted to a waiting queue. If the waiting queue is full, the task is dropped immediately and reported back to the caller. Once in the waiting queue, the task is in pending state. If a task enters the waiting queue it will be executed for sure unless the engine is interrupted\
Scheduling — A dedicated scheduler moves tasks from the waiting queue to the task queue. If the task queue is full, the scheduler blocks until space is available — tasks wait in the waiting queue rather than being dropped at this stage. More on why I implemented a scheduler is down the line.\
Execution — Workers constantly listen for tasks on the task queue (blocking receive). When a task arrives, the worker sets its state to running and begins execution. Each task has a configurable time limit — if execution exceeds this limit, the task is marked as failed with a transient classification and sent for retry.\
Completion — A successfully completed task reaches the completed terminal state and its result is recorded.
Failure and Retry — Failed tasks are classified into three categories:

Permanent — bad input or invalid task type. Retrying would produce the same result so the task is dropped immediately.\
Transient — temporary failures such as timeouts. The task is eligible for retry up to a configurable retry limit.\
System — unexpected failures such as panics. Treated similarly to transient failures. This also should catch the tasks that fail due to the lack of system resource.

Failed retryable tasks are pushed to a retry queue and the retry count is incremneted. The scheduler moves them back to the task queue where they are picked up and executed again. Once the retry limit is reached, the task is marked as permanently failed and dropped.
## Why a Separate Scheduler
Early versions had workers putting failed tasks directly back into the task queue. This caused deadlock. What happened was that, the task queue got filled up and so the worker was waiting for the task queue to have space. This meant that the worker is waiting (or spinning you might say) wasting resources and reduing the efficiency of concurrency. So I had to scrap this simple appraoch. Also the consumer should not be the producer from what I understand.
The scheduler solves this through separation of concerns. Workers only execute (since it is the consumer). The scheduler owns the responsibility of feeding tasks into the task queue from both the waiting queue and the retry queue. If the task queue is full, the scheduler waits — workers remain free to continue processing.
## Backpressure
Backpressure is the signal a consumer sends to a producer to slow down when it cannot keep up (basically means the worker was too slow to complete the task). This system implements backpressure at three stages:

Waiting queue full — new task submissions are rejected immediately with a clear error. For a human-facing system, instant rejection is preferable to silent indefinite waiting. (This was designed under the assumption that the engine will be used by humans. If it was supposed to be used by bots, then tasks would not be dropped but rather would be waiiting till it gets into the wait queue).\
Task queue full — tasks wait in the waiting queue until space becomes available. This is a blocking send.\
Retry queue full — workers block until the retry queue has space. This is the original deadlock scenario that motivated the scheduler separation.

## Graceful Shutdown
Graceful shutdown ensures workers complete their current task before exiting, so no task is left in running state. When Ctrl+C is received:

New task submissions stop\
Tasks in the waiting queue and task queue are discarded — the interrupt is intentional\
Workers finish their current task and exit\
Metrics are printed

Shutdown ordering matters. Closing channels in the wrong order causes panics. The correct sequence is:

Close waitingQueue — no new tasks enter the system\
Close retryQueue — no more retries on normal path\
Wait for scheduler to exit — ensures nothing is still writing to taskQueue. So no new tasks can be picked.\
Close taskQueue — now safe, no writers remain\
Wait for workers to exit

A key challenge was that tasks sitting in queues when interrupt fires never reach taskWg.Done(). Main listens for the interrupt signal directly to avoid blocking forever on taskWg.Wait().\
## Metrics
Collected at runtime, printed at exit:

Total tasks submitted\
Completed tasks\
Dropped tasks (submission rejections + retry exhaustions(considered as failed) tracked separately)\
Failed tasks\
Retry count and retry rate\
Drop rate\
Average execution latency (time spent actually executing)\
Average total latency (time from submission to completion, includes queue wait time)

The difference between execution latency and total latency reflects Little's Law in practice. From what I understand, the law is essentially that, with a large queue (I mean by size) and fixed worker counts, tasks spend the majority of their lifetime waiting, not executing which increases the total latency.
## Limitations and Future Work

Task types — current tasks are arithmetic operations for testing. Real use cases would involve I/O, network calls, or data processing. This will be done in the future when I have a task I wanna complete concurrently\
Input mechanisms — tasks are created programmatically in a loop. Future versions would accept tasks via REST API, gRPC, or CLI using Cobra.\
Dynamic capacity — queue sizes and worker counts are fixed at startup. A future version would support dynamic adjustment via CLI commands based on observed load.\
Testing — no automated tests exist. A concurrent system like this particularly needs integration tests that verify task counts, shutdown correctness, and race-free operation under load. The reason for no tests being me not knowing any testing framework as of today (05.08.26)\
Observability — metrics are printed once at exit. A real system would expose them continuously via an HTTP endpoint or metrics system.
