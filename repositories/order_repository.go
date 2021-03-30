package repositories

import (
	"database/sql"
	"imoc-product/common"
	"imoc-product/datamodels"
	"strconv"
)

type IOrderRepository interface {
	Conn() error
	Insert(*datamodels.Order) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Order) error
	SelectByKey(int64) (*datamodels.Order, error)
	SelectAll() ([]*datamodels.Order, error)
	SelectAllWithInfo() (map[int]map[string]string, error)
}

type OrderMangerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func NewOrderManagerRepository(table string, mysqlConn *sql.DB) IOrderRepository {
	return &OrderMangerRepository{table: table, mysqlConn: mysqlConn}
}

func (o *OrderMangerRepository) Conn() error {
	if o.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		o.mysqlConn = mysql
	}
	if o.table == "" {
		o.table = "order_table"
	}
	return nil
}

func (o *OrderMangerRepository) Insert(order *datamodels.Order) (orderID int64, err error) {
	if err = o.Conn(); err != nil {
		return
	}

	sql := "INSERT " + o.table + " set userID=?, productID=?, orderStatus=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	result, err := stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	if err != nil {
		return
	}

	return result.LastInsertId()
}

func (o *OrderMangerRepository) Delete(productID int64) bool {
	if err := o.Conn(); err != nil {
		return false
	}

	sql := "DELETE from " + o.table + " WHERE ID=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}
	_, err = stmt.Exec(productID)
	if err != nil {
		return false
	}
	return true
}

func (o *OrderMangerRepository) Update(order *datamodels.Order) (err error) {
	if err = o.Conn(); err != nil {
		return
	}

	sql := "UPDATE " + o.table + " set userID=?, productID=?, orderStatus=? WHERE ID=" + strconv.FormatInt(order.ID, 10)
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	_, err = stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	return

}

func (o *OrderMangerRepository) SelectByKey(orderID int64) (orderResult *datamodels.Order, err error) {
	if err = o.Conn(); err != nil {
		return &datamodels.Order{}, err
	}

	sql := "SELECT * FROM " + o.table + " WHERE ID=" + strconv.FormatInt(orderID, 10)
	row, err := o.mysqlConn.Query(sql)
	if err != nil {
		return &datamodels.Order{}, err
	}
	defer row.Close()

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.Order{}, nil
	}

	orderResult = &datamodels.Order{}
	common.DataToStructByTagSql(result, orderResult)
	return
}

func (o *OrderMangerRepository) SelectAll() (orderArray []*datamodels.Order, err error) {
	if err = o.Conn(); err != nil {
		return nil, err
	}

	sql := "SELECT * FROM " + o.table
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := common.GetResultRows(rows)
	if len(result) == 0 {
		return nil, nil
	}

	for _, v := range result {
		order := &datamodels.Order{}
		common.DataToStructByTagSql(v, order)
		orderArray = append(orderArray, order)
	}
	return
}

func (o *OrderMangerRepository) SelectAllWithInfo() (OrderMap map[int]map[string]string, err error) {
	if err = o.Conn(); err != nil {
		return nil, err
	}

	sql := "SELECT o.ID, p.productName, o.orderStatus From imooc.order_table as o left join product as p on o.productID=p.ID"
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return common.GetResultRows(rows), nil
}
