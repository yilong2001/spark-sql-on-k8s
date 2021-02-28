package common

import(
	//log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	restful "github.com/emicklei/go-restful"

	//"github.com/dgrijalva/jwt-go"

)

type RestResponseType struct {
	Code      int          `json:"code"`
	Message   string       `json:"msg"`
	MessageCN string       `json:"msgcn"`
	Body      RestResponseBody `json:"body,omitempty"`
}

type RestResponseBody struct {
	Obj     interface{}   `json:"obj,omitempty"`
	List     []interface{} `json:"list,omitempty"`
	PageNum  int           `json:"pageNumber,omitempty"`
	PageSize int           `json:"pageSize,omitempty"`
	Total    int           `json:"total,omitempty"`
}

func NewRestResponseType(code int, 
	message string, 
	messageCN string, 
	obj interface{}, 
	list []interface{}) RestResponseType {
	return RestResponseType{
		Code:      code,
		Message:   message,
		MessageCN: messageCN,
		Body: RestResponseBody {
			Obj: obj,
			List: list,
		},
	}
}

// ********** restful **********
//NewPostRestfulSuccessResponse: 201 一般用于 post 或 put 返回资源已创建 
func NewPostRestfulSuccessResponse(obj interface{}, 
	list []interface{}, 
	response *restful.Response) {
	response.WriteHeaderAndJson(201, 
		NewRestResponseType(201, "", "", obj, list), restful.MIME_JSON)
	return
}

//NewRestfulSuccessResponse : 返回 200
func NewRestfulSuccessResponse(obj interface{}, 
	list []interface{}, 
	response *restful.Response) {
	response.WriteHeaderAndJson(200, 
		NewRestResponseType(200, "", "", obj, list), restful.MIME_JSON)
	return
}

//NewRestfulSuccessMessageResponse : 返回 200
func NewRestfulSuccessMessageResponse(obj interface{}, 
	list []interface{}, 
	message, messageCN string, 
	response *restful.Response) {
	response.WriteHeaderAndJson(200, 
		NewRestResponseType(200, message, messageCN, obj, list), restful.MIME_JSON)
	return
}

//NewRestfulFaliResponse : 返回失败码
func NewRestfulFaliResponse(code int, message string, messageCN string, response *restful.Response) {
	response.WriteHeaderAndJson(code, 
		NewRestResponseType(code, message, messageCN, nil, nil), restful.MIME_JSON)
	return
}

// ********** gin **********
//NewPostGinSuccessResponse: 201 一般用于 post 或 put 返回资源已创建 
func NewPostGinSuccessResponse(obj interface{}, 
	list []interface{}, 
	ctx *gin.Context) {
	ctx.JSON(201, NewRestResponseType(201, "", "", obj, list))
	return
}

//NewGinSuccessResponse : 返回 200
func NewGinSuccessResponse(obj interface{}, 
	list []interface{}, 
	ctx *gin.Context) {
	ctx.JSON(200, NewRestResponseType(200, "", "", obj, list))
	return
}

//NewGinSuccessMessageResponse : 返回 200
func NewGinSuccessMessageResponse(obj interface{}, 
	list []interface{}, 
	message, messageCN string, 
	ctx *gin.Context) {
	ctx.JSON(200, NewRestResponseType(200, message, messageCN, obj, list))
	return
}

//NewGinFaliResponse : 返回失败码
func NewGinFaliResponse(code int, 
	message string, messageCN string, ctx *gin.Context) {
	ctx.JSON(code, NewRestResponseType(code, message, messageCN, nil, nil))
	return
}

