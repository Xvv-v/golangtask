package main

import "fmt"

//Dog  定义狗结构体
type Dog struct {
	dogName string
}

//Cat  定义猫结构体
type Cat struct {
	catName string
}

//Climb  定义爬接口
type Climb interface {
	climb()
}

//Run  定义跑接口
type Run interface {
	run()
}

//  实现爬接口
func (cat *Cat) climb() {

	fmt.Println(cat.catName, "在爬树")
}

//  猫实现跑接口
func (dog *Dog) run() {

	fmt.Println(dog.dogName, "再跑")
}

//  狗实现跑接口
func (cat *Cat) run() {

	fmt.Println(cat.catName, "再跑")
}
func main() {

	//  创建结构体实例到映射
	animals := map[string]interface{}{"猫咪": &Cat{catName: "咪咪"}, "狗狗": &Dog{dogName: "旺财"}}

	//  遍历map
	for name, value := range animals {

		//  判断是否实现了爬接口
		type1, ok1 := value.(Climb)

		//  判断是否实现了跑接口
		type2, ok2 := value.(Run)

		//  打印信息
		fmt.Printf("name: %s 实现了Climb: %v 实现了Run: %v\n", name, ok1, ok2)

		//  如果实现了Climb调用爬行
		if ok1 {

			type1.climb()
		}
		//  如果实现了Run调用跑
		if ok2 {

			type2.run()
		}

	}

}
