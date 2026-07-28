package normalizer

import "strconv"

func strconvFormatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func strconvFormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
