package mapreduce

import (
	"fmt"
	"log"
	"sync"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
type CallReply struct {
	success bool
}

func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int
	var nOther int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		nOther = nReduce
	case reducePhase:
		ntasks = nReduce
		nOther = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, nOther)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//

	var wg sync.WaitGroup
	taskChan := make(chan int)

	go func() {
		for i := 0; i < ntasks; i++ {
			wg.Add(1)
			taskChan <- i
		}

		wg.Wait()
		close(taskChan)
	}()

	for taskId := range taskChan {
		worker, ok := <-registerChan
		if !ok {
			continue
		}

		mapFile := mapFiles[taskId]
		taskNumber := taskId

		var reply CallReply
		args := DoTaskArgs{
			JobName:       jobName,
			File:          mapFile,
			Phase:         phase,
			TaskNumber:    taskNumber,
			NumOtherPhase: nOther,
		}

		go func(worker string, args DoTaskArgs) {
			result := call(worker, "Worker.DoTask", args, &reply)

			if result {
				wg.Done()
				log.Printf("Schedule Dotask %v", taskNumber)
				registerChan <- worker // re-add (worker num is less than tasks, so add useless work to pool.)
			} else {
				log.Printf("Schedule Dotask fail  %v re-run", taskNumber)
				taskChan <- args.TaskNumber
			}
		}(worker, args)
	}

	fmt.Printf("Schedule: %v done\n", phase)
}
