package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var DB *sql.DB

//对数据表中的内容进行结构体化
type users struct {
	id int
	username string
	password string
	SecurityQuestions string
    SecurityAnswers string
}

var u users

//创建一个sql.DB
func initDB()(err error) {
	MyUsers := "root:@tcp(127.0.0.1:3306)/My_users"
	DB, err :=sql.Open("mysql", MyUsers)
	if err != nil {
		log.Fatal(err)
	}
    //设置数据库最大连接数
	DB.SetConnMaxLifetime(10)

    //设置数据库最大闲置连接数
	DB.SetMaxIdleConns(10)

	//验证是否成功连接数据库
	err = DB.Ping()
	if  err != nil {
		fmt.Println("fail")
		return  err
	}else {
		fmt.Println("connect successfully")
		return nil
	}
}

//连接数据库参数
const (
	username = "root"
	password = "root"
	ip = "localhost"
	port = "3306.redis"
	DBNAME = "project"
)

//用户注册
func userRegister(context *gin.Context) {
	//获取用户名与密码
	userName, _ := context.GetPostForm("userName")
	password, _ := context.GetPostForm("password")
	//查询列表
	rows, err := DB.Query("select username from user ")
	if err != nil {
		fmt.Println(err)
	}

	//copy
	err = rows.Scan(&u.username)
	if err != nil {
		fmt.Println(err)
	}

	//如果有相同的数据 —> 用户名已被注册
	if userName == u.username {
		context.JSON(200,gin.H{
			"success":false,
			"msg":"用户名已被注册",
		})
	}else {
		//无相同数据，设置用户名，密码与密保
		//获取密保
		securityQ, _ :=context.GetPostForm("securityQ")
		securityA, _ :=context.GetPostForm("securityA")
		//插入用户信息
		newUser, err := DB.Exec("insert into user (username,password,SecurityQuestions,SecurityAnswers)values (?,?,?,?)",userName,password,securityQ,securityA)
		if err != nil {
			fmt.Println("执行失败")
		} else {
			//获取被影响的行数
			rows,_:= newUser.RowsAffected()
			if rows !=1 {
				context.JSON(200,gin.H{
					"success": false,
				})
			} else {
				context.JSON(200,gin.H{
					"success": true,
					"msg":"userName"+"你已经成功创建你自己啦，记住信息哦(⊙o⊙)",
				})
			}
		}
	}
}

//用户登录
func userLogin(context *gin.Context) {
	//获取用户名和密码
	userName, _:= context.GetPostForm("userName")
	password,_ := context.GetPostForm("password")

	rows, err := DB.Query("select username from user where username=userName")
	if err != nil {
		fmt.Println(err)
	}
	//copy
	err = rows.Scan(&u.username)
	if err != nil {
		fmt.Println(err)
	}

	//没找到用户
	if userName != u.username {
		context.JSON(200,gin.H{
			"success":false,
			"code":400,
			"msg":"初来乍到，先注册一个用户吧",
		})
	} else {
		yourpassword, _ := DB.Query("select password from user where username = userName")

		err = yourpassword.Scan(&u.password)
		if err != nil {
			fmt.Println(err)
		}
		//密码错误
		if password != u.password {
			context.JSON(200,gin.H{
				"success": false,
				"code":400,
				"msg":"(ㄒoㄒ)密码错了，看看能不能找回来",
			})
			//采用密保函数将找回密码并进行修改
			Security(context)
		} else {
			context.JSON(200,gin.H{
				"success": true,
				"code":200,
				"msg":"脑子不错！！！登录成功！",
			})
		}
	}
}

//Security : 找回密码并修改
func Security(context *gin.Context) {
	securityQ, err:= DB.Query("select SecurityQuestions from uer where username = userName")
	if err != nil {
		fmt.Println(err)
	} else {
		context.JSON(200,gin.H{
			"msg":securityQ,
		})
	}

	//输入密保答案
	yourA, _:= context.GetPostForm("SecurityA")
	//获取密保答案
	securityA, err:= DB.Query("select securityA from user where id=(select id where username=userName) ")
	if err != nil {
		fmt.Println(err)
	} else {
		err = securityA.Scan(&u.SecurityAnswers)
		if err != nil {
			fmt.Println(err)
		}

		//答案正确，返回旧密码并修改
		if yourA==u.SecurityAnswers {
			password, err:=DB.Query("select password from uesr where id=(select id from user username=userName)")
			if err != nil{
				fmt.Println(err)
			}else {
				context.JSON(200, gin.H{
					"msg":      "答案正确！",
					"password": password,
				})
				//修改密码
				userName, _ := context.GetPostForm("userName")
				newPassword, _:= context.GetPostForm("newpassword")
				results, err:=DB.Exec("update user set password='"+newPassword+"' where username='"+userName+"'")
				if err != nil {
					fmt.Println(err)
				}
				//检验被影响的行数
				rows, _:=results.RowsAffected()
				if rows == 1{  //更新成功
					context.JSON(200,gin.H{
						"success":true,
						"code":200,
						"msg":"密码更新成功，不要又忘了哦",
					})
				}else {       //更新失败
					context.JSON(200,gin.H{
						"success":false,
						"code":400,
						"msg":"出错了，更新失败",
					})
				}
			}
		}else {   //密保答案错误，无法找回密码
			context.JSON(200,gin.H{
				"success":false,
				"msg":"很遗憾，你可能无法找回密码了~~",
			})
		}
	}
}

func main() {
	engine := gin.Default()

	_, err := sql.Open("mysql", "root:root@/My_users")

	if err!= nil{
		log.Fatal(err)
	}

	routerGroup := engine.Group("/user")

	routerGroup.POST("/register", userRegister)
	routerGroup.POST("/login", userLogin)
	routerGroup.POST("/security",Security)

	engine.Run()
}