package main

import (
	"context"
	"testing"
	"time"
)

func TestSaveID(t *testing.T) {
	put, get := saveId()
	id := get()
	if id != 0 {
		t.Errorf("raw id was not zero, it was %d", id)
	}
	put(28)
	id = get()
	if id != 28 {
		t.Errorf("expected id of 28, got %d", id)
	}
	put(39)
	id = get()
	if id == 28 {
		t.Errorf("old id, 28, was still there, it did not update to 39")
	}
	if id != 39 {
		t.Errorf("got %d back instead of 39", id)
	}
}

func TestContextStore(t *testing.T) {
	putCancel, getCancel := contextStore()
	ctx, canFunc, ktutor := getCancel()
	if ctx != nil {
		t.Errorf("uninitiaized ctx was not nil")
	}
	if canFunc != nil {
		t.Errorf("uninitialized cancel functon was not nil")
	}
	if ktutor != false {
		t.Errorf("uninitialized ktutor variable was not false")
	}
	ctx, cancel := context.WithCancel(context.Background())
	putCancel(ctx, cancel, true)
	work := func(ctx context.Context, c chan string) {
		for {
			time.Sleep(2 * time.Millisecond)
			select {
			case <-ctx.Done():
				c <- "finished"
				return
			default:
				continue
			}
		}
	}
	finished := make(chan string)
	go work(ctx, finished)
	_, cancel2, ktutor2 := getCancel()
	if cancel2 == nil {
		t.Errorf("did not get the cancel functon back")
	}
	cancel2()
	f := <-finished
	if f != "finished" {
		t.Errorf("Cancel did not work")
	}
	if ktutor2 != true {
		t.Errorf("ktutor did not update to true")
	}

}
