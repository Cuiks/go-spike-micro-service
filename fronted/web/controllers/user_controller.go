package controllers

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"imoc-product/datamodels"
	"imoc-product/encrypt"
	"imoc-product/services"
	"imoc-product/tool"
	"strconv"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.IUserService
	Session     *sessions.Session
}

func (c *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}

func (c *UserController) PostRegister() {
	var (
		nickName = c.Ctx.FormValue("nickName")
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)
	// ozzo-validation(表单认证)

	user := &datamodels.User{
		UserName:     userName,
		NickName:     nickName,
		HashPassword: password,
	}

	_, err := c.UserService.AddUser(user)
	if err != nil {
		c.Ctx.Application().Logger().Debug(err)
		c.Ctx.Redirect("/user/error")
		return
	}
	c.Ctx.Redirect("/user/login")
	return
}

func (c *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login.html",
	}
}

func (c *UserController) PostLogin() mvc.Response {
	// 1.获取用户提交的表单信息
	var (
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)
	// 2.验证账号密码
	user, isOk := c.UserService.IsPwdSuccess(userName, password)
	if !isOk {
		return mvc.Response{
			Path: "/user/login",
		}
	}
	// 3.写入用户id到cookie
	tool.GlobalCookie(c.Ctx, "uid", strconv.FormatInt(user.ID, 10))
	uidByte := []byte(strconv.FormatInt(user.ID, 10))
	uidString, err := encrypt.EnPwdCode(uidByte)
	if err != nil{
		fmt.Println(err)
	}
	// 写入用户浏览器
	tool.GlobalCookie(c.Ctx, "sign", uidString)

	return mvc.Response{
		Path: "/product/",
	}
}
