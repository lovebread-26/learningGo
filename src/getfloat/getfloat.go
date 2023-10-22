package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func OpenFile(filename string) (*os.File, error) {
	fmt.Println("Opening", filename)
	return os.Open(filename)
}

func CloseFile(filename *os.File) {
	fmt.Println("Closeing", filename)
	filename.Close()
}

func GetFloats(filename string) ([]float64, error) {
	var nums []float64
	file, err := OpenFile(filename)
	if err != nil {
		return nil, err
	}

	// 保证函数会被执行，无论何种情况退出
	defer CloseFile(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		number, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return nil, err
		}
		nums = append(nums, number)
	}
	// CloseFile(file)

	return nums, nil
}

func DirInfo(filename string) error {
	// file, err := ioutil.ReadDir(filename)
	fmt.Println(filename)
	file, err := os.ReadDir(filename)
	if err != nil {
		// log.Fatal(err)
		return err
	}
	// panic("This is a test panic")

	for _, name := range file {
		if name.IsDir() {
			// fmt.Println(name.Name())
			// newFile := fmt.Sprintf("%s%s", "../", name.Name())
			// 将新的目录与原目录拼接起来
			newFile := filepath.Join(filename, name.Name())
			// fmt.Println("newFile:", newFile)
			err = DirInfo(newFile)
			if err != nil {
				return err
			}
		} else {
			// fmt.Println("It's file", name.Name())
		}
	}

	return nil
}

func main() {

	err := DirInfo(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// nums, err := GetFloats(os.Args[1])
	// var sum float64

	// if err != nil {
	// 	// fmt.Printf("err: %v\n", err)
	// 	log.Fatal(err)
	// }
	// fmt.Println("nums", nums)
	// for _, num := range nums {
	// 	sum += num
	// }
	// fmt.Println("sum is:", sum)
}
