package repositories

import (
	"database/sql"
	"imoc-product/common"
	"imoc-product/datamodels"
	"strconv"
)

// 第一步，先开发接口
// 第二步，实现接口

type IProduct interface {
	// 链接数据库
	Conn() error
	Insert(*datamodels.Product) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Product) error
	SelectByKey(int64) (*datamodels.Product, error)
	SelectAll() ([]*datamodels.Product, error)
	SubProductNum(productID int64) error
}

type ProductManager struct {
	table     string
	mysqlConn *sql.DB
}

func (p *ProductManager) SubProductNum(productID int64) error {
	if err := p.Conn(); err != nil {
		return err
	}
	sql := "update " + p.table + " set " + " productNum=productNum-1 where ID = " + strconv.FormatInt(productID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func NewProductManager(table string, db *sql.DB) IProduct {
	return &ProductManager{table: table, mysqlConn: db}
}

// 数据库连接
func (p *ProductManager) Conn() (err error) {
	if p.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		p.mysqlConn = mysql
	}
	if p.table == "" {
		p.table = "product"
	}
	return
}

// 商品的插入
func (p *ProductManager) Insert(product *datamodels.Product) (productId int64, err error) {
	// 1.判断连接是否存在
	if err = p.Conn(); err != nil {
		return
	}
	// 2.准备sql
	sql := "INSERT product SET productName=?, productNum=?, productImage=?, productUrl=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return
	}
	// 3.传入参数
	result, err := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return
	}
	return result.LastInsertId()
}

// 商品的删除
func (p *ProductManager) Delete(productID int64) bool {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return false
	}
	// 2.准备sql
	sql := "DELETE from product WHERE ID=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}
	// 3.传入参数
	_, err = stmt.Exec(productID)
	if err != nil {
		return false
	}
	return true
}

// 商品的更新
func (p *ProductManager) Update(product *datamodels.Product) error {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return err
	}
	// 2.准备sql
	sql := "UPDATE product SET productName=?, productNum=?, productImage=?, productUrl=? " +
		"where ID=" + strconv.FormatInt(product.ID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}
	// 3.传入参数
	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return err
	}
	return nil
}

// 根据商品ID查询商品
func (p *ProductManager) SelectByKey(productID int64) (productResult *datamodels.Product, err error) {
	// 1.判断连接是否存在
	if err = p.Conn(); err != nil {
		return &datamodels.Product{}, err
	}
	// 2.准备sql
	sql := "SELECT * FROM " + p.table + " WHERE ID=" + strconv.FormatInt(productID, 10)
	row, err := p.mysqlConn.Query(sql)
	if err != nil {
		return &datamodels.Product{}, err
	}
	defer row.Close()
	// 3.传入参数
	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.Product{}, nil
	}
	productResult = &datamodels.Product{}
	common.DataToStructByTagSql(result, productResult)
	return
}

// 查询所有商品
func (p *ProductManager) SelectAll() (productArray []*datamodels.Product, err error) {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return nil, err
	}
	// 2.准备sql
	sql := "SELECT * FROM " + p.table
	rows, err := p.mysqlConn.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 3.传入参数
	results := common.GetResultRows(rows)
	if len(results) == 0 {
		return nil, nil
	}
	for _, v := range results {
		product := &datamodels.Product{}
		common.DataToStructByTagSql(v, product)
		productArray = append(productArray, product)
	}
	return
}
