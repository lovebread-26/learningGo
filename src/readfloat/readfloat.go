//读取文本文件里的浮点数，一次一行

package readfloat

import (
	"bufio"
	"os"
	"strconv"
)

// 使用切片返回，避免数组大小不一致
func ReadFloatFromFile(fileName string) ([]float64, error) {
	var numbers []float64
	file, err := os.Open(fileName)
	if err != nil {
		//fmt.Println("open file", fileName, "error")
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	//i := 0
	for scanner.Scan() {
		number, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return nil, err
		}
		//i++
		numbers = append(numbers, number)
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return numbers, nil
}
