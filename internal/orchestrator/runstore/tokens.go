package runstore

import (
	"encoding/json"
	"log"
)

func ExtractTokens(payloadJSON string) (input, output, total int64, err error) {
	if payloadJSON == "" {
		return 0, 0, 0, nil
	}
	var payload any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		log.Printf("orchestrator token extraction skipped malformed payload: %v", err)
		return 0, 0, 0, nil
	}
	summary, ok := findTokenSummary(payload)
	if !ok {
		return 0, 0, 0, nil
	}
	if summary.total == 0 {
		summary.total = summary.input + summary.output
	}
	return summary.input, summary.output, summary.total, nil
}

type extractedTokens struct {
	input  int64
	output int64
	total  int64
}

func findTokenSummary(value any) (extractedTokens, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"total_token_usage", "token_usage", "usage"} {
			if nested, ok := typed[key]; ok {
				if summary, ok := extractTokenMap(nested); ok {
					return summary, true
				}
			}
		}
		if summary, ok := extractTokenMap(typed); ok {
			return summary, true
		}
		for key, nested := range typed {
			if key == "last_token_usage" {
				continue
			}
			if summary, ok := findTokenSummary(nested); ok {
				return summary, true
			}
		}
	case []any:
		for _, item := range typed {
			if summary, ok := findTokenSummary(item); ok {
				return summary, true
			}
		}
	}
	return extractedTokens{}, false
}

func extractTokenMap(value any) (extractedTokens, bool) {
	fields, ok := value.(map[string]any)
	if !ok {
		return extractedTokens{}, false
	}
	input, inputOK := intField(fields, "input_tokens", "inputTokens", "input_token_count", "inputTokenCount")
	output, outputOK := intField(fields, "output_tokens", "outputTokens", "output_token_count", "outputTokenCount")
	total, totalOK := intField(fields, "total_tokens", "totalTokens", "total_token_count", "totalTokenCount")
	if !inputOK && !outputOK && !totalOK {
		return extractedTokens{}, false
	}
	return extractedTokens{input: input, output: output, total: total}, true
}

func intField(fields map[string]any, keys ...string) (int64, bool) {
	for _, key := range keys {
		value, ok := fields[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			if typed >= 0 {
				return int64(typed), true
			}
		case int64:
			if typed >= 0 {
				return typed, true
			}
		case int:
			if typed >= 0 {
				return int64(typed), true
			}
		}
	}
	return 0, false
}
