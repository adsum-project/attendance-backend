package query

import (
	"fmt"
	"strings"
)

func Guid(col string) string {
	return "LOWER(CONVERT(VARCHAR(36), " + col + "))"
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
		s, ok := v.(*string)
		if !ok || s == nil {
			continue
		}
		castType := castColumns[col]
		if castType != "" {
			sets = append(sets, col+" = CAST(@p"+fmt.Sprint(p)+" AS "+castType+")")
		} else {
			sets = append(sets, col+" = @p"+fmt.Sprint(p))
		}
		args = append(args, *s)
		p++
	}
	if len(sets) == 0 {
		return "", nil, "@p1"
	}
	return strings.Join(sets, ", "), args, "@p" + fmt.Sprint(p)
}
