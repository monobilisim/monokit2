package lib

import "fmt"

func CreateMarkdownTable[T any](headers []string, valuesArray [][]T) string {
	var table string

	for _, header := range headers {
		table += "| " + header + " "
	}
	table += "|\n"

	for range headers {
		table += "| --- "
	}
	table += "|\n"

	for _, values := range valuesArray {
		for _, value := range values {
			table += "| " + fmt.Sprintf("%v", value) + " "
		}

		table += "|\n"
	}

	return table
}
