package bot

import "strconv"

func sumMapValues(m map[string]float64) float64 {
	var sum float64
	for _, value := range m {
		sum += value
	}
	return sum
}

func parseCategoryId(idStr string) (int64, error) {
	return strconv.ParseInt(idStr, 10, 64)
}
