package engine

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kataras/golog"

	"github.com/liu-willow/leopard/server/websocketServer/iFace"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

type (
	messageHandle func(iFace.IClient, []byte)
	errorHandle   func(iFace.IClient, error)
	closeHandle   func(iFace.IClient, int, string) error
	engineHandle  func(iFace.IClient)

	HandleFunc struct {
		Message    messageHandle
		MessageEnd messageHandle
		Error      errorHandle
		Close      closeHandle
		Connect    engineHandle
		Disconnect engineHandle
		Ping       engineHandle
	}

	Engine struct {
		addr     string
		Config   *iFace.Config
		upgrader *websocket.Upgrader
		handles  *HandleFunc
		hub      *hub
		logger   *golog.Logger
		route    iFace.IRoute
		clients  iFace.IClients
	}

	Options = struct {
		Config   *iFace.Config
		Upgrader *websocket.Upgrader
		HandleFunc
		Hub *hub
	}
)

var (
	dHeaders = map[string][]string{"server": {"leopard"}}
	dHub     = newHub(dClients)
	dConfig  = &iFace.Config{
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        (60 * time.Second * 9) / 10,
		MaxMessageSize:    512,
		MessageBufferSize: 256,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
	}
	leopard = &Engine{
		addr:     ":18765",
		Config:   dConfig,
		upgrader: dUpgrader,
		handles: &HandleFunc{
			Message:    func(client iFace.IClient, bytes []byte) {},
			MessageEnd: func(client iFace.IClient, bytes []byte) {},
			Error:      func(client iFace.IClient, err error) {},
			Close:      func(client iFace.IClient, i int, s string) error { return nil },
			Connect:    func(client iFace.IClient) {},
			Disconnect: func(client iFace.IClient) {},
			Ping:       func(client iFace.IClient) {},
		},
		logger:  golog.New(),
		hub:     dHub,
		route:   dRoute,
		clients: dClients,
	}
	dUpgrader = &websocket.Upgrader{
		ReadBufferSize:  dConfig.ReadBufferSize,
		WriteBufferSize: dConfig.WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func GetServer() *Engine {
	return leopard
}

func New() *Engine {
	return leopard
}

func (l *Engine) WithAddr(addr string) {
	l.addr = addr
}

func (l *Engine) WithConfig(config *iFace.Config) {
	l.Config = config
}

func (l *Engine) WithUpgrader(upgrade *websocket.Upgrader) {
	l.upgrader = upgrade
}

func (l *Engine) WithHandler(handle HandleFunc) {
	_t := reflect.TypeOf(handle)
	for k := 0; k < _t.NumField(); k++ {
		switch _t.Field(k).Name {
		case "Message":
			l.OnMessage(handle.Message)
			l.OnMessageBinary(handle.Message)
		case "MessageEnd":
			l.OnMessageEnd(handle.MessageEnd)
			l.OnMessageBinaryEnd(handle.MessageEnd)
		case "Error":
			l.OnError(handle.Error)
		case "Close":
			l.OnClose(handle.Close)
		case "Connect":
			l.OnConnect(handle.Connect)
		case "Disconnect":
			l.OnDisconnect(handle.Disconnect)
		case "Ping":
			l.OnPing(handle.Ping)
		}
	}
}

func (l *Engine) WithHub(hub *hub) {
	if l.IsHubClose() && hub != nil {
		l.hub = hub
	}
}

func (l *Engine) WithRouteRule(iRoute iFace.IRoute) iFace.IRoute {
	return l.AddSpaceRule(iRoute)
}

func (l *Engine) WithOptions(option Options) {
	_t := reflect.TypeOf(option)
	for k := 0; k < _t.NumField(); k++ {
		switch _t.Field(k).Name {
		case "Config":
			l.WithConfig(option.Config)
		case "Upgrader":
			l.WithUpgrader(option.Upgrader)
		case "HandleFunc":
			l.WithHandler(option.HandleFunc)
		case "Hub":
			l.WithHub(option.Hub)

		}
	}
}

// WithHeader 返回请求头
func (l *Engine) WithHeader(headers map[string][]string) {
	for k, v := range headers {
		dHeaders[k] = v
	}
}

/******************************************** 公用方法 可以直接调用, 也可以配置到WithXXX的参数里注入 start **************************************************************/

func (l *Engine) OnConnect(fn func(iFace.IClient)) {
	if fn != nil {
		l.handles.Connect = fn
	}
}

func (l *Engine) OnMessage(fn func(iFace.IClient, []byte)) {
	if fn != nil {
		l.handles.Message = fn
	}
}
func (l *Engine) OnMessageEnd(fn func(iFace.IClient, []byte)) {
	if fn != nil {
		l.handles.MessageEnd = fn
	}
}

func (l *Engine) OnMessageBinary(fn func(iFace.IClient, []byte)) {
	if fn != nil {
		l.handles.Message = fn
	}
}
func (l *Engine) OnMessageBinaryEnd(fn func(iFace.IClient, []byte)) {
	if fn != nil {
		l.handles.MessageEnd = fn
	}
}

func (l *Engine) OnPing(fn func(iFace.IClient)) {
	if fn != nil {
		l.handles.Ping = fn
	}
}

func (l *Engine) OnClose(fn func(iFace.IClient, int, string) error) {
	if fn != nil {
		l.handles.Close = fn
	}
}
func (l *Engine) OnError(fn func(iFace.IClient, error)) {
	if fn != nil {
		l.handles.Error = fn
	}
}
func (l *Engine) OnDisconnect(fn func(iFace.IClient)) {
	if fn != nil {
		l.handles.Disconnect = fn
	}
}

// Broadcast messageType: websocket.TextMessage/websocket.BinaryMessage
func (l *Engine) Broadcast(message iFace.Envelope) error {
	if l.hub.closed() {
		return ErrorHubClosed
	}

	l.hub.broadcast <- &message

	return nil
}

// AddSpaceRule 路由访问逻辑控制器
func (l *Engine) AddSpaceRule(route iFace.IRoute) iFace.IRoute {
	l.route = route
	return l.route
}

func (l *Engine) AddSpace(space interface{}, pattern ...string) {
	l.route.Register(space, pattern...)
}

func (l *Engine) CallSpace(pattern string, msg interface{}) (interface{}, error) {
	return l.route.Call(pattern, msg)
}

func (l *Engine) IsHubClose() bool {
	return l.hub.closed()
}

func (l *Engine) Logger() *golog.Logger {
	return l.logger
}

func (l *Engine) Total() int {
	return l.hub.clients.Count()
}

// OnHandShake 握手
func (l *Engine) onHandShake(w http.ResponseWriter, r *http.Request) error {
	if l.hub.closed() {
		return ErrorHubClosed
	}

	conn, err := l.upgrader.Upgrade(w, r, dHeaders)

	if err != nil {
		return err
	}
	client := &Client{
		uuid:       uuid.NewString(),
		context:    nil,
		conn:       conn,
		output:     make(chan *iFace.Envelope, l.Config.MessageBufferSize),
		outputDone: make(chan struct{}),
		leopard:    l,
		open:       true,
		mutex:      &sync.RWMutex{},
	}

	l.hub.register <- client

	l.handles.Connect(client)

	go client.WLoop()

	client.RLoop()

	if !l.hub.closed() {
		l.hub.unregister <- client
	}

	client.Close()

	l.handles.Disconnect(client)

	return nil
}

/******************************************** 公用方法 可以直接调用, 也可以配置到WithXXX的参数里注入 end **************************************************************/

func (l *Engine) Run() {
	go l.hub.Run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("当前在线: [%d]", l.Total())))
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { l.onHandShake(w, r) })
	l.Logger().Infof("%s server start %s", strings.Repeat("-", 30), strings.Repeat("-", 30))
	l.Logger().Infof("listen on: [%s]", l.addr)
	http.ListenAndServe(l.addr, nil)
}
