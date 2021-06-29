package service

import (
	"Week02/dao"
)

func QueryUser(id int) (interface{}, error) {
	// 处理具体的业务逻辑
	return dao.QueryUser(id)
}
