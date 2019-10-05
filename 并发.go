package main

import (
	"fmt"
	"time"
)

func running() {

	var num int
	for {
		num++
		fmt.Println("tick:", num)

		//  延时一秒
		time.Sleep(time.Second)
	}
}

func main() {

	/*
		//  开启一个线程
		go running()

		var str string
		fmt.Scanln(&str)
		fmt.Println(str)*/

	//  用匿名函数创建上面的goroutine例子，匿名函数没有参数
	go func() {

		var num int
		for {
			num++
			fmt.Println("tick:", num)

			//  延时一秒
			time.Sleep(time.Second)
		}
	}()

	var str string
	fmt.Scanln(&str)
	fmt.Println(str)

}
