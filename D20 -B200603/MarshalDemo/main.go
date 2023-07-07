package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name    string
	Age     int
	address string
}

func (p Person) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FieldA string `json:"name1q11"`
		FieldB int    `json:"age"`
		FieldC string `json:"address"`
	}{
		FieldA: p.Name,
		FieldB: p.Age,
		FieldC: p.address,
	})
}

func main() {
	p := Person{Name: "John hai", Age: 20, address: "南昌"}
	jsonData, _ := json.Marshal(p)
	fmt.Println(string(jsonData))

	fmt.Println(p)

}
