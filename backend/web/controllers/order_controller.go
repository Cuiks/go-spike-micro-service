package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"imoc-product/common"
	"imoc-product/datamodels"
	"imoc-product/services"
	"strconv"
)

type OrderController struct {
	Ctx          iris.Context
	OrderService services.IOrderService
}

func (o *OrderController) GetAll() mvc.View {
	orderArray, err := o.OrderService.GetAllOrderInfo()
	if err != nil {
		o.Ctx.Application().Logger().Debug("查询订单信息失败！")
	}
	return mvc.View{
		Name: "order/view.html",
		Data: iris.Map{
			"order": orderArray,
		},
	}
}

func (o *OrderController) GetManager() mvc.View {
	idString := o.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 16)
	if err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}
	order, err := o.OrderService.GetOrderByID(id)
	if err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}

	return mvc.View{
		Name: "order/manager.html",
		Data: iris.Map{
			"order": order,
		},
	}
}

func (o *OrderController) PostUpdate() {
	order := &datamodels.Order{}
	err := o.Ctx.Request().ParseForm()
	if err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "imooc"})
	if err = dec.Decode(o.Ctx.Request().Form, order); err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}
	err = o.OrderService.UpdateOrder(order)
	if err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}
	o.Ctx.Redirect("/order/all")
}

func (o *OrderController) GetDelete() {
	idString := o.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		o.Ctx.Application().Logger().Debug(err)
	}
	isOk := o.OrderService.DeleteOrderByID(id)
	if isOk {
		o.Ctx.Application().Logger().Debug("删除订单成功！")
	} else {
		o.Ctx.Application().Logger().Debug("删除订单失败！")
	}

	o.Ctx.Redirect("/order/all")
}
