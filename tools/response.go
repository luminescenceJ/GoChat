package tools

import "github.com/gin-gonic/gin"

// ResponseData 表示统一响应的JSON格式
type ResponseData struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// ErrorResponse 是一个辅助函数，用于创建错误响应
// 参数：
//
//	c *gin.Context：Gin上下文对象，用于处理HTTP请求和响应
//	code int：HTTP状态码，表示请求处理的结果
//	message string：响应消息，用于描述响应的错误信息或提示信息
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, ResponseData{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// SuccessResponse 是一个辅助函数，用于创建成功响应
// 参数：
//
//	c *gin.Context：Gin上下文对象，用于处理HTTP请求和响应
//	code int：HTTP状态码，表示请求处理的结果
//	data interface{}：响应数据，用于描述请求处理成功后返回的具体数据
func SuccessResponse(c *gin.Context, code int, data interface{}) {
	c.JSON(code, ResponseData{
		Code:    code,
		Message: "成功",
		Data:    data,
	})
}

//// UnifiedResponseMiddleware 是处理统一HTTP响应格式的中间件
//// 该中间件将在将响应发送给客户端之前拦截响应，并根据你指定的格式进行格式化。
//// 返回值：
////
////	gin.HandlerFunc：Gin中间件处理函数
//func UnifiedResponseMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		c.Next()
//
//		// 检查是否在处理请求时发生了错误
//		// 如果发生了错误，通过ErrorResponse函数创建一个错误响应，并返回给客户端
//		if len(c.Errors) > 0 {
//			err := c.Errors.Last()
//			ErrorResponse(c, http.StatusInternalServerError, err.Error())
//			return
//		}
//
//		// 检查是否设置了响应状态码
//		// 如果未设置响应状态码，默认将状态码设置为200（OK）
//		if c.Writer.Status() == 0 {
//			c.Writer.WriteHeader(http.StatusOK)
//		}
//
//		// 如果没有错误，则格式化响应
//		// 检查是否设置了"response_data"键的值，如果有，则调用SuccessResponse函数创建一个成功响应，并返回给客户端
//		if c.Writer.Status() >= http.StatusOK && c.Writer.Status() < http.StatusMultipleChoices {
//			data, exists := c.Get("response_data")
//			if exists {
//				SuccessResponse(c, c.Writer.Status(), data)
//				return
//			}
//		}
//	}
//}
