package requestlog

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	redactedValue  = "[REDACTED]"
	truncatedValue = "\n[TRUNCATED]"
)

var sensitiveHeaders = map[string]struct{}{
	"authorization":       {},
	"cookie":              {},
	"proxy-authorization": {},
	"set-cookie":          {},
	"x-api-key":           {},
	"x-auth-token":        {},
}

var secretFields = map[string]struct{}{
	"access_token":     {},
	"api_key":          {},
	"authorization":    {},
	"captcha_token":    {},
	"client_token":     {},
	"cookie":           {},
	"new_password":     {},
	"old_password":     {},
	"passwd":           {},
	"password":         {},
	"password_confirm": {},
	"refresh_token":    {},
	"secret":           {},
	"token":            {},
	"user_answer":      {},
}

func sanitizeHeaders(headers http.Header) string {
	clean := make(http.Header, len(headers))
	for key, values := range headers {
		if _, sensitive := sensitiveHeaders[strings.ToLower(key)]; sensitive {
			clean[key] = []string{redactedValue}
			continue
		}
		clean[key] = append([]string(nil), values...)
	}
	encoded, err := json.Marshal(clean)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func sanitizeBody(contentType string, body []byte, truncated bool) string {
	if len(body) == 0 {
		return ""
	}
	mediaType, _, _ := mime.ParseMediaType(contentType)
	var result string
	switch {
	case mediaType == "application/json" || strings.HasSuffix(mediaType, "+json"):
		result = sanitizeJSON(body)
	case mediaType == "application/x-www-form-urlencoded":
		result = sanitizeForm(body)
	default:
		result = "[UNSTRUCTURED BODY OMITTED]"
	}
	if truncated {
		result += truncatedValue
	}
	return result
}

func sanitizeJSON(body []byte) string {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		return "[INVALID JSON OMITTED]"
	}
	clean := sanitizeJSONValue(value, "")
	encoded, err := json.Marshal(clean)
	if err != nil {
		return "[JSON OMITTED]"
	}
	return string(encoded)
}

func sanitizeJSONValue(value any, key string) any {
	normalizedKey := normalizeField(key)
	if _, sensitive := secretFields[normalizedKey]; sensitive {
		return redactedValue
	}
	switch normalizedKey {
	case "email", "user_email":
		if text, ok := value.(string); ok {
			return maskEmail(text)
		}
	case "mobile", "phone", "phone_number":
		if text, ok := value.(string); ok {
			return maskIdentifier(text)
		}
	case "id_card", "id_number", "identity_number":
		if text, ok := value.(string); ok {
			return maskIdentifier(text)
		}
	}

	switch typed := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(typed))
		for childKey, childValue := range typed {
			clean[childKey] = sanitizeJSONValue(childValue, childKey)
		}
		return clean
	case []any:
		clean := make([]any, len(typed))
		for index, childValue := range typed {
			clean[index] = sanitizeJSONValue(childValue, key)
		}
		return clean
	default:
		return value
	}
}

func sanitizeForm(body []byte) string {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "[INVALID FORM OMITTED]"
	}
	for key := range values {
		normalizedKey := normalizeField(key)
		if _, sensitive := secretFields[normalizedKey]; sensitive {
			values[key] = []string{redactedValue}
			continue
		}
		for index, value := range values[key] {
			switch normalizedKey {
			case "email", "user_email":
				values[key][index] = maskEmail(value)
			case "mobile", "phone", "phone_number", "id_card", "id_number", "identity_number":
				values[key][index] = maskIdentifier(value)
			}
		}
	}
	return values.Encode()
}

func normalizeField(field string) string {
	return strings.NewReplacer("-", "_", " ", "_").Replace(strings.ToLower(field))
}

func maskEmail(value string) string {
	local, domain, found := strings.Cut(value, "@")
	if !found || local == "" {
		return redactedValue
	}
	return local[:1] + "***@" + domain
}

func maskIdentifier(value string) string {
	if len(value) <= 7 {
		return redactedValue
	}
	return value[:3] + "****" + value[len(value)-4:]
}
