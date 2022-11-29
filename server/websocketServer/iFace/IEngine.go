package iFace

type IEngine interface {
	// OnConnect 连接事件
	OnConnect(fn func(IClient))
	// OnMessage 新消息
	OnMessage(fn func(IClient, []byte))
	// OnClose 关闭
	OnClose(fn func(IClient, int, string) error)
	// OnError 出错
	OnError(fn func(IClient, error))
	// OnDisconnect 断开连接
	OnDisconnect(fn func(IClient))
	// Broadcast 广播
	Broadcast(message Envelope) error
	// AddRoute 注入路由类
	AddRoute(route IRoute) IRoute
	// AddSpace 注册控制逻辑
	AddSpace(space interface{}, pattern ...string)
	// CallRoute 调用控制器方法
	CallRoute(pattern string, msg interface{}) (interface{}, error)
	// IsHubClose 消息中转服务是否关闭
	IsHubClose() bool

	Run()
}
