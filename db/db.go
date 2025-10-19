package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type User struct {
	Id       int
	Password string
	Name     string
}

// InitDB 初始化数据库
func InitDB() (err error) {
	dsn := "root:214451@tcp(127.0.0.1:3306)/tcp?charset=utf8mb4&parseTime=True"
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

// SelectAll 所有用户信息
func (U *User) SelectAll(db *sql.DB) ([]User, error) {
	sqlStr := "select * from user "

	ret, err := db.Query(sqlStr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer func(ret *sql.Rows) {
		err := ret.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ret)

	var users []User
	for ret.Next() {
		var u User
		err := ret.Scan(&u.Id, &u.Password, &u.Name)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		users = append(users, u)
		fmt.Println(u)
		fmt.Printf("id:%d name:%s,passwored:%s\n", u.Id, u.Name, u.Password)
	}

	return users, nil
}

// Insert 插入新用户
func (U *User) Insert(db *sql.DB) error {
	sqlStr := "insert into user (password,name) values (?,?)"
	ret, err := db.Exec(sqlStr, U.Password, U.Name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	theID, err := ret.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return err
	}
	fmt.Printf("insert success, the id is %d.\n", theID)
	return nil
}

// Exists 判断用户是否存在
func (U *User) Exists(db *sql.DB) (bool, error) {
	sqlStr := "SELECT COUNT(*) FROM user WHERE name = ?"
	var count int
	err := db.QueryRow(sqlStr, U.Name).Scan(&count)
	if err != nil {
		fmt.Println("check user exists error:", err)
		return false, err
	}

	return count > 0, nil
}

// Boolean 判断密码是否正确
func (U *User) Boolean(db *sql.DB) (bool, error) {
	var password string

	sqlStr := "select password from user where name = ?"
	ret, err := db.Query(sqlStr, U.Name)

	defer func(ret *sql.Rows) {
		err := ret.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ret)

	if err != nil {
		fmt.Println("check boolean error:", err)
		return false, err
	}

	if ret.Next() {
		// 将查询结果扫描到password变量
		if err := ret.Scan(&password); err != nil {
			return false, fmt.Errorf("提取密码失败: %v", err)
		}

	}
	return U.Password == password, nil
}
