package main

import (
	"fmt"
	"log"

	qp "github.com/amhester/charlotte/query"
)

func main() {
	test := "person(id:1)[p1: id, name, age]-frindsWith->person(age: \"p1.age-5<p1.age<p1.age+5\")[friends: id, name, age]-posted->post(publishDate: \">@now-30000\")[wall: id, author, publishDate, title, description]=>"
	q, err := qp.ParseQuery(test)
	if err != nil {
		log.Fatal(err)
	}
	next := q
	for next != nil {
		fmt.Printf("%v\n", next.ToString())
		next = next.Next
	}
}
