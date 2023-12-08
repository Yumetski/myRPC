package client

// Call 记录一次RPC调用所需要的信息
type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call // 调用结束通知对方
}

func (call *Call) done() {
	call.Done <- call
}
