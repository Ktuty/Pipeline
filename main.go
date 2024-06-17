package main

import (
	"bufio"
	"fmt"
	"log"
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
	log.Printf("Creating new ring buffer of size %d\n", size)
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
	log.Printf("Pushed element %d to ring buffer\n", el)
}

func (rb *RingBuffer) Get() []int {
	if rb.position <= 0 {
		return nil
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()

	var output []int = rb.buffer[:rb.position+1]
	rb.position = -1

	log.Printf("Got elements from ring buffer: %v\n", output)
	return output
}

func read(nextStage chan<- int, done chan<- bool) {
	scanner := bufio.NewScanner(os.Stdin)
	var data string
	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			log.Println("Program is exiting")
			fmt.Println("Программа завершила работу!")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			log.Printf("Invalid number: %s\n", data)
			fmt.Println("Некорректное число")
			continue
		}
		nextStage <- i
		log.Printf("Read number %d from input\n", i)
	}
}

func negatives(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			if data > 0 {
				output <- data
				log.Printf("Sent positive number %d to output\n", data)
			} else {
				log.Printf("Ignored non-positive number %d\n", data)
			}
		case <-done:
			log.Println("Negatives function is exiting")
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
				log.Printf("Sent number %d divisible by 3 to output\n", data)
			} else {
				log.Printf("Ignored number %d not divisible by 3\n", data)
			}
		case <-done:
			log.Println("NotDividedThree function is exiting")
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
			log.Printf("Buffered number %d\n", data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			for _, data := range bufferData {
				output <- data
				log.Printf("Sent buffered number %d to output\n", data)
			}
		case <-done:
			log.Println("BufferStage function is exiting")
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
			log.Printf("Processed data: %d\n", data)
			fmt.Println("Обработанные данные: ", data)
		case <-done:
			log.Println("Main function is exiting")
			return
		}
	}
}
