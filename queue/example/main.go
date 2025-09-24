package main

import (
	"fmt"

	"github.com/koepkeca/gsds/queue"
)

const (
	//The GO Race detector can manage a max of 8192 concurrent routines.
	//Read more about this here: https://golang.org/doc/articles/race_detector.html
	nbrOfRoutines = 7654 //The GO Race detector can manage a max of 8192 concurrent routines.
)

func main() {
	Q := queue.New()
	defer Q.Close()
	for i := 0; i < nbrOfRoutines; i++ {
		go func(j int) {
			Q.Enqueue(j)
		}(i)
	}
	fmt.Printf("%d elements", Q.Len())
	next := Q.Dequeue()
	for next != nil {
		fmt.Printf("%d\n", next.(int))
		next = Q.Dequeue()
	}
}
