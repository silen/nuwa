package conf

type (
	ReturnMap struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
		Data    any    `json:"data"`
	}
)
