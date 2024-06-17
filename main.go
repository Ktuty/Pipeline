package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingBuffer struct {
	buffer   []int
	size     int
	position int
	mu       sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{make([]int, size), size, -1, sync.Mutex{}}
}

func (rb *RingBuffer) Push(el int) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.position == rb.size-1 {
		for i := 1; i <= rb.size-1; i++ {
			rb.buffer[i-1] = rb.buffer[i]
		}
		rb.buffer[rb.position] = el
	} else {
		rb.position++
		rb.buffer[rb.position] = el
	}
}

func (rb *RingBuffer) Get() []int {
	if rb.position <= 0 {
		return nil
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()

	var output []int = rb.buffer[:rb.position+1]
	rb.position = -1

	return output
}

func read(nextStage chan<- int, done chan<- bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			fmt.Println("Программа завершила работу!")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Некорректное число")
			continue
		}
		nextStage <- i
	}
}

func negatives(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			if data > 0 {
				output <- data
			}
		case <-done:
			return
		}
	}
}

func notDividedThreeFunc(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			if data%3 == 0 {
				output <- data
			}
		case <-done:
			return
		}
	}
}

func bufferStageFunc(input <-chan int, output chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingBuffer(size)
	for {
		select {
		case data := <-input:
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			for _, data := range bufferData {
				output <- data
			}
		case <-done:
			return
		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	go read(input, done)

	negativeFilterChannel := make(chan int)
	go negatives(input, negativeFilterChannel, done)

	notDividedChannel := make(chan int)
	go notDividedThreeFunc(negativeFilterChannel, notDividedChannel, done)

	bufferedIntChannel := make(chan int)
	bufferSize := 10
	bufferDrainInterval := 15 * time.Second
	go bufferStageFunc(notDividedChannel, bufferedIntChannel, done, bufferSize, bufferDrainInterval)

	for {
		select {
		case data := <-bufferedIntChannel:
			fmt.Println("Обработанные данные: ", data)
		case <-done:
			return
		}
	}
}
