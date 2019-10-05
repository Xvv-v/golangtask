//在并发中关于main的goroutine执行过快问题及部分解决方法

package main

import(
	"fmt"
	"time"
	"sync"
)

//print1 是一个打印函数
func print1(s string){
	fmt.Println(s)
}

//print2 是一个传通道变量的函数
func print2(c chan string){
	//data := <- c
	fmt.Println(<- c)
}

//
//main 也是一个goroutine 每开一个goroutine就相当于新开了一个固定代码的程序并运行
//每个goroutine在运行完时才会结束，而在main中创建goroutine后main的goroutine不会停止
//当main的goroutine运行结束时会强制结束其他没有运行完的goroutine
func main(){

	//创建print1()函数的goroutine，并加了延迟语句给了print1()的goroutine执行的时间，能输出
	go print1("I am time")
	time.Sleep(time.Second)

	//创建一个无缓冲的通道，通过通道共享信息时的堵塞机制来暂时阻断main的执行
	c := make(chan string)
	go print2(c)
	c <- "I am channel"	//没有接收通道c 里面的信息时，main会卡在这。

	//对于需要等待多个goroutine完成再进行下一步操作时
	var wg sync.WaitGroup
	//开 5个后台打印线程
	for i:=0;i<5;i++{
		wg.Add(1)	//代表需要等待的goroutine个数每次 +1
		go func(){
			fmt.Println("hello,world!")
			wg.Done()	//表示已经完成一个goroutine
		}()
	}
	//等待所有全部goroutine完成，没达到要求时会暂时堵塞main
	wg.Wait()

	//这个goroutine被创建后来不及执行完毕，main已经结束，所以不会输出 
	go print1("end of program")
}