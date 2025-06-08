package app

type Request struct {
	Input1    int    `json:"input1"`
	Input2    int    `json:"input2"`
	Operation string `json:"operation"`
}

type Response struct {
	Result int    `json:"result"`
	Error  string `json:"error,omitempty"`
}
