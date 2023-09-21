package main

import "fmt"

type Liter float64
type Gallen float64
type MilLiter float64

func liter2Gallen(l Liter) Gallen {
	return Gallen(l * 0.264)
}

func gallen2Liter(g Gallen) Liter {
	return Liter(g * 3.785)
}

func (l Liter) ToGallen() Gallen {
	return Gallen(l * 0.264)
}

func (m MilLiter) ToGallen() Gallen {
	return Gallen(m * 0.000264)
}

func main() {
	var carBuf Liter
	var busBuf Gallen
	moto := MilLiter(1130.4321)

	carBuf += gallen2Liter(20.1)
	busBuf += liter2Gallen(20.2)

	fmt.Println("carbuf:", carBuf, "busbuf", busBuf, "moto:", moto.ToGallen())
}
