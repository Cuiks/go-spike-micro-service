package main

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"imoc-product/common"
	"imoc-product/fronted/middlerware"
	"imoc-product/fronted/web/controllers"
	"imoc-product/rabbitmq"
	"imoc-product/repositories"
	"imoc-product/services"
	"log"
)

func main() {
	// 1.创建iris实例
	app := iris.New()
	// 2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	// 3.注册模板
	template := iris.HTML("./fronted/web/views", ".html").Layout(
		"shared/layout.html").Reload(true)
	app.RegisterView(template)
	// 4.设置模板
	app.HandleDir("/public", "./fronted/web/public")
	// 访问生成好的html静态文件
	app.HandleDir("/html", "./fronted/web/htmlProductShow")
	// 出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})
	db, err := common.NewMysqlConn()
	if err != nil {
		log.Println(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user := repositories.NewUserManagerRepository("user", db)
	userService := services.NewUserService(user)
	userPro := mvc.New(app.Party("/user"))
	userPro.Register(userService, ctx)
	userPro.Handle(new(controllers.UserController))

	rabbitmq := rabbitmq.NewRabbitMQSimple("imoocProduct")

	// 注册product控制器
	productRepo := repositories.NewProductManager("product", db)
	productService := services.NewProductService(productRepo)
	orderRepo := repositories.NewOrderManagerRepository("order_table", db)
	orderService := services.NewOrderService(orderRepo)
	productParty := app.Party("/product")
	product := mvc.New(productParty)
	// 使用中间件
	productParty.Use(middlerware.AuthConProduct)
	product.Register(productService, orderService, ctx, rabbitmq)
	product.Handle(new(controllers.ProductController))

	app.Run(
		iris.Addr("0.0.0.0:8082"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
