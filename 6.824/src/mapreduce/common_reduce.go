package mapreduce

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTask int, // which reduce task this is
	outFile string, // write the output here
	nMap int,       // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// doReduce manages one reduce task: it should read the intermediate
	// files for the task, sort the intermediate key/value pairs by key,
	// call the user-defined reduce function (reduceF) for each key, and
	// write reduceF's output to disk.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTask) yields the file
	// name from map task m.
	//
	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
	// Your code here (Part I).
	//

	var kvs = make(map[string][]string)
	var keys []string

	imFiles := make([]*os.File, nMap)

	defer func() {
		for i := 0; i < nMap; i++ {
			imFiles[i].Close()
		}
	}()

	for i := 0; i < nMap; i++ {
		reduceFileName := reduceName(jobName, i, reduceTask)
		if reduceFile, err := os.Open(reduceFileName); err != nil {
			log.Printf("open reduce im name : %s fail", reduceFileName)
			continue
		} else {
			var kv KeyValue
			decoder := json.NewDecoder(reduceFile)
			err := decoder.Decode(&kv)

			for err == nil { // read error until nil.
				// add keys kv.
				if _, ok := kvs[kv.Key]; !ok {
					keys = append(keys, kv.Key) // ensure keys is in key list.
				}
				kvs[kv.Key] = append(kvs[kv.Key], kv.Value)

				err = decoder.Decode(&kv)
			}
		}
	}

	// sort string type keys.
	sort.Strings(keys)
	outputFile, err := os.Create(outFile)
	if err != nil {
		log.Printf("create output file %s failed", outFile)
		return
	}

	defer func() {
		outputFile.Close()
	}()

	outputEncoder := json.NewEncoder(outputFile)
	for _, key := range keys {
		if err := outputEncoder.Encode(KeyValue{key, reduceF(key, kvs[key])}); err != nil {
			log.Printf("encode kvs fail key : %s , outFile : %s", key, outFile)
		}
	}
}
