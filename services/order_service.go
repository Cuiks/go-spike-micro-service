package services

import (
	"imoc-product/datamodels"
	"imoc-product/repositories"
)

type IOrderService interface {
	GetOrderByID(int64) (*datamodels.Order, error)
	DeleteOrderByID(int64) bool
	UpdateOrder(*datamodels.Order) error
	InsertOrder(*datamodels.Order) (int64, error)
	GetAllOrder() ([]*datamodels.Order, error)
	GetAllOrderInfo() (map[int]map[string]string, error)
	InsertOrderByMessage(message *datamodels.Message) (int64, error)
}

func NewOrderService(repository repositories.IOrderRepository) IOrderService {
	return &OrderService{OrderRepository: repository}
}

type OrderService struct {
	OrderRepository repositories.IOrderRepository
}

func (o *OrderService) InsertOrderByMessage(message *datamodels.Message) (int64, error) {
	order := &datamodels.Order{
		UserId:      message.UserID,
		ProductId:   message.ProductID,
		OrderStatus: datamodels.OrderSuccess,
	}
	return o.OrderRepository.Insert(order)
}

func (o *OrderService) GetOrderByID(orderID int64) (order *datamodels.Order, err error) {
	return o.OrderRepository.SelectByKey(orderID)
}

func (o *OrderService) DeleteOrderByID(orderID int64) bool {
	return o.OrderRepository.Delete(orderID)
}

func (o *OrderService) UpdateOrder(order *datamodels.Order) error {
	return o.OrderRepository.Update(order)
}

func (o *OrderService) InsertOrder(order *datamodels.Order) (int64, error) {
	return o.OrderRepository.Insert(order)
}

func (o *OrderService) GetAllOrder() ([]*datamodels.Order, error) {
	return o.OrderRepository.SelectAll()
}

func (o *OrderService) GetAllOrderInfo() (map[int]map[string]string, error) {
	return o.OrderRepository.SelectAllWithInfo()
}
