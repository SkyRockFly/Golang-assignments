package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	var mu sync.Mutex
	var waiter sync.WaitGroup

	for rawData := range in {

		var data string

		switch t := rawData.(type) {
		case string:
			data = t
		case int:
			data = strconv.Itoa(t)
		default:
			fmt.Println("Invalid data")
			continue
		}

		waiter.Add(1)

		go func(data string) {

			var crcHash, md5crcHash string
			var wg sync.WaitGroup

			wg.Add(1)

			go func() {
				crcHash = DataSignerCrc32(data)
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				mu.Lock()
				tempHash := DataSignerMd5(data)
				mu.Unlock()
				md5crcHash = DataSignerCrc32(tempHash)

				wg.Done()

			}()

			wg.Wait()

			singleHash := crcHash + "~" + md5crcHash

			out <- singleHash

			waiter.Done()

		}(data)

	}

	waiter.Wait()

}

func MultiHash(in, out chan interface{}) {
	var wg, waiter sync.WaitGroup
	for data := range in {
		waiter.Add(1)
		data := data.(string)

		go func(data string) {
			hashArr := make([]string, 6)

			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(i int) {
					hashArr[i] = DataSignerCrc32(strconv.Itoa(i) + data)
					wg.Done()

				}(i)

			}

			wg.Wait()

			var multiHash string

			for _, output := range hashArr {

				multiHash += output

			}

			out <- multiHash

			waiter.Done()

		}(data)

	}

	waiter.Wait()

}

func CombineResults(in, out chan interface{}) {
	hashes := make([]string, 0)
	for data := range in {
		data := data.(string)
		hashes = append(hashes, data)
	}

	sort.Slice(hashes, func(i int, j int) bool {
		return hashes[i] < hashes[j]
	})

	var combinedResult string
	for i, hash := range hashes {
		if i == len(hashes)-1 {
			combinedResult += hash
		} else {
			combinedResult += hash + "_"
		}

	}

	out <- combinedResult

}

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup
	input := make(chan interface{})

	for _, worker := range jobs {
		output := make(chan interface{})

		wg.Add(1)
		go func(work job, in chan interface{}, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			work(in, out)
		}(worker, input, output)

		input = output

	}

	wg.Wait()

}

/*func main() {
	inputData := []int{0, 1}

	hashJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}

		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			fmt.Println(data)
		}),
	}

	ExecutePipeline(hashJobs...)

}
*/

// сюда писать код
