package query_parser

import (
	"testing"
)

func BenchmarkQueryParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseQuery("person(id:1)[p1: id, name, age]-frindsWith->person(age: \"p1.age-5<p1.age<p1.age+5\")[friends: id, name, age]-posted->post(publishDate: \">@now-30000\")[wall: id, author, publishDate, title, description]=>")
	}
}
