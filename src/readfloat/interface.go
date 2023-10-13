package readfloat

import "fmt"

//定义接口，包含三个方法：无入参和返回值；有入参float64和无返回值；无入参和有string返回值
type MyInterface interface {
	MethodWithoutParameters()
	MethodWithFloat(float64)
	MethodWithReturn() string
}

type Mytype int

//定义满足接口的方法
func (m Mytype) MethodWithoutParameters() {
	fmt.Println("Method without parameters")
}

func (m Mytype) MethodWithFloat(f float64) {
	fmt.Println("Method with float64", f)
}

func (m Mytype) MethodWithReturn() string {
	return "Method with return"
}

//定义一个不是interface的方法
func (m Mytype) MethodNotInterface() {
	fmt.Println("Method not interface")
}
