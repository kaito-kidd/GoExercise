**我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？**

1、预先定义一个的 `sentinel error`：

```go
package common

import "github.com/pkg/errors"

var ErrNotFound = errors.New("not found")
```

2、`DAO` 层：

- `sql.ErrNoRows` 错误需要进行转换，转换为自己预先定义 `sentinel error`，使用 `errors.Warp` 包装，并填充自己的上下文（例如查询参数），这么做的目的是屏蔽底层数据库的差异
- 非 `sql.ErrNoRows` 的其他错误，不需要做任何转换，直接使用 `errors.Wrap` 包装，填充上下文信息（查询参数），返回给调用者

代码如下：

```go
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
```

3、`service` 层：

处理具体的业务逻辑，然后直接调用 `DAO` 层拿到数据，如果有错误，直接返回，无需 `wrap` 错误。

代码如下：

```go
package service

import (
    "Week02/dao"
)

func QueryUser(id int) (interface{}, error) {
    // 处理自己的业务逻辑
    return dao.QueryUser(id)
}
```

4、`API` 层：

- 调用 `service` 层，拿到业务数据或错误结果
- 如果 `service` 没有返回 `error`，向客户端响应 `200` 状态码 和业务数据
- 如果 `service` 层返回了 `error`，统一走返回错误响应的逻辑
    - 如果 `error` 是自己包装的 `sentinel error`，例如 `common.ErrNotFound`，则响应 `404` 错误码
    - 如果是其他 `error`，在这里统一打印错误堆栈，并给客户端响应 `500` 错误码和错误信息

代码如下：

```go
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
    // 自定义错误 响应404
    if errors.Is(err, common.ErrNotFound) {
        log.Printf("404 %s", err)
        return
    }

    // 其他错误 统一记录日志和错误堆栈 响应500
    log.Printf("internal error: %+v", err)
    log.Printf("500 %s", err)
}
```
