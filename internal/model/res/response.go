package res

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"plane_war/internal/utils"
)

const (
	Success = 0
	Error   = 1
)

type Response struct {
	Code int    `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}
type ListResponse[T any] struct {
	Count int64 `json:"count"`
	List  T     `json:"list"`
}

func Result(code int, data any, msg string, c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}
func Ok(data any, msg string, c *gin.Context) {
	Result(Success, data, msg, c)
}
func OkWithData(data any, c *gin.Context) {
	Result(Success, data, "成功", c)
}
func OkWithList(lst any, count int64, c *gin.Context) {
	OkWithData(ListResponse[any]{
		List:  lst,
		Count: count,
	}, c)
}
func OkWithMsg(msg string, c *gin.Context) {
	Result(Success, map[string]any{}, msg, c)
}
func OkWith(c *gin.Context) {
	Result(Success, map[string]any{}, "成功！", c)
}
func Fail(data any, msg string, c *gin.Context) {
	Result(Error, data, msg, c)
}
func FailWithMsg(msg string, c *gin.Context) {
	Result(Error, map[string]any{}, msg, c)
}
func FailWithCode(code ErrorCode, c *gin.Context) {
	msg, ok := ErrorMap[ErrorCode(code)]
	if ok {
		Result(int(code), map[string]any{}, msg, c)
		return
	}
	Result(Error, map[string]any{}, "未知错误", c)
}
func FailWithError(err error, obj any, c *gin.Context) {
	msg := utils.GetValidMsg(err, obj)
	FailWithMsg(msg, c)
}
