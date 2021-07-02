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
	go Print(wg, ch1, ch2, true)
	go Print(wg, ch1, ch2, false)

	ch1 <- struct{}{}

	wg.Wait()
}

func Print(wg *sync.WaitGroup, ch1, ch2 chan struct{}, flag bool) {
	defer wg.Done()

	start, end := ch1, ch2
	if flag {
		start, end = ch2, ch1
	}

	for i := 1; i <= 100; i++ {
		<-start
		if flag && i%2 == 1 {
			fmt.Println(i)
		} else if !flag && i%2 == 0 {
			fmt.Println(i)
		}
		end <- struct{}{}
	}
}
