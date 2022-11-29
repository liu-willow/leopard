package iFace

type IHub interface {
	Run()
	Count() int
	All() map[string]IClient
}
