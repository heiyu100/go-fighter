package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	// 导入时使用 _ 表示只执行包的初始化逻辑，但不引入包名
	_ "github.com/go-sql-driver/mysql"
)

// 数据库配置
/*
常量定义格式：const name [type] = value
显式类型定义 const b string = "abc"
隐式类型定义 const b = "abc"

批量声明多个常量：
const (
    e  = 2.7182818
    pi = 3.1415926
)

*/
const (
	DBHost     = "localhost"
	DBPort     = 3306
	DBUser     = "root"
	DBPassword = "123456789"
	DBName     = "a32"
)

// Book 结构体

/*
结构体就是这些类型中的一种复合类型，结构体是由零个或多个任意类型的值聚合成的实体，每个值都可以称为结构体的成员。
结构体的定义格式如下：

	type 类型名 struct {
	    字段1 字段1类型
	    字段2 字段2类型
	    …
	}
*/
type Book struct {
	// ID 字段表示图书的唯一标识，使用 `json:"id"` 标签指定 JSON 序列化时的字段名
	ID int `json:"id"`
	// Title 字段表示图书的标题，使用 `json:"title"` 标签指定 JSON 序列化时的字段名
	Title string `json:"title"`
	// Author 字段表示图书的作者，使用 `json:"author"` 标签指定 JSON 序列化时的字段名
	Author string `json:"author"`
}

var db *sql.DB //连接池对象，*sql.DB 表示这个变量是一个指向 sql.DB 类型的指针

func init() {
	// 构建 MySQL 数据源名称
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", DBUser, DBPassword, DBHost, DBPort, DBName)

	// 使用 sql.Open 打开 MySQL 数据库连接，返回一个数据库连接对象
	var err error
	db, err = sql.Open("mysql", dbURL)
	if err != nil {
		// 如果打开连接过程中发生错误，使用 panic 终止程序，并打印错误信息
		panic(err.Error())
	}

	// 使用 db.Ping() 发送一个 Ping 数据包到数据库，检查数据库连接是否正常
	err = db.Ping()
	if err != nil {
		// 如果 Ping 失败，使用 panic 终止程序，并打印错误信息
		panic(err.Error())
	}

	// 调用 createTable 函数，用于创建数据库表
	createTable()
}

func main() {
	// 创建一个新的 Gin 实例
	r := gin.Default()

	// 路由定义
	r.GET("/books", GetBooks)
	r.GET("/books/:id", GetBook)
	r.POST("/books", CreateBook)
	r.PUT("/books/:id", UpdateBook)
	r.DELETE("/books/:id", DeleteBook)

	// 启动服务器
	r.Run(":8080")
}

func createTable() {
	// 使用 db.Exec 执行 SQL 语句，创建名为 "books" 的表
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			author VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		// 如果执行 SQL 语句过程中发生错误，使用 panic 终止程序，并打印错误信息
		panic(err.Error())
	}
}

// 获取所有图书
func GetBooks(c *gin.Context) {
	// 使用 db.Query 执行 SQL 查询语句，查询所有图书信息
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		// 如果查询过程中发生错误，返回 HTTP 500 Internal Server Error 响应，附带错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	// 创建一个 Book 类型的切片，用于存储查询到的图书信息
	var books []Book
	// 遍历查询结果集
	for rows.Next() {
		// 创建一个 Book 结构体变量，用于存储从数据库查询得到的每本图书的信息
		var book Book
		// 使用 rows.Scan 将查询结果扫描到 book 变量中
		err := rows.Scan(&book.ID, &book.Title, &book.Author)
		if err != nil {
			// 如果扫描过程中发生错误，返回 HTTP 500 Internal Server Error 响应，附带错误信息
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// 将查询到的图书信息追加到 books 切片中
		books = append(books, book)
	}
	// 返回 HTTP 200 OK 响应，附带所有查询到的图书信息
	c.JSON(http.StatusOK, books)
}

// 获取单本图书
func GetBook(c *gin.Context) {
	// 从 URL 参数中获取图书的 ID
	id := c.Param("id")
	// 创建一个 Book 结构体变量，用于存储从数据库查询得到的图书信息
	var book Book
	// 使用 db.QueryRow 执行 SQL 查询语句，查询指定 ID 的图书信息，并将结果扫描到 book 变量
	err := db.QueryRow("SELECT * FROM books WHERE id=?", id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		// 如果查询过程中发生错误，返回 HTTP 404 Not Found 响应，表示未找到对应的图书
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}
	// 如果查询成功，返回 HTTP 200 OK 响应，附带查询到的图书信息
	c.JSON(http.StatusOK, book)
}

func GetBook1(c *gin.Context) {

}

// 创建新的图书
func CreateBook(c *gin.Context) {
	// 创建一个 Book 结构体变量，用于存储从请求中解析得到的书籍信息
	var book Book
	// 通过 ShouldBindJSON 解析请求体中的 JSON 数据，并将结果存入 book 变量
	if err := c.ShouldBindJSON(&book); err != nil {
		// 如果解析失败，返回 HTTP 400 Bad Request 响应，并附带错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 使用数据库执行 SQL 插入语句，将书籍信息插入数据库
	result, err := db.Exec("INSERT INTO books (title, author) VALUES (?, ?)", book.Title, book.Author)
	if err != nil {
		// 如果插入过程中发生错误，返回 HTTP 500 Internal Server Error 响应，并附带错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 获取插入操作生成的自增主键值（书籍的 ID）
	id, _ := result.LastInsertId()
	// 将获取到的 ID 赋值给 Book 结构体中的 ID 字段
	book.ID = int(id)
	// 返回 HTTP 201 Created 响应，附带创建的书籍信息
	c.JSON(http.StatusCreated, book)
}

// 更新图书信息
func UpdateBook(c *gin.Context) {
	// 从 URL 参数中获取图书的 ID
	id := c.Param("id")

	// 创建一个 Book 结构体变量，用于存储从请求体中解析得到的更新后的图书信息
	var updatedBook Book
	// 使用 c.ShouldBindJSON 解析请求体中的 JSON 数据，并将结果存入 updatedBook 变量
	if err := c.ShouldBindJSON(&updatedBook); err != nil {
		// 如果解析失败，返回 HTTP 400 Bad Request 响应，并附带错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 使用 db.Exec 执行 SQL 更新语句，更新指定 ID 的图书信息
	_, err := db.Exec("UPDATE books SET title=?, author=? WHERE id=?", updatedBook.Title, updatedBook.Author, id)
	if err != nil {
		// 如果更新过程中发生错误，返回 HTTP 500 Internal Server Error 响应，并附带错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 返回 HTTP 200 OK 响应，附带更新后的图书信息
	c.JSON(http.StatusOK, updatedBook)
}

// 删除图书
func DeleteBook(c *gin.Context) {
	// 从 URL 参数中获取图书的 ID
	id := c.Param("id")
	// 使用 db.Exec 执行 SQL 删除语句，删除指定 ID 的图书信息
	_, err := db.Exec("DELETE FROM books WHERE id=?", id)
	if err != nil {
		// 如果删除过程中发生错误，返回 HTTP 500 Internal Server Error 响应，并附带错误信息
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 返回 HTTP 200 OK 响应，附带删除成功的消息
	c.JSON(http.StatusOK, gin.H{"message": "Book deleted"})
}
