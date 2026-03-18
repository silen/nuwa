package binding

import "testing"

type validatorSample struct {
	Name string `binding:"required"`
}

func TestDefaultValidatorReturnsTranslatedValidationError(t *testing.T) {
	t.Parallel()

	v := Validator()
	err := v.ValidateStruct(validatorSample{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if err.Error() == "" {
		t.Fatalf("expected translated validation error message")
	}
}

func TestDefaultValidatorAcceptsValidStruct(t *testing.T) {
	t.Parallel()

	v := Validator()
	if err := v.ValidateStruct(validatorSample{Name: "nuwa"}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
