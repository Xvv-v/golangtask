//类型断言
package main

import(
	"fmt"
)

//打印函数
func print(s string){
	fmt.Println(s)
}

//adder 为包含加法的接口
type adder interface{
	add(int,int)int
}

//Number 为一个数值结构体
type Number struct{}

func (Number)add(a,b int)int{
	return a + b
}

func main(){

	//golang中所有类型都实现了interface{}
	//所以interface{}类型能接收所有类型结构
	var x interface{}

	x = "hello,world"	//此时是把string类型的字符串转化为了interface{}类型
	// print(x)			//传入只能接收字符串的函数时会报错因为类型不匹配
	s := x.(string)		//但是x里面存的就是一个字符串，所以可以通过类型断言转化为string类型
	print(s)			//正常输出

	x = Number{}
	var w adder
	var ok bool

	//Number 实现了adder 也实现了interface{}
	//所以x能类型断言为adder类型
	//w也能断言interface{}
	w,ok = x.(adder)
	fmt.Println(w,ok)  			//{} true
	x,ok= w.(interface{})		
	fmt.Println(x,ok)			//{} true
}

