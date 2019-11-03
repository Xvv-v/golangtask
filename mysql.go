package mydatabase

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"sync/atomic"

	"gopkg.in/yaml.v2"

	//"sync"

	_ "github.com/go-sql-driver/mysql"
)

//Config 配置
type Config struct {
	driveName string `yaml:"mysql"`
	network   string `yaml:"tcp"`
	ip        string `yaml:"127.0.0.1"`
	port      string `yaml:"3306"`
}

var config = Config{}

//解析yaml文件
func marshal() {

	buffer, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		log.Println("解析yaml文件出错:\n", err)
	}
}

//Pool 资源池
type Pool struct {
	closed   bool                      //是否关闭
	lock     sync.Mutex                //锁
	resource chan io.Closer            //通道，池中储存的资源
	factory  func() (io.Closer, error) //资源创建工厂函数
}

//New 创建资源池类 工厂函数
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {

	if size <= 0 {
		return nil, errors.New("size的值太小了")
	}
	return &Pool{
		resource: make(chan io.Closer, size),
		factory:  fn,
	}, nil
}

//获得资源
func (pool *Pool) acquireConn() (io.Closer, error) {

	select {
	case r, ok := <-pool.resource:
		if !ok {
			return nil, errors.New("资源池已经被关闭")
		}
		return r, nil
	default:
		log.Println("生成新资源")
		return pool.factory()
	}
}

//将一个使用完的资源放回池中
func (pool *Pool) releseConn(res io.Closer) {

	//保证操作是安全的
	//release和close用的是同一把锁，二者同一时间只能执行一个
	pool.lock.Lock()
	defer pool.lock.Unlock()

	//如果资源池关闭了，就把这个资源也关了，然后结束函数
	if pool.closed {
		res.Close()
		return
	}

	select {
	case pool.resource <- res:
		log.Println("资源已放到资源池")
	default:
		log.Println("资源池已满，资源释放")
		res.Close()
	}
}

//关闭资源池
func (pool *Pool) closePool() {

	pool.lock.Lock()
	defer pool.lock.Unlock()

	//如果已经关了就结束函数
	if pool.closed {
		return
	}
	//关闭通道
	close(pool.resource)

	//关闭通道里的资源
	for res := range pool.resource {
		res.Close()
	}

}

const maxOpen = 10

//Srouce 具体的资源类
type Srouce struct {
	id int32
}

//Close 实现io.Closer
func (sur *Srouce) Close() error {
	return nil
}

var idCount int32 //定义一个全局的共享的变量,更新时用原子函数锁住

func createConn() (io.Closer, error) {

	//原子函数锁住,更新加1
	id := atomic.AddInt32(&idCount, 1)
	log.Println("创建新资源:", id)
	return &Srouce{
		id: id,
	}, nil
}

//MySQL 连接池
type MySQL struct {
	db       *sql.DB
	pool     *Pool //资源池
	srou     *Srouce
	stmt     *sql.Stmt
	tx       *sql.Tx
	username string //数据库用户名
	password string //数据库密码
	connNum  int    //当前连接数
	close    bool   //关闭连接
}

//初始化
func (mysql *MySQL) openSQL(name, password string) {

	var err error
	//创建资源池
	mysql.pool, err = New(createConn, maxOpen)
	if err != nil {
		log.Println(err)
		return
	}
	err = nil
	dsn := name + ":" + password + "@" + config.network + "(" + config.ip + config.port + ")"
	mysql.db, err = sql.Open(config.driveName, dsn)
	if err != nil {
		log.Fatal(err)
	}
	err = mysql.db.Ping()
	if err != err {
		log.Fatal(err)
	}
}

//准备sql语句
func (mysql *MySQL) prepareSQL(sql string) (*sql.Stmt, error) {

	var err error
	mysql.stmt, err = mysql.db.Prepare(sql)
	if err != nil {
		log.Println("prepare SQL failed:", err)
		return nil, nil
	}
	return mysql.stmt, nil
}

//建表
func (mysql *MySQL) createTable(sql string) {
	if _, err := mysql.db.Exec(sql); err != nil {

		log.Println("create table failed:", err)
		return
	}
	log.Println("create table succeed!")
}

//更新数据
func (mysql *MySQL) updateSQL(sql string, args ...interface{}) {

	mysql.stmt, _ = mysql.prepareSQL(sql)
	defer mysql.stmt.Close()
	result, err := mysql.stmt.Exec(args)
	if err != nil {
		log.Println("uodate failed:", err)
		return
	}
	fmt.Println(result.LastInsertId())
}

//删除
func (mysql *MySQL) deleteSQL(sql string, args ...interface{}) {

	mysql.stmt, _ = mysql.prepareSQL(sql)
	defer mysql.stmt.Close()
	result, err := mysql.stmt.Exec(args)
	if err != nil {
		log.Println("delete failed:", err)
		return
	}
	id, resErr := result.LastInsertId()
	if resErr != nil {
		log.Println("Get LastInsertId failed:", resErr)
		return
	}
	fmt.Println("Affected rows:", id)
}

//插入
func (mysql *MySQL) insertSQL(sql string, args ...interface{}) {

	mysql.stmt, _ = mysql.prepareSQL(sql)
	defer mysql.stmt.Close()
	result, err := mysql.stmt.Exec(args)
	if err != nil {
		log.Println("update failed:", err)
		return
	}
	fmt.Println(result.LastInsertId())
}

//多行查询
func (mysql *MySQL) querySQL(sql string) *sql.Rows {

	conn, connErr := createConn()
	if connErr != nil {
		return nil
	}
	mysql.connNum++
	defer mysql.pool.releseConn(conn)
	rows, err := mysql.db.Query(sql)
	if err != nil {
		log.Println("Query failed:", err)
		return nil
	}

	mysql.pool.releseConn(conn)
	return rows
}

//单行查询
func (mysql *MySQL) queryRow(sql string, args ...interface{}) *sql.Row {

	conn, err := createConn()
	if err != nil {
		return nil
	}
	mysql.connNum++
	defer mysql.pool.releseConn(conn)
	row := mysql.db.QueryRow(sql, args)
	mysql.pool.releseConn(conn)
	return row
}

//开启一个事务
func (mysql *MySQL) affair(sql string) *sql.Tx {

	var txErr error
	mysql.tx, txErr = mysql.db.Begin()
	if txErr != nil {
		log.Println("affair begin failed:", txErr)
		return nil
	}
	return mysql.tx
}

//Commit 回滚
func (mysql *MySQL) Commit() error {

	err := mysql.tx.Commit()
	if err != err {
		return err
	}
	return nil
}

//回滚
func (mysql *MySQL) rollBack() error {

	err := mysql.tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}
