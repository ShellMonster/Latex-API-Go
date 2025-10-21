package api

import "testing"

func TestValidateFormula(t *testing.T) {
	input := "  E=mc^2  "
	out, err := validateFormula(input)
	if err != nil {
		t.Fatalf("期望通过校验，结果报错: %v", err)
	}
	if out != "E=mc^2" {
		t.Fatalf("trim 结果不符合预期: %q", out)
	}
}

func TestValidateFormula_Empty(t *testing.T) {
	if _, err := validateFormula("   "); err != ErrEmptyFormula {
		t.Fatalf("应返回 ErrEmptyFormula，实际: %v", err)
	}
}

func TestValidateFormula_TooLarge(t *testing.T) {
	large := make([]byte, maxFormulaBytes+1)
	for i := range large {
		large[i] = 'a'
	}
	if _, err := validateFormula(string(large)); err != ErrFormulaTooLarge {
		t.Fatalf("应返回 ErrFormulaTooLarge，实际: %v", err)
	}
}

func TestValidateFormula_InvalidControl(t *testing.T) {
	input := "abc\x07def"
	if _, err := validateFormula(input); err != ErrInvalidCharacters {
		t.Fatalf("应返回 ErrInvalidCharacters，实际: %v", err)
	}
}
