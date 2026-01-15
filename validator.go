// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuwa

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
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

// validateStruct receives struct type
func (v *defaultValidator) validateStruct(obj any) error {
	v.lazyinit()
	err := v.validate.Struct(obj)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			return errors.New(e.Translate(Trans))
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

// 定义一个全局翻译器T
var Trans ut.Translator

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		validate := validator.New()

		// 注册自定义验证器
		registerCustomValidators(validate)

		chinese := zh.New()
		uni := ut.New(chinese, chinese)
		Trans, _ = uni.GetTranslator("zh_CN")

		// 注册中文翻译
		zhTranslations.RegisterDefaultTranslations(validate, Trans)
		registerCustomTranslations(validate, Trans)

		v.validate = validate
		v.validate.SetTagName("binding")
	})
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators(validate *validator.Validate) {
	// 手机号验证
	validate.RegisterValidation("mobile", validateMobile)

	// 身份证号验证
	validate.RegisterValidation("idcard", validateIDCard)

	// IP地址验证
	validate.RegisterValidation("ipaddr", validateIPAddr)

	// MAC地址验证
	validate.RegisterValidation("mac", validateMAC)

	// 中文验证
	validate.RegisterValidation("chinese", validateChinese)
}

// validateMobile 验证手机号
func validateMobile(fl validator.FieldLevel) bool {
	mobile := fl.Field().String()
	// 匹配中国大陆手机号格式
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return reg.MatchString(mobile)
}

// validateIDCard 验证身份证号
func validateIDCard(fl validator.FieldLevel) bool {
	idcard := fl.Field().String()
	// 简单验证身份证格式（18位）
	reg := regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`)
	return reg.MatchString(idcard)
}

// validateIPAddr 验证IP地址
func validateIPAddr(fl validator.FieldLevel) bool {
	ip := fl.Field().String()
	// 验证IPv4地址格式
	reg := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return reg.MatchString(ip)
}

// validateMAC 验证MAC地址
func validateMAC(fl validator.FieldLevel) bool {
	mac := fl.Field().String()
	// 验证MAC地址格式
	reg := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	return reg.MatchString(mac)
}

// validateChinese 验证纯中文
func validateChinese(fl validator.FieldLevel) bool {
	text := fl.Field().String()
	// 验证是否只包含中文字符
	reg := regexp.MustCompile(`^[\u4e00-\u9fa5]+$`)
	return reg.MatchString(text)
}

// registerCustomTranslations 注册自定义翻译
func registerCustomTranslations(validate *validator.Validate, trans ut.Translator) {
	// 手机号验证翻译
	validate.RegisterTranslation("mobile", trans, func(ut ut.Translator) error {
		return ut.Add("mobile", "{0}不是有效的手机号", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mobile", fe.Field())
		return t
	})

	// 身份证验证翻译
	validate.RegisterTranslation("idcard", trans, func(ut ut.Translator) error {
		return ut.Add("idcard", "{0}不是有效的身份证号", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("idcard", fe.Field())
		return t
	})

	// IP地址验证翻译
	validate.RegisterTranslation("ipaddr", trans, func(ut ut.Translator) error {
		return ut.Add("ipaddr", "{0}不是有效的IP地址", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("ipaddr", fe.Field())
		return t
	})

	// MAC地址验证翻译
	validate.RegisterTranslation("mac", trans, func(ut ut.Translator) error {
		return ut.Add("mac", "{0}不是有效的MAC地址", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("mac", fe.Field())
		return t
	})

	// 中文验证翻译
	validate.RegisterTranslation("chinese", trans, func(ut ut.Translator) error {
		return ut.Add("chinese", "{0}应该只包含中文字符", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("chinese", fe.Field())
		return t
	})
}
