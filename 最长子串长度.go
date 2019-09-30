package main

import "fmt"

func main() {

	//  定义变量，输入字符串
	var str string
	fmt.Print("请输入一串字符：")
	fmt.Scanln(&str)

	//  输入字符的长度(如果想要处理中文，把 byte 换成 rune)
	//  用 map 来判断是否由重复
	mymap := make(map[byte]int)
	start := 0
	maxLength := 0
	for i, value := range []byte(str) {

		//  如果该字符存在且大于start，strat往后移一位
		if lastIndex, ok := mymap[value]; ok && lastIndex >= start {

			start = mymap[value] + 1
		}

		//  修正最大长度
		if i-start+1 > maxLength {

			maxLength = i - start + 1
		}

		//  将值存入
		mymap[value] = i
	}

	//  输出最大长度
	fmt.Println("最大长度为", maxLength)
}
