package main

import "fmt"

// Go 语言的数组是值，其长度是其类型的一部分，作为函数参数时，是 值传递，函数中的修改对调用者不可见
func change1(nums [3]int) {
	nums[0] = 4
}

// 传递进来数组的内存地址，然后定义指针变量指向该地址，则会改变数组的值
func change2(nums *[3]int) {
	nums[0] = 5
}

// Go 语言中对数组的处理，一般采用 切片 的方式，切片包含对底层数组内容的引用，作为函数参数时，类似于 指针传递，函数中的修改对调用者可见
func change3(nums []int) {
	nums[0] = 6
}

func nameless() {
	// 定义一个匿名函数并将其赋值给变量add
	add := func(a, b int) int {
		return a + b
	}

	// 调用匿名函数
	result := add(3, 5)
	fmt.Println("3 + 5 =", result)

	// 在函数内部使用匿名函数
	multiply := func(x, y int) int {
		return x * y
	}

	product := multiply(4, 6)
	fmt.Println("4 * 6 =", product)

	// 将匿名函数作为参数传递给其他函数
	calculate := func(operation func(int, int) int, x, y int) int {
		return operation(x, y)
	}

	sum := calculate(add, 2, 8)
	fmt.Println("2 + 8 =", sum)

	// 也可以直接在函数调用中定义匿名函数
	difference := calculate(func(a, b int) int {
		return a - b
	}, 10, 4)
	fmt.Println("10 - 4 =", difference)
}

func main() {
	var nums1 = [3]int{1, 2, 3}
	var nums2 = []int{1, 2, 3}
	change1(nums1)
	fmt.Println(nums1) //  [1 2 3]
	change2(&nums1)
	fmt.Println(nums1) //  [5 2 3]
	change3(nums2)
	fmt.Println(nums2) //  [6 2 3]

	nameless()
}
