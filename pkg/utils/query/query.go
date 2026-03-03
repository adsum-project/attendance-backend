package query

import (
	"fmt"
	"strings"
)

// Guid returns LOWER(CONVERT(VARCHAR(36), col)) for SELECT output (Go expects lowercase UUID strings).
func Guid(col string) string {
	return "LOWER(CONVERT(VARCHAR(36), " + col + "))"
}

// GuidWhere returns "col = CONVERT(uniqueidentifier, param)" for WHERE clauses.
// Use instead of Guid(col) = LOWER(@p1) - avoids applying functions to the column (better for indexes).
func GuidWhere(col, param string) string {
	return col + " = CONVERT(uniqueidentifier, " + param + ")"
}

// Date returns CONVERT(VARCHAR(10), col, 23) for YYYY-MM-DD output.
func Date(col string) string {
	return "CONVERT(VARCHAR(10), " + col + ", 23)"
}

// Time returns CONVERT(VARCHAR(8), col, 108) for HH:MI:SS output.
func Time(col string) string {
	return "CONVERT(VARCHAR(8), " + col + ", 108)"
}

// DateTimeISO returns CONVERT(VARCHAR(33), col, 127) for ISO8601 datetime output.
func DateTimeISO(col string) string {
	return "CONVERT(VARCHAR(33), " + col + ", 127)"
}

// Room returns UPPER(LTRIM(RTRIM(col))) for normalized room display.
func Room(col string) string {
	return "UPPER(LTRIM(RTRIM(" + col + ")))"
}

// OrderBy returns an ORDER BY clause. If sortBy is empty or not in allowedColumns, returns defaultOrder.
// sortOrder must be "asc" or "desc"; otherwise asc is used.
// allowedColumns maps API sort keys (e.g. "courseCode") to SQL column names (e.g. "course_code").
func OrderBy(sortBy, sortOrder, defaultOrder string, allowedColumns map[string]string) string {
	col, ok := allowedColumns[sortBy]
	if !ok || sortBy == "" {
		return defaultOrder
	}
	dir := "ASC"
	if sortOrder == "desc" {
		dir = "DESC"
	}
	return "ORDER BY " + col + " " + dir
}

func Update(updates map[string]any) (clause string, args []any, nextParam string) {
	var sets []string
	p := 1
	for col, v := range updates {
		if v == nil {
			continue
		}
		s, ok := v.(*string)
		if !ok || s == nil {
			continue
		}
		sets = append(sets, col+" = @p"+fmt.Sprint(p))
		args = append(args, *s)
		p++
	}
	if len(sets) == 0 {
		return "", nil, "@p1"
	}
	return strings.Join(sets, ", "), args, "@p" + fmt.Sprint(p)
}

func UpdateAndCast(updates map[string]any, castColumns map[string]string) (clause string, args []any, nextParam string) {
	var sets []string
	p := 1
	for col, v := range updates {
		if v == nil {
			continue
		}
		castType := castColumns[col]
		switch val := v.(type) {
		case *string:
			if val == nil {
				continue
			}
			if castType != "" {
				sets = append(sets, col+" = CAST(@p"+fmt.Sprint(p)+" AS "+castType+")")
			} else {
				sets = append(sets, col+" = @p"+fmt.Sprint(p))
			}
			args = append(args, *val)
			p++
		case *int:
			if val == nil {
				continue
			}
			sets = append(sets, col+" = @p"+fmt.Sprint(p))
			args = append(args, *val)
			p++
		}
	}
	if len(sets) == 0 {
		return "", nil, "@p1"
	}
	return strings.Join(sets, ", "), args, "@p"+fmt.Sprint(p)
}
