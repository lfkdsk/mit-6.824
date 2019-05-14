package main

import (
	"fmt"
	"mapreduce"
	"os"
	"sort"

	//"sort"
	"strconv"
	"strings"
	"unicode"
)

//
// The map function is called once for each file of input. The first
// argument is the name of the input file, and the second is the
// file's complete contents. You should ignore the input file name,
// and look only at the contents argument. The return value is a slice
// of key/value pairs.
//
func mapF(filename string, contents string) []mapreduce.KeyValue {
	fmt.Printf("%s: map phase\n", filename)
	words := strings.FieldsFunc(contents, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	sort.Strings(words)

	kvs := make(map[string]int)
	var results []mapreduce.KeyValue

	for _, word := range words {
		kvs[word] = kvs[word] + 1
	}

	for k, v := range kvs {
		results = append(results, mapreduce.KeyValue{Key: k, Value: strconv.Itoa(v)})
	}

	//for _, word := range words {
	//	results = append(results, mapreduce.KeyValue{Key: word, Value: "1"})
	//}

	return results
}

//
// The reduce function is called once for each key generated by the
// map tasks, with a list of all the values created for that key by
// any map task.
//
func reduceF(key string, values []string) string {
	sum := 0
	for _, value := range values {
		intValue, _ := strconv.Atoi(value)
		sum += intValue
	}

	return strconv.Itoa(sum)
	//return strconv.Itoa(len(values))
}

// Can be run in 3 ways:
// 1) Sequential (e.g., go run wc.go master sequential x1.txt .. xN.txt)
// 2) Master (e.g., go run wc.go master localhost:7777 x1.txt .. xN.txt)
// 3) Worker (e.g., go run wc.go worker localhost:7777 localhost:7778 &)
func main() {
	if len(os.Args) < 4 {
		fmt.Printf("%s: see usage comments in file\n", os.Args[0])
	} else if os.Args[1] == "master" {
		var mr *mapreduce.Master
		if os.Args[2] == "sequential" {
			mr = mapreduce.Sequential("wcseq", os.Args[3:], 3, mapF, reduceF)
		} else {
			mr = mapreduce.Distributed("wcseq", os.Args[3:], 3, os.Args[2])
		}
		mr.Wait()
		// mr.CleanupFiles()
	} else {
		mapreduce.RunWorker(os.Args[2], os.Args[3], mapF, reduceF, 100, nil)
	}
}
