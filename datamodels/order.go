package datamodels

type Order struct {
	ID          int64 `json:"id" sql:"ID" imooc:"id"`
	UserId      int64 `json:"UserId" sql:"userID" imooc:"UserId"`
	ProductId   int64 `json:"ProductId" sql:"productID" imooc:"ProductId"`
	OrderStatus int64 `json:"OrderStatus" sql:"orderStatus" imooc:"OrderStatus"`
}

const (
	OrderWait = iota
	OrderSuccess
	OrderFailed
)
