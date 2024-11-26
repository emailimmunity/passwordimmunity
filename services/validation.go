package services

import (
	"context"
	"reflect"
	"regexp"
	"strings"

	"github.com/emailimmunity/passwordimmunity/db/models"
	"github.com/go-playground/validator/v10"
)

type ValidationService interface {
	Validate(ctx context.Context, data interface{}) error
	ValidateField(ctx context.Context, field string, value interface{}, tag string) error
	RegisterValidation(tag string, fn validator.Func) error
	RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) error
}

type validationService struct {
	validator *validator.Validate
}

func NewValidationService() ValidationService {
	v := validator.New()

	// Register custom validations
	service := &validationService{validator: v}
	service.registerCustomValidations()

	return service
}

func (s *validationService) Validate(ctx context.Context, data interface{}) error {
	return s.validator.StructCtx(ctx, data)
}

func (s *validationService) ValidateField(ctx context.Context, field string, value interface{}, tag string) error {
	return s.validator.VarCtx(ctx, value, tag)
}

func (s *validationService) RegisterValidation(tag string, fn validator.Func) error {
	return s.validator.RegisterValidation(tag, fn)
}

func (s *validationService) RegisterStructValidation(fn validator.StructLevelFunc, types ...interface{}) error {
	s.validator.RegisterStructValidation(fn, types...)
	return nil
}

func (s *validationService) registerCustomValidations() {
	// Password strength validation
	s.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		return validatePasswordStrength(password)
	})

	// Username validation
	s.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		return validateUsername(username)
	})

	// Organization name validation
	s.RegisterValidation("orgname", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		return validateOrgName(name)
	})

	// URL validation with custom requirements
	s.RegisterValidation("secureurl", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		return validateSecureURL(url)
	})

	// Register struct-level validations
	s.RegisterStructValidation(validateVaultItem, models.VaultItem{})
	s.RegisterStructValidation(validateCollection, models.Collection{})
}

func validatePasswordStrength(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)
	)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func validateUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	return matched
}

func validateOrgName(name string) bool {
	if len(name) < 2 || len(name) > 64 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\s_-]+$`, name)
	return matched
}

func validateSecureURL(url string) bool {
	if !strings.HasPrefix(url, "https://") {
		return false
	}

	matched, _ := regexp.MatchString(`^https://[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z]{2,}(/.*)?$`, url)
	return matched
}

func validateVaultItem(sl validator.StructLevel) {
	item := sl.Current().Interface().(models.VaultItem)

	// Validate item type-specific fields
	switch item.Type {
	case "login":
		if item.Login == nil {
			sl.ReportError(item.Login, "Login", "login", "required", "")
		}
	case "card":
		if item.Card == nil {
			sl.ReportError(item.Card, "Card", "card", "required", "")
		}
	case "identity":
		if item.Identity == nil {
			sl.ReportError(item.Identity, "Identity", "identity", "required", "")
		}
	}
}

func validateCollection(sl validator.StructLevel) {
	collection := sl.Current().Interface().(models.Collection)

	// Validate collection-specific rules
	if collection.IsShared && len(collection.Groups) == 0 && len(collection.Users) == 0 {
		sl.ReportError(collection.IsShared, "IsShared", "isshared", "shared_access", "")
	}
}
