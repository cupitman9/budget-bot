package bot

import (
	"fmt"
	"strconv"
	"strings"
)

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

func getStats(incomeCategories, expenseCategories map[string]float64) string {
	totalIncome := sumMapValues(incomeCategories)
	totalExpense := sumMapValues(expenseCategories)
	netIncome := totalIncome - totalExpense

	var response strings.Builder
	response.WriteString("📊 *Статистика за период*\n\n")

	response.WriteString(fmt.Sprintf("💰 *Доход*: %.1f\n", totalIncome))
	for category, amount := range incomeCategories {
		response.WriteString(fmt.Sprintf("  - %s: %.1f\n", category, amount))
	}

	response.WriteString(fmt.Sprintf("\n💸 *Расход*: %.1f\n", totalExpense))
	for category, amount := range expenseCategories {
		response.WriteString(fmt.Sprintf("  - %s: %.1f\n", category, amount))
	}

	response.WriteString(fmt.Sprintf("\n💹 *Итого*: %.1f", netIncome))

	return response.String()
}
