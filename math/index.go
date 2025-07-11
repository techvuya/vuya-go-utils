package mathutils

import (
	"fmt"
	"strconv"
	"strings"
)

func ConvertStringToUint64(numberText string) (number uint64) {
	number, err := strconv.ParseUint(numberText, 10, 64)
	if err != nil {
		return 0
	}
	return number
}

func ConvertStringToFloat64(numberText string) (float64, error) {
	number, err := strconv.ParseFloat(numberText, 64)
	if err != nil {
		return 0, err
	}
	return number, nil
}

func ConvertUint64ToString(number uint64) string {
	numberString := strconv.FormatUint(number, 10)
	return numberString
}

func ConvertFloatToUint(floatTwoDecimals float64) uint64 {
	s := fmt.Sprintf("%.2f", floatTwoDecimals)
	c := strings.Replace(s, ".", "", -1)
	cents, _ := strconv.ParseUint(c, 10, 64)
	return cents
}

func ConvertFloatToStringRemoveDecimals(floatTwoDecimals float64) string {
	return strconv.FormatFloat(floatTwoDecimals, 'f', 0, 64)
}

// func paddingUint64(n uint64) uint64 {
// 	var p uint64 = 1
// 	for p < n {
// 		p *= 10
// 	}

// 	return p
// }
