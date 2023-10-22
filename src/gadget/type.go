package main

import (
	"fmt"
	"readfloat"
)

func PlayList(device readfloat.TapeInterface, songs []string) {
	for _, song := range songs {
		device.Play(song)
	}
	device.Stop()

	// 断言取回具体的类型
	recoder, ok := device.(readfloat.TapeRecorder)
	if ok {
		recoder.Record()
	} else {
		// fmt.Println("It is not a recoder", device)
		fmt.Printf("It is not a recoder %#v\n", device)
	}
}

func main() {
	player := readfloat.TapePlayer{Batteries: "full"}
	mixtype := []string{"aaaaa", "bbbbb", "ccccc"}
	PlayList(player, mixtype)

	recoder := readfloat.TapeRecorder{Microphone: 6}
	mixtype = []string{"ddddd", "eeeee", "fffff"}
	PlayList(recoder, mixtype)
	// recoder.Record()

	//var value readfloat.MyInterface
	// value := readfloat.Mytype(5)

	//调用接口
	// value.MethodWithoutParameters()
	// value.MethodWithFloat(123.456)
	// fmt.Println(value.MethodWithReturn())
	// value.MethodNotInterface()
	// 打印一个错误信息
	err := recoder.Error()
	fmt.Println(err)

	// 打印自定义的string
	str := readfloat.TapeRecorder{}
	fmt.Println(str)
}
