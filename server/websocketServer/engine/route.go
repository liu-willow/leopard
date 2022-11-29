package engine

import (
	"errors"
	"github.com/liu-willow/leopard/server/websocketServer/iFace"
	"reflect"
	"strings"
)

type Route struct {
	rule map[string]func(message interface{}) (interface{}, error)
}

var _ iFace.IRoute = (*Route)(nil)

var dRoute = &Route{rule: make(map[string]func(message interface{}) (interface{}, error))}

func (r *Route) Register(space interface{}, prefix ...string) {
	if len(prefix) < 1 {
		prefix = []string{reflect.TypeOf(space).Elem().Name()}
	}
	v := reflect.ValueOf(space)
	t := v.Type()
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.IsExported() {
			index := prefix[0] + "-" + strings.ToLower(m.Name)
			r.rule[index] = v.Method(i).Interface().(func(message interface{}) (interface{}, error))
		}
	}
}
func (r *Route) Call(pattern string, msg interface{}) (interface{}, error) {
	fn, ok := r.rule[pattern]
	if !ok {
		return nil, errors.New("方法不存在")
	}
	return fn(msg)
}
