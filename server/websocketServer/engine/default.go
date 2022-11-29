package engine

type (
	Request struct {
		Cmd  string      `json:"cmd"`
		Body interface{} `json:"body"`
		Ext  interface{} `json:"ext"`
	}

	Response struct {
		IsErr bool        `json:"isErr"`
		Cmd   string      `json:"cmd"`
		Body  interface{} `json:"body"`
		Ext   interface{} `json:"ext"`
	}

	Message struct {
		UUID string
		Body interface{}
	}
)
