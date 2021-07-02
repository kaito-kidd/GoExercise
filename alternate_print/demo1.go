package main

import (
	"fmt"
	"sync"
)

// 2个协程 交替打印1,2,3,4...100
func main() {
	wg := &sync.WaitGroup{}

	ch1 := make(chan struct{}, 1)
	ch2 := make(chan struct{}, 1)

	wg.Add(2)
	go Print1(wg, ch1, ch2)
	go Print2(wg, ch1, ch2)

	ch1 <- struct{}{}

	wg.Wait()
}

func Print1(wg *sync.WaitGroup, ch1, ch2 chan struct{}) {
	defer wg.Done()

	for i := 1; i <= 100; i += 2 {
		<-ch1
		fmt.Println(i)
		ch2 <- struct{}{}
	}
}

func Print2(wg *sync.WaitGroup, ch1, ch2 chan struct{}) {
	defer wg.Done()

	for i := 2; i <= 100; i += 2 {
		<-ch2
		fmt.Println(i)
		ch1 <- struct{}{}
	}
}
