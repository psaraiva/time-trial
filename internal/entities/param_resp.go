package entities

import (
	"fmt"
	"sync"
	"unicode"
)

type PropertyType string

const (
	PropertyTypeString      PropertyType = "string"
	PropertyTypeInt         PropertyType = "int"
	PropertyTypeFloat       PropertyType = "float"
	PropertyTypeUUID        PropertyType = "uuid"
	PropertyTypeStringFunny PropertyType = "string-funny"
)

type PropertyStringConfig struct {
	Chars string `json:"chars"`
}

type PropertyIntConfig struct {
	IsAcceptNegativeValue bool `json:"isAcceptNegativeValue"`
}

type PropertyFloatConfig struct {
	FloatPrecision        int  `json:"floatPrecision"`
	IsAcceptNegativeValue bool `json:"isAcceptNegativeValue"`
}

type PropertyUUIDConfig struct {
	Version int `json:"version"`
}

type PropertyBase struct {
	Name       string       `json:"name"`
	Type       PropertyType `json:"type"`
	IsRequired bool         `json:"isRequired"`
	MaxLength  int          `json:"maxLength"`
	MinLength  int          `json:"minLength"`
}

type Property struct {
	PropertyBase
	PropertyString *PropertyStringConfig `json:"propertyString,omitempty"`
	PropertyInt    *PropertyIntConfig    `json:"propertyInt,omitempty"`
	PropertyFloat  *PropertyFloatConfig  `json:"propertyFloat,omitempty"`
	PropertyUUID   *PropertyUUIDConfig   `json:"propertyUUID,omitempty"`
}

type ItemConfig struct {
	IsCollection bool       `json:"isColection"`
	Quantity     int        `json:"quantity"`
	Properties   []Property `json:"properties"`
}

type ResponseConfig struct {
	StatusCode int        `json:"statusCode"`
	Item       ItemConfig `json:"item"`
}

type ParamResp struct {
	mu     sync.RWMutex
	config *ResponseConfig
}

func NewParamResp() *ParamResp {
	return &ParamResp{}
}

func (p *ParamResp) Set(config *ResponseConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
}

func (p *ParamResp) Get() *ResponseConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

func (p *ParamResp) IsActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config != nil
}

func (p *ParamResp) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = nil
}

var validStatusCodes = func() map[int]struct{} {
	codes := []int{
		// 2xx
		200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
		// 4xx
		400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410,
		411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423,
		424, 425, 426, 428, 429, 431, 451,
		// 5xx
		500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511,
	}
	m := make(map[int]struct{}, len(codes))
	for _, c := range codes {
		m[c] = struct{}{}
	}
	return m
}()

func ValidateResponseConfig(c *ResponseConfig) error {
	if _, ok := validStatusCodes[c.StatusCode]; !ok {
		return fmt.Errorf("statusCode %d is not a valid 2xx, 4xx or 5xx HTTP status code", c.StatusCode)
	}
	if len(c.Item.Properties) == 0 {
		return fmt.Errorf("item.properties must have at least one property")
	}
	if c.Item.IsCollection && c.Item.Quantity < 1 {
		return fmt.Errorf("item.quantity must be at least 1 when isColection is true")
	}
	for i, prop := range c.Item.Properties {
		if err := validateProperty(prop, i); err != nil {
			return err
		}
	}
	return nil
}

func validateProperty(p Property, idx int) error {
	if p.Name == "" {
		return fmt.Errorf("property[%d].name cannot be empty", idx)
	}
	for _, r := range p.Name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("property[%d].name %q contains invalid character %q", idx, p.Name, r)
		}
	}

	switch p.Type {
	case PropertyTypeString, PropertyTypeInt, PropertyTypeFloat, PropertyTypeUUID, PropertyTypeStringFunny:
	default:
		return fmt.Errorf("property[%d].type must be one of: string, int, float, uuid, string-funny", idx)
	}

	if p.Type != PropertyTypeStringFunny && p.MinLength > p.MaxLength {
		return fmt.Errorf("property[%d].minLength cannot be greater than maxLength", idx)
	}

	switch p.Type {
	case PropertyTypeString:
		if p.PropertyString == nil {
			return fmt.Errorf("property[%d] of type string requires propertyString config", idx)
		}
		if err := validateStringChars(p.PropertyString.Chars, idx); err != nil {
			return err
		}
	case PropertyTypeInt:
		if p.PropertyInt == nil {
			return fmt.Errorf("property[%d] of type int requires propertyInt config", idx)
		}
	case PropertyTypeFloat:
		if p.PropertyFloat == nil {
			return fmt.Errorf("property[%d] of type float requires propertyFloat config", idx)
		}
		if p.PropertyFloat.FloatPrecision < 0 {
			return fmt.Errorf("property[%d].propertyFloat.floatPrecision cannot be negative", idx)
		}
		if p.PropertyFloat.FloatPrecision > 5 {
			return fmt.Errorf("property[%d].propertyFloat.floatPrecision cannot exceed 5", idx)
		}
	case PropertyTypeUUID:
		if p.PropertyUUID == nil {
			return fmt.Errorf("property[%d] of type uuid requires propertyUUID config", idx)
		}
		if p.PropertyUUID.Version != 1 && p.PropertyUUID.Version != 4 && p.PropertyUUID.Version != 7 {
			return fmt.Errorf("property[%d].propertyUUID.version must be 1, 4 or 7", idx)
		}
	}

	return nil
}

func validateStringChars(chars string, idx int) error {
	if chars == "" {
		return fmt.Errorf("property[%d].propertyString.chars cannot be empty", idx)
	}
	for _, r := range chars {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("property[%d].propertyString.chars must contain only letters (A-Z, a-z), got %q", idx, r)
		}
	}
	return nil
}
