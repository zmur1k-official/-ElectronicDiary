package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// writeJSON отправляет JSON-ответ с заданным HTTP-статусом.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}

// writeError отправляет JSON-ошибку в едином формате.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": fmt.Sprintf("%s", message),
	})
}

// normalizeClassName приводит обозначение класса к каноничному виду.
func normalizeClassName(s string) string {
	clean := strings.ToUpper(strings.TrimSpace(s))
	clean = strings.ReplaceAll(clean, " ", "")
	replacer := strings.NewReplacer(
		"А", "A",
		"В", "B",
		"Е", "E",
		"К", "K",
		"М", "M",
		"Н", "H",
		"О", "O",
		"Р", "P",
		"С", "C",
		"Т", "T",
		"У", "Y",
		"Х", "X",
	)
	clean = replacer.Replace(clean)
	var b strings.Builder
	for _, r := range clean {
		if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
