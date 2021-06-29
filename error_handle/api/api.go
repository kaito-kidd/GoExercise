package api

import (
	"errors"
	"log"

	"Week02/common"
	"Week02/service"
)

// recover func
func recoverFunc() {
	if err := recover(); err != nil {
		sendErrorResponse(err.(error))
	}
}

// api
func HandleQueryUser() {
	// recover保护 防止进程崩溃
	defer recoverFunc()

	id := 100
	user, err := service.QueryUser(id)
	// 有错误 统一走错误响应
	if err != nil {
		sendErrorResponse(err)
		return
	}
	// 无错误 正常响应
	sendResponse(user)
}

func sendResponse(data interface{}) {
	// 正常响应200和数据
	log.Printf("200 %v", data)
}

func sendErrorResponse(err error) {
	// 数据找不到错误 响应404
	if errors.Is(err, common.ErrNotFound) {
		log.Printf("404 %s", err)
		return
	}

	// 其他错误 统一记录日志和错误堆栈 响应500
	log.Printf("internal error: %+v", err)
	log.Printf("500 %s", err)
}
