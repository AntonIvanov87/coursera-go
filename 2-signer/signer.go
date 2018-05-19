package main

import (
	"strconv"
	"sort"
	"strings"
	"sync"
	"bytes"
)

func ExecutePipeline(jobs ...job) {
	var inChan chan interface{}
	var waitGroup sync.WaitGroup
	for _, jb := range jobs {
		outChan := make(chan interface{})
		waitGroup.Add(1)

		go func(jb job, inChan, outChan chan interface{}) {
			jb(inChan, outChan)
			close(outChan)
			waitGroup.Done()
		}(jb, inChan, outChan)

		inChan = outChan
	}

	waitGroup.Wait()
}

func SingleHash(in, out chan interface{}) {
	var waitGroup sync.WaitGroup
	for dataRaw := range in {
		dataInt, ok := dataRaw.(int)
		if !ok {
			panic("Can not convert data to int")
		}
		dataStr := strconv.Itoa(dataInt)
		waitGroup.Add(1)
		go func(data string) {
			defer waitGroup.Done()
			crc32Chan := callDataSignerCrc32(data)
			md5Chan := callDataSignerMd5(data)
			crc32Md5Chan := callDataSignerCrc32(<-md5Chan)
			result := <-crc32Chan + "~" + <-crc32Md5Chan
			out <- result
		}(dataStr)
	}
	waitGroup.Wait()
}

var dataSignerMutex = &sync.Mutex{}

func callDataSignerCrc32(data string) chan string {
	return callStrToStrFuncAsync(DataSignerCrc32, data)
}

func callDataSignerMd5(data string) chan string {
	return callStrToStrFuncAsync(
		func(data string) string {
			dataSignerMutex.Lock()
			defer dataSignerMutex.Unlock()
			return DataSignerMd5(data)
		},
		data)
}

func callStrToStrFuncAsync(stringToStringFunc func(string) string, data string) chan string {
	resultChan := make(chan string)
	go func() {
		resultChan <- stringToStringFunc(data)
		close(resultChan)
	}()
	return resultChan
}

func MultiHash(in, out chan interface{}) {
	var waitGroup sync.WaitGroup
	for dataRaw := range in {
		dataStr := rawToString(dataRaw)

		waitGroup.Add(1)
		go func(data string) {
			defer waitGroup.Done()
			crc32Chans := make([]chan string, 0, 5)
			for i := 0; i <= 5; i++ {
				crc32Chan := callDataSignerCrc32(strconv.Itoa(i)+data)
				crc32Chans = append(crc32Chans, crc32Chan)
			}

			var resultBuf bytes.Buffer
			for _, crc32Chan := range crc32Chans {
				resultBuf.WriteString(<-crc32Chan)
			}
			out <- resultBuf.String()
		}(dataStr)
	}
	waitGroup.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for dataRaw := range in {
		dataStr := rawToString(dataRaw)
		results = append(results, dataStr)
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}

func rawToString(raw interface{}) string {
	data, ok := raw.(string)
	if !ok {
		panic("Can not convert data to string")
	}
	return data
}
