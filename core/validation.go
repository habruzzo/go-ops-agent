package core

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation functionality for the framework
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()
	return &Validator{validator: v}
}

// ValidateStruct validates a struct using the validator tags
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validator.Struct(s); err != nil {
		return NewValidationError("validator", "validate", fmt.Sprintf("validation failed: %v", err))
	}
	return nil
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validator.Var(field, tag); err != nil {
		return NewValidationError("validator", "validate-var", fmt.Sprintf("validation failed: %v", err))
	}
	return nil
}

// ValidateFrameworkConfig validates a FrameworkConfig
func (v *Validator) ValidateFrameworkConfig(config *FrameworkConfig) error {
	// Validate the main config struct
	if err := v.ValidateStruct(config); err != nil {
		return err
	}

	// Validate individual plugin configs
	for i, plugin := range config.Plugins {
		if err := v.ValidateStruct(&plugin); err != nil {
			return NewValidationError("validator", "validate-plugin", 
				fmt.Sprintf("plugin %d (%s) validation failed: %v", i, plugin.Name, err))
		}
	}

	return nil
}

// ValidatePluginConfig validates a PluginConfig
func (v *Validator) ValidatePluginConfig(config *PluginConfig) error {
	return v.ValidateStruct(config)
}

// GetValidationErrors returns detailed validation errors
func (v *Validator) GetValidationErrors(err error) []ValidationErrorDetail {
	var details []ValidationErrorDetail
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			detail := ValidationErrorDetail{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Value:   fieldError.Value(),
				Message: getValidationMessage(fieldError),
			}
			details = append(details, detail)
		}
	}
	
	return details
}

// ValidationErrorDetail provides detailed information about a validation error
type ValidationErrorDetail struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// Global validator instance
var globalValidator = NewValidator()

// ValidateStruct validates a struct using the global validator
func ValidateStruct(s interface{}) error {
	return globalValidator.ValidateStruct(s)
}

// ValidateVar validates a single variable using the global validator
func ValidateVar(field interface{}, tag string) error {
	return globalValidator.ValidateVar(field, tag)
}

// ValidateFrameworkConfig validates a FrameworkConfig using the global validator
func ValidateFrameworkConfig(config *FrameworkConfig) error {
	return globalValidator.ValidateFrameworkConfig(config)
}

// ValidatePluginConfig validates a PluginConfig using the global validator
func ValidatePluginConfig(config *PluginConfig) error {
	return globalValidator.ValidatePluginConfig(config)
}
