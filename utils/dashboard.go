package utils

import "github.com/gin-gonic/gin"

func FormatPerformanceData(rawData []map[string]interface{}) []gin.H {
	var formatted []gin.H
	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for _, item := range rawData {
		id := item["_id"].(map[string]interface{})
		count := item["count"]
		if count == nil {
			count = 0
		}

		formatted = append(formatted, gin.H{
			"month": monthNames[int(id["month"].(int32))],
			"year":  id["year"],
			"count": count,
		})
	}
	return formatted
}
