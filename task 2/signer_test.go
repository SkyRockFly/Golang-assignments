package main

import "testing"

func BenchmarkSingleHashConcat(b *testing.B) {
	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	input := make(chan any, 7)
	output := make(chan any, 7)

	for val := range inputData {
		input <- val
	}
	close(input)
	SingleHashConcat(input, output)

	for range inputData {
		<-output

	}

	close(output)

}

func BenchmarkSingleHashBuilder(b *testing.B) {

	inputData := []int{0, 1, 1, 2, 3, 5, 8}

	input := make(chan any, 7)

	output := make(chan any, 7)

	for val := range inputData {
		input <- val
	}
	close(input)
	SingleHashBuilder(input, output)

	for range inputData {
		<-output

	}

	close(output)

}
