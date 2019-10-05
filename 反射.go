package main

import (
	"fmt"
	"reflect"
)

//Test  测试
type Test struct {
	name string
	age  int
}

func main() {

	//  创建结构体实例(相当于一个指针)
	test := Test{name: "小明", age: 18}

	//  获得结构体反射的类型实例
	typeOftest := reflect.TypeOf(test)

	//  遍历成员
	for i := 0; i < typeOftest.NumField(); i++ {

		//  获得每个成员的结构体字段类型
		typeOfstruct := typeOftest.Field(i)

		//  打印信息
		fmt.Println(typeOfstruct.Name, typeOfstruct.Tag)
	}

	//  通过字段名找到字段信息
}
