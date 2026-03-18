package binding

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	ginbinding "github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var defaultBindingValidator ginbinding.StructValidator = &defaultValidator{}

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
	trans    ut.Translator
}

type SliceValidationError []error

// Error concatenates all error elements in SliceValidationError into a single string separated by \n.
func (err SliceValidationError) Error() string {
	n := len(err)
	switch n {
	case 0:
		return ""
	default:
		var b strings.Builder
		if err[0] != nil {
			fmt.Fprintf(&b, "[%d]: %s", 0, err[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if err[i] != nil {
					b.WriteString("\n")
					fmt.Fprintf(&b, "[%d]: %s", i, err[i].Error())
				}
			}
		}
		return b.String()
	}
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}

	value := reflect.ValueOf(obj)
	switch value.Kind() {
	case reflect.Ptr:
		if value.Elem().Kind() != reflect.Struct {
			return v.ValidateStruct(value.Elem().Interface())
		}
		return v.validateStruct(obj)
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := v.ValidateStruct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

func (v *defaultValidator) validateStruct(obj any) error {
	v.lazyinit()
	err := v.validate.Struct(obj)
	if err != nil {
		var errs validator.ValidationErrors
		if !errors.As(err, &errs) {
			return err
		}
		for _, e := range errs {
			return errors.New(e.Translate(v.trans))
		}
	}
	return nil
}

// Engine returns the underlying validator engine which powers the default
// Validator instance. This is useful if you want to register custom validations
// or struct level validations. See validator GoDoc for more info -
// https://pkg.go.dev/github.com/go-playground/validator/v10
func (v *defaultValidator) Engine() any {
	v.lazyinit()
	return v.validate
}

// Validator returns Nuwa's default Gin binding validator implementation.
func Validator() ginbinding.StructValidator {
	return defaultBindingValidator
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		validate := validator.New()
		chinese := zh.New()
		uni := ut.New(chinese, chinese)
		v.trans, _ = uni.GetTranslator("zh_CN")
		zhTranslations.RegisterDefaultTranslations(validate, v.trans)

		v.validate = validate
		v.validate.SetTagName("binding")
	})
}
