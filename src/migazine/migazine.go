package migazine

import "fmt"

type Subscriber struct {
	Name    string
	Rate    float64
	Active  bool
	Address HomeAddress
}

type HomeAddress struct {
	Street string
	City   string
	Code   string
}

func PrintSubInfo(s *Subscriber) {
	fmt.Println(s.Name, s.Rate, s.Active)
	PrintAddressInfo(&s.Address)
}

func PrintAddressInfo(s *HomeAddress) {
	fmt.Println(s.Street, s.City, s.Code)
}

func DefaultSubInfo(name string) *Subscriber {
	var s Subscriber

	s.Name = name
	s.Active = true
	s.Rate = 5.9

	return &s
}

func DefaultAddressInfo(street string, city string, code string) *HomeAddress {
	var s HomeAddress

	s.Street = street
	s.City = city
	s.Code = code

	return &s
}

// func main() {
// 	var sub1, sub2 *subscriber
// 	sub1 = defaultSubInfo("allen")
// 	sub2 = defaultSubInfo("bell")
// 	sub2.rate = 4.3

// 	printSubInfo(sub1)
// 	// printSubInfo(sub2)
// 	printSubInfo(sub2)
// }
