package tools

import (
	"bytes"
	"encoding/json"
)

func JsonStringToAny(jsonStr string, out any) (err error) {
	d := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	d.UseNumber()
	err = d.Decode(&out)
	return
}

// StructToStruct 从目标struct装载数据进另一个struct
func StructToStruct(from, to interface{}) error {
	str, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return JsonStringToAny(string(str), to)
}

func Struct2Map(from any) (to map[string]any) {
	str, _ := json.Marshal(from)
	JsonStringToAny(string(str), &to)
	return
}
