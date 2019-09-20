package main

import (
	"fmt"
	"net/http"
)

//func Notify(uri string, evCh chan<- *Event) error ;
func main() {
	evCh := make(chan *Event)
	fmt.Printf("hello world\n")
	go Notify("https://smee.io/ZHyCz7qkmd6znOOP", evCh)

	for ev := range evCh {
		fmt.Printf("%v %v %d", ev.Type, ev.Data, ev.DataLen)
		if ev.DataLen <= 2 {
			continue
		}
		if ev.Type == "" {
			ev.Type = "application/json"
		}
		_, err := http.Post("http://127.0.0.1:3000", ev.Type, ev.Data)
		if err != nil {
			fmt.Printf("#v", err)
		}
	}
}
