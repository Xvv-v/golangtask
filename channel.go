package main

import "fmt"

func main() {

	//  声明一个通道类型
	//var str1 chan string

	//  用make创建一个通道类型
	str1 := make(chan string)

	//  创建匿名goroutine
	go func() {

		fmt.Println("start goroutine")

		//  通过通道通知main的goroutine
		str1 <- "ok"

		fmt.Println("exit goroutine")

	}()

	fmt.Println("wait goroutine")

	//  等待匿名的goroutine
	<-str1

	fmt.Println("all done")
}
