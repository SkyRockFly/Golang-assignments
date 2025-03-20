package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

const singleHash_maxGoroutines = 20
const multiHash_maxGoroutines = 100

func SingleHash(in, out chan any) {
	var mu sync.Mutex
	var waiter sync.WaitGroup
	var singleHash strings.Builder

	for rawData := range in {

		var data string

		threads := make(chan struct{}, singleHash_maxGoroutines)

		switch t := rawData.(type) {
		case string:
			data = t
		case int:
			data = strconv.Itoa(t)
		default:
			log.Info().Interface("data:", data).
				Msg("Invalid data")
			continue
		}

		waiter.Add(1)

		go func(data string) {

			threads <- struct{}{}
			defer func() { <-threads }()

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

			singleHash.WriteString(crcHash)
			singleHash.WriteString("~")
			singleHash.WriteString(md5crcHash)

			log.Info().Str("SingleHash", singleHash.String()).Msg("SingleHash output")

			out <- singleHash.String()

			waiter.Done()

		}(data)

	}

	waiter.Wait()

}

func MultiHash(in, out chan any) {
	var wg, waiter sync.WaitGroup
	threads := make(chan struct{}, multiHash_maxGoroutines)
	for data := range in {
		waiter.Add(1)
		data := data.(string)

		go func(data string) {
			hashArr := make([]string, 6)

			threads <- struct{}{}
			defer func() { <-threads }()

			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func() {
					hashArr[i] = DataSignerCrc32(strconv.Itoa(i) + data)
					log.Info().Str("Input data", data).
						Int("hashArr iteration", i).
						Str("hasharr[i]", hashArr[i]).
						Msg("HashArr[i] iteration value")
					wg.Done()

				}()

			}

			wg.Wait()

			var multiHash string

			for _, output := range hashArr {

				multiHash += output

			}

			log.Info().Str("Input data", data).Str("MultiHash", multiHash).Msg("MultiHash value")

			out <- multiHash

			waiter.Done()

		}(data)

	}

	waiter.Wait()

}

func CombineResults(in, out chan any) {
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
	input := make(chan any)

	for _, worker := range jobs {
		output := make(chan any)

		wg.Add(1)
		go func(work job, in chan any, out chan any) {
			defer wg.Done()
			defer close(out)
			work(in, out)
		}(worker, input, output)

		input = output

	}

	wg.Wait()

}
