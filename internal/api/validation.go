package api

import (
	"errors"
	"strings"
)

const (
	maxFormulaBytes = 5 * 1024 // 5KB 限制
)

var (
	ErrEmptyFormula      = errors.New("公式内容为空")
	ErrFormulaTooLarge   = errors.New("公式内容超过 5KB 限制")
	ErrInvalidCharacters = errors.New("公式包含非法控制字符")
)

func validateFormula(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrEmptyFormula
	}

	if len([]byte(trimmed)) > maxFormulaBytes {
		return "", ErrFormulaTooLarge
	}

	if containsInvalidControl(trimmed) {
		return "", ErrInvalidCharacters
	}

	return trimmed, nil
}

func containsInvalidControl(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f {
			return true
		}
	}
	return false
}
