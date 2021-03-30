package repositories

import (
	"database/sql"
	"errors"
	"imoc-product/common"
	"imoc-product/datamodels"
	"strconv"
)

type IUserRepository interface {
	Conn() (err error)
	Select(userName string) (user *datamodels.User, err error)
	Insert(user *datamodels.User) (userId int64, err error)
}

type UserManagerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func NewUserManagerRepository(table string, db *sql.DB) IUserRepository {
	return &UserManagerRepository{table: table, mysqlConn: db}
}

func (u *UserManagerRepository) Conn() (err error) {
	if u.mysqlConn == nil {
		mysql, errMysql := common.NewMysqlConn()
		if errMysql != nil {
			return errMysql
		}
		u.mysqlConn = mysql
	}
	if u.table == "" {
		u.table = "user"
	}
	return
}

func (u *UserManagerRepository) Select(userName string) (user *datamodels.User, err error) {
	if userName == "" {
		return &datamodels.User{}, errors.New("条件不能为空！")
	}
	if err := u.Conn(); err != nil {
		return &datamodels.User{}, err
	}

	sql := "SELECT * FROM " + u.table + " WHERE userName=?"
	row, errRow := u.mysqlConn.Query(sql, userName)
	if errRow != nil {
		return &datamodels.User{}, errRow
	}
	defer row.Close()

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在！")
	}

	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return
}

func (u *UserManagerRepository) Insert(user *datamodels.User) (userId int64, err error) {
	if err := u.Conn(); err != nil {
		return userId, err
	}

	sql := "INSERT " + u.table + " SET nickName=?, userName=?, passWord=?"
	stmt, err := u.mysqlConn.Prepare(sql)
	if err != nil {
		return userId, err
	}

	result, err := stmt.Exec(user.NickName, user.UserName, user.HashPassword)
	if err != nil {
		return userId, err
	}
	return result.LastInsertId()
}

func (u *UserManagerRepository) SelectByID(userId int64) (user *datamodels.User, err error) {
	if err := u.Conn(); err != nil {
		return &datamodels.User{}, err
	}
	sql := "SELECT * FROM " + u.table + " WHERE ID=" + strconv.FormatInt(userId, 10)
	row, err := u.mysqlConn.Query(sql)
	if err != nil {
		return &datamodels.User{}, err
	}
	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在！")
	}

	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return

}
