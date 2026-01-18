package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type FieldError struct {
	Field string      `json:"field"`
	Value interface{} `json:"value,omitempty"`
	Error string      `json:"error"`
}

var (
	vOnce sync.Once
	v     *validator.Validate
)

func getValidator() *validator.Validate {
	vOnce.Do(func() {
		v = validator.New(validator.WithRequiredStructEnabled())
	})
	return v
}

// DecodeStrict decodifica JSON en dst:
// - body vacío => error
// - JSON inválido => error
// - campos desconocidos => error
// - trailing data (ej: {}{}) => error
func DecodeStrict(body []byte, dst any) *fiber.Error {
	if len(bytes.TrimSpace(body)) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "empty body")
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return fiber.NewError(fiber.StatusBadRequest, "empty body")
		}
		if field, ok := unknownFieldName(err); ok {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("unknown field: %s", field))
		}
		return fiber.NewError(fiber.StatusBadRequest, "invalid json")
	}

	// Debe terminar exactamente aquí (sin basura extra)
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json (trailing data)")
	}

	return nil
}

func ValidateStructDTO(dto any) []FieldError {
	err := getValidator().Struct(dto)
	if err == nil {
		return nil
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "body", Error: "validation error"}}
	}

	t := reflect.TypeOf(dto)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	out := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		out = append(out, FieldError{
			Field: jsonFieldName(t, fe.StructField()),
			Value: fe.Value(),
			Error: validationMessage(fe),
		})
	}
	return out
}

// BindAndValidate combina: strict decode + validate tags
func BindAndValidate(c *fiber.Ctx, dst any) []FieldError {
	if ferr := DecodeStrict(c.Body(), dst); ferr != nil {
		return []FieldError{{Field: "body", Error: ferr.Message}}
	}
	return ValidateStructDTO(dst)
}

func RespondValidation(c *fiber.Ctx, errs []FieldError) error {
	if len(errs) == 0 {
		return nil
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"valid":  false,
		"error":  "validation_failed",
		"fields": errs,
	})
}

// --- helpers

func unknownFieldName(err error) (string, bool) {
	// encoding/json devuelve: `json: unknown field "x"`
	const prefix = `json: unknown field "`
	msg := err.Error()
	if strings.HasPrefix(msg, prefix) && strings.HasSuffix(msg, `"`) {
		return strings.TrimSuffix(strings.TrimPrefix(msg, prefix), `"`), true
	}
	return "", false
}

func jsonFieldName(t reflect.Type, structField string) string {
	if t.Kind() != reflect.Struct {
		return strings.ToLower(structField)
	}
	f, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}
	tag := f.Tag.Get("json")
	if tag == "" || tag == "-" {
		return strings.ToLower(structField)
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return strings.ToLower(structField)
	}
	return name
}

func validationMessage(fe validator.FieldError) string {
	// Mensajes cortos (se pueden mejorar según necesidad)
	switch fe.Tag() {
	case "required":
		return "required"
	case "email":
		return "invalid email"
	case "min":
		return fmt.Sprintf("min %s", fe.Param())
	case "max":
		return fmt.Sprintf("max %s", fe.Param())
	default:
		return fe.Tag()
	}
}
