package main

import (
	"errors"
	"fmt"
	"log"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func (d *Date) SetYear(year int) error {
	if year < 1 {
		return errors.New("invalid year")
	}
	d.Year = year
	return nil
}

func (d *Date) SetMonth(month int) error {
	if month < 1 || month > 12 {
		return errors.New("invalid month")
	}
	d.Month = month
	return nil
}

func (d *Date) SetDay(day int) error {
	if day < 1 || day > 31 {
		return errors.New("invalid day")
	}
	d.Day = day
	return nil
}

func main() {
	day := Date{Year: 2023, Month: 9, Day: 21}
	var day1 Date
	day2 := Date{}

	//day1.SetYear(2023)
	err1 := day1.SetYear(2023)
	if err1 != nil {
		log.Fatal(err1)
	}

	err2 := day1.SetMonth(9)
	if err2 != nil {
		log.Fatal(err2)
	}

	err3 := day1.SetDay(31)
	if err3 != nil {
		log.Fatal(err3)
	}

	fmt.Println("day", day)
	fmt.Println("day1", day1)
	fmt.Println("day2", day2)
}
