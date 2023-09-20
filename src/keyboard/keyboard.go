// keyboard包的描述
// keyboard提供一个函数，通过键盘输入得到一个浮点数，已经将数据的类型由string转换为float32
package keyboard

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Getfloat函数说明
// 入参：NULL
// 出参：float，error
func Getfloat() (float64, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	input = strings.TrimSpace(input)
	number, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}
