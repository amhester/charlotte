package main

import (
	"fmt"
	"log"
	"time"

	qp "github.com/amhester/charlotte/query"
)

func main() {
	test := "person(id:1)[p1: id, name, age]-frindsWith->person(age: \"p1.age-5<p1.age<p1.age+5\")[friends: id, name, age]-posted->post(publishDate: \">@now-30000\")[wall: id, author, publishDate, title, description]=>"
	start := time.Now().UnixNano()
	q, err := qp.ParseQuery(test)
	end := time.Now().UnixNano()
	fmt.Printf("%v\n", end-start)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", q)
}
