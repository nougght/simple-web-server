package util

import (
	"encoding/json"
	"fmt"
	"strings"
)

// преобразование списка в строку с разделителем
func SliceToCommaString(slice []string) string {
	return strings.Join(slice, ",")
}

func DecodeJson(rawData []byte, obj any) error {
	if err := json.Unmarshal(rawData, obj); err != nil {
		return fmt.Errorf("json decoding error: %s", err.Error())
	}
	return nil
}
