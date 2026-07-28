package normalizer

import (
	"html"
	"regexp"
	"strings"
)

var (
	richTextTag = regexp.MustCompile(`<[^>]+>`)
	placeholder = regexp.MustCompile(`\{(\d+)\}`)
	cusTag      = regexp.MustCompile(`\{Cus:[^,]+,([^}]+)\}`)
	genderTag   = regexp.MustCompile(`\{(?:Male|Female)=([^;}]+)(?:;[^}]+)?\}`)
	sapTag      = regexp.MustCompile(`<SapTag=\d+>(.*?)</SapTag>`)
)

func CleanText(value string) string {
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = richTextTag.ReplaceAllString(value, "")
	value = html.UnescapeString(value)
	lines := strings.Split(value, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func ApplyParams(text string, params []any) (string, []string) {
	if text == "" {
		return "", nil
	}
	warnings := make([]string, 0)
	text = placeholder.ReplaceAllStringFunc(text, func(token string) string {
		match := placeholder.FindStringSubmatch(token)
		index := 0
		for _, digit := range match[1] {
			index = index*10 + int(digit-'0')
		}
		if index >= len(params) {
			warnings = append(warnings, "missing parameter "+token)
			return token
		}
		switch value := params[index].(type) {
		case []any:
			if len(value) == 0 {
				return ""
			}
			return stringify(value[0])
		default:
			return stringify(value)
		}
	})
	text = cusTag.ReplaceAllStringFunc(text, func(token string) string {
		match := cusTag.FindStringSubmatch(token)
		options := strings.Fields(match[1])
		if len(options) == 0 {
			return ""
		}
		parts := strings.SplitN(options[0], "=", 2)
		return parts[len(parts)-1]
	})
	text = genderTag.ReplaceAllString(text, "$1")
	text = sapTag.ReplaceAllString(text, "$1")
	return text, warnings
}

func stringify(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(toJSONScalar(typed)), `"`, ""))
	}
}

func toJSONScalar(value any) string {
	switch typed := value.(type) {
	case float64:
		return strings.TrimRight(strings.TrimRight(formatFloat(typed), "0"), ".")
	case float32:
		return strings.TrimRight(strings.TrimRight(formatFloat(float64(typed)), "0"), ".")
	case int:
		return formatInt(int64(typed))
	case int64:
		return formatInt(typed)
	default:
		return ""
	}
}

func formatFloat(value float64) string {
	return strconvFormatFloat(value)
}

func formatInt(value int64) string {
	return strconvFormatInt(value)
}
