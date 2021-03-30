package controllers

import (
	"database/sql"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"imoc-product/repositories"
	"imoc-product/services"
)

type IBController interface {
	InitController(application *iris.Application, controllerName string)
	RegisterValue() string
	RegisterController()
}

type BController struct {
	Ctx   iris.Context
	Db    *sql.DB
	table string
	path  string
}

func (b *BController) InitController(application *iris.Application, controllerName string) {
	productRepository := repositories.NewProductManager(b.table, b.Db)
	productService := services.NewProductService(productRepository)
	productParty := application.Party(b.path)
	product := mvc.New(productParty)
	product.Register(b.Ctx, productService)
	product.Handle(new(ProductController))
}
