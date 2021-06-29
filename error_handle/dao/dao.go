package dao

import (
	"database/sql"

	"Week02/common"

	"github.com/pkg/errors"
)

func QueryUser(id int) (interface{}, error) {
	r, err := queryUserFromDB()
	// 包装成自己的错误 屏蔽数据库差异
	if err == sql.ErrNoRows {
		return nil, errors.Wrapf(common.ErrNotFound, "user id: %d", id)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "query user error, user id: %d", id)
	}
	return r, nil
}

// 模拟抛出sql.ErrNoRows
func queryUserFromDB() (interface{}, error) {
	return nil, sql.ErrNoRows
}
