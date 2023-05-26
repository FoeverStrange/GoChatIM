package service

import (
	"ginchat/models"
	"html/template"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("test.html", "view/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "test")

	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "welcome",
	// })
}
func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("view/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")

	// c.JSON(http.StatusOK, gin.H{
	// 	"message": "welcome",
	// })
}

// 跳转进入到聊天界面
func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("view/chat/index.html",
		"view/chat/foot.html",
		"view/chat/head.html",
		"view/chat/tabmenu.html",
		"view/chat/concat.html",
		"view/chat/group.html",
		"view/chat/main.html",
		"view/chat/profile.html")
	if err != nil {
		panic(err)
	}
	userId, _ := strconv.Atoi(c.Query("userId"))
	token := c.Query("token")
	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token
	// fmt.Println("ToChat>>>>", user)
	ind.Execute(c.Writer, user)
}
func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
