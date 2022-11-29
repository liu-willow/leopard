package iFace

type IRoute interface {
	Register(space interface{}, pattern ...string)
	Call(pattern string, msg interface{}) (interface{}, error)
}
