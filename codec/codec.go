package codec

import "io"

/**
序列化
*/

type Header struct {
	ServiceMethod string //服务名和方法名
	Seq           uint64 //请求ID
	Error         string
}

// Codec 对消息进行编解码的接口
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.Reader) Codec
type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)

}
