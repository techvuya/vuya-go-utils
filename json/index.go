package jsonutils

import (
	"encoding/json"
	"io"
)

func ConvertStructToJSONString(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func ConvertJSONStringToStruct(jsonString string, data interface{}) error {
	err := json.Unmarshal([]byte(jsonString), data)
	return err
}

func DecodeJSONBody(r io.Reader, d interface{}) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	errParse := json.Unmarshal(body, d)
	if errParse != nil {
		return err
	}
	return nil
}
