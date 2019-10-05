package main

import (
	"fmt"
	"reflect"
)

func test1() {

	fmt.Println("这是test1函数")
}

func test2() {

	fmt.Println("这是test2函数")
}

func test3() {

	fmt.Println("这是test3函数")
}

func test4() {

	fmt.Println("这是test4函数")
}

func add() int {

	return 10 + 20
}

func main() {

	//  输入函数
	var name string

	fmt.Println("输入函数名：")
	fmt.Scanln(&name)

	//  为map初始化
	var mymap map[string]interface{}
	mymap = map[string]interface{}{"test1": test1, "test2": test2, "test3": test3, "test4": test4, "add": add}

	//  遍历map
	for index, value := range mymap {

		if index == name {

			// 将函数包装为反射值对象
			funcValue := reflect.ValueOf(value)

			// 构造函数参数,传入空参
			initList := []reflect.Value{}

			// 反射调用函数
			tList := funcValue.Call(initList)

			//  如果有返回值输出返回值
			if tList != nil {

				// 获取第一个返回值, 取整数值
				fmt.Println(tList[0].Int())
			}

		}
	}

	//  就先写好了几个函数然后初始化好 map 实现了一下，不知道学长你是不是这意思
	//  然后没有找到怎么解决函数有参数的情况

}
