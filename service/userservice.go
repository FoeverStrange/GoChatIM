package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [post]
func GetUserList(c *gin.Context) {
	// data := make([]*models.UserBasic, 10)
	data := models.GetUserList()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": data,
	})

}

// FindUserByNameAndPwd
// @Summary 登录
// @Tags 用户模块
// @param name formData string false "用户名"
// @param passWord formData string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	// user := models.UserBasic{}
	Name := c.Request.FormValue("name")
	PassWord := c.Request.FormValue("passWord")
	// fmt.Println(Name)
	user := models.FindUserByName(Name)
	if user.Name == "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户不存在",
			"data":    user,
		})
		return
	}
	flag := utils.ValidPassword(PassWord, user.Salt, user.PassWord)
	if !flag {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户密码错误",
			"data":    user,
		})
		return
	}
	user = models.FindUserByNameAndPwd(Name, PassWord)
	// user := models.FindUserByNameAndPwd(Name, PassWord)
	c.JSON(http.StatusOK, gin.H{
		"code":    0, //正确、成功,-1则为失败
		"message": "登录成功",
		"data":    user,
	})

}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name formData string false "用户名"
// @param passWord formData string false "密码"
// @param repassWord formData string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	passWord := c.Request.FormValue("passWord")
	if user.Name == "" || passWord == "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户名或密码不能为空！",
			"data":    user,
		})
		return
	}
	repassWord := c.Request.FormValue("repassWord")

	salt := fmt.Sprintf("%06d", rand.Int31())

	user_temp := models.FindUserByName(user.Name)
	if user_temp.Name != "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户名已注册",
			"data":    user,
		})
		return
	}
	if passWord != repassWord {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "两次密码不一致",
			"data":    user,
		})

	} else {
		// user.PassWord = passWord
		user.Salt = salt
		user.PassWord = utils.MakePassword(passWord, salt)
		models.CreateUser(user)
		c.JSON(200, gin.H{
			"code":    0,
			"message": "新增用户成功",
			"data":    user,
		})
	}
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id formData string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Request.FormValue("id"))
	user.ID = uint(id)
	result := models.DeleteUser(user)
	if result.Error != nil {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "删除操作错误",
			"data":    user,
		})

	} else {
		if result.RowsAffected == 0 {
			c.JSON(200, gin.H{
				"code":    -1,
				"message": "用户不存在",
				"data":    user,
			})

		} else {
			c.JSON(200, gin.H{
				"code":    0,
				"message": "删除用户成功",
				"data":    user,
			})

		}
	}
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Request.FormValue("id"))
	user.ID = uint(id)
	user.Name = c.Request.FormValue("name")
	user.PassWord = c.Request.FormValue("password")
	user.Email = c.Request.FormValue("email")
	user.Phone = c.Request.FormValue("phone")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户修改失败，参数不正确",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0,
			"message": "用户修改成功",
			"data":    user,
		})
	}

}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	//upGrade的CheckOrigin字段被
	//设置为一个回调函数。这个回调函数被用于验证WebSocket连接请求的来源，
	//防止跨域站点伪造请求。在这个例子中，回调函数返回了true，表示允许所有来源的请求连接。
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Upgrader是github.com/gorilla/websocket包中的结构体，
// 用于将HTTP请求升级为WebSocket连接。它提供了一些选项和回调函数来自定义连接的行为。

// 使用upGrade.Upgrade方法将HTTP连接升级为WebSocket连接。
// 如果升级成功，ws变量将持有对应的websocket.Conn对象，可以用于后续的消息传递和处理
func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
	}
	// utils.Publish(c, utils.PublishKey, "test")
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	fmt.Println("Going to handler msg")
	MsgHandler(ws, c)
}
func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	fmt.Println("msg subscribe: ", msg)
	if err != nil {
		fmt.Println(err)
	}
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	err = ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(m)
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	// id := c.Query("userId")
	users := models.SearchFriend(uint(id))
	// c.JSON(200, gin.H{
	// 	"code":    0,
	// 	"message": "查询好友列表成功",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}
