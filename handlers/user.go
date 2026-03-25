// 这是user.go文件，用于处理用户注册登录逻辑
package handlers

import (
	"sync"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// sync.map作为存储中心
var users = sync.Map{}

// 用户结构体
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 进行子路由注册
func RegisterRoutes(r *gin.Engine) {
	user := r.Group("/user")
	{
		user.POST("/register", Register)
		user.POST("/login", Login)
	}
}

// 增加每个用户的每分钟限制登录次数的限制
var loginAttempts = sync.Map{}


// 注册需要邮箱和密码
func Register(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.JSON(400, gin.H{"error": "邮箱和密码不能为空"})
		return
	}

	// 查看邮箱中是否有已经有了email
	_,ok := users.Load(email)
	if ok {
		c.JSON(400, gin.H{"error": "邮箱已经存在"})
		return
	}

	// 密码做hash处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "密码hash处理失败"})
		return
	}

	// 如果邮箱不存在，则将用户信息存储到sync.map中
	users.Store(email, User{Email: email, Password: string(hashedPassword)})
	c.JSON(200, gin.H{"message": "注册成功"})

}


func Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.JSON(400, gin.H{"error": "邮箱和密码不能为空"})
		return
	}

	// 查看邮箱中是否有已经有了email
	value,ok := users.Load(email)
	if !ok {
		c.JSON(400, gin.H{"error": "邮箱不存在"})
		return
	}

	// 将密码做hash处理
	hashedPassword := value.(User).Password
	
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		c.JSON(400, gin.H{"error": "密码错误"})
		return
	}

	c.JSON(200, gin.H{"message": "登录成功"})
	
}

// 用户可以更新个人资料（昵称，头像，简介）
func UpdateProfile(c *gin.Context) {
	email := c.PostForm("email")
	nickname := c.PostForm("nickname")
	avatar := c.PostForm("avatar")
	description := c.PostForm("description")

	strings.TrimSpace(nickname) != "" {
		c.JSON(400, gin.H{"error": "昵称不能为空"})
		return
	}

	//昵称长度限制
	if len(nickname) > 10 {
		c.JSON(400, gin.H{"error": "昵称长度不能超过10个字符"})
		return
	}

	// 校验avatar是合法的url
	_, err := url.Parse(avatar)
	if err != nil {
		c.JSON(400, gin.H{"error": "头像不是合法的url"})
		return
	}

	// 更新用户资料
	users.Store(email, User{Email: email, Nickname: nickname, Avatar: avatar, Description: description})
	c.JSON(200, gin.H{"message": "更新成功"})
	return
	
}