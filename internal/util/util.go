package util

import "strings"

// преобразование списка в строку с разделителем
func SliceToCommaString(slice []string) string {
	return strings.Join(slice, ",")
}
