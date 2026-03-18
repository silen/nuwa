package jsonutil

import (
	"bytes"
	"encoding/json"
)

func JsonStringToAny(jsonStr string, out any) (err error) {
	d := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	d.UseNumber()
	err = d.Decode(out)
	return
}

func Any2Any(from, to any) error {
	str, err := json.Marshal(from)
	if err != nil {
		return err
	}
	return JsonStringToAny(string(str), to)
}

func Any2Map(from any) (to map[string]any, err error) {
	str, err := json.Marshal(from)
	if err != nil {
		return nil, err
	}
	err = JsonStringToAny(string(str), &to)
	return
}
