//接口的嵌套使用

package main

import(
	"fmt"
)

//SalaryCount 是一个计算工资的总接口
//包含有三种不同的计算工资方法的接口
//不同员工类型的工资结算方式不同
//接口名首字母大写被导包时可以被调用
type SalaryCount interface{
	ManagerSalary	//经理
	JobberSalary	//临时工
	SalesmanSalary	//销售人员
}

//ManagerSalary 是经理 工资方法接口
type ManagerSalary interface{
	managerSalary(float64)float64
}

//JobberSalary 是临时工 工资方法接口
type JobberSalary interface{
	jobberSalary(int,float64)float64

}

//SalesmanSalary 是销售人员 工资方法接口
type SalesmanSalary interface{
	salesmanSalary(float64,float64,int)float64

}

//Staff 是一个员工结构
type Staff struct{

}

//实现接口方法时  func(看这里是什么类型就是为哪个结构实现接口)
//在这里func (n Staff)是为Staff结构体实现接口,且是结构体对象
//通过对象n可以调用结构内的属性

//经理拿固定工资
func (n Staff)managerSalary(sal float64)float64{
	return sal
}

//临时工 工作小时数 * 每个小时的工资
func (n Staff)jobberSalary(hour int,sal float64)float64{
	return float64(hour) * sal			//运算操作必须要类型一致，不然导致精度丢失......
}

//销售人员 固定工资 + 每件产品提成 * 销售产品数
func (n Staff)salesmanSalary(sal float64,money float64,num int)float64{
	return sal + money * float64(num)
}
func main(){
	var sal SalaryCount
	
	//当为一个接口对象 new(type) 时，此接口对象就能调用 type 类型结构实现的接口方法
	// sal.managerSalary(6000) 调用错误
	sal = new(Staff)


	//内置接口实现的方法，能通过外接口来调用
	fmt.Println(sal.managerSalary(6000))			//6000
	fmt.Println(sal.jobberSalary(120,20))			//2400
	fmt.Println(sal.salesmanSalary(2000,200,10))	//4000
}