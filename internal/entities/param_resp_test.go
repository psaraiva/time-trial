package entities

import (
	"testing"
)

func makeStringProperty(name string, min, max int, chars string) Property {
	return Property{
		PropertyBase: PropertyBase{
			Name:       name,
			Type:       PropertyTypeString,
			IsRequired: true,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyString: &PropertyStringConfig{Chars: chars},
	}
}

func makeIntProperty(name string, min, max int, acceptNeg bool) Property {
	return Property{
		PropertyBase: PropertyBase{
			Name:       name,
			Type:       PropertyTypeInt,
			IsRequired: true,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyInt: &PropertyIntConfig{IsAcceptNegativeValue: acceptNeg},
	}
}

func makeFloatProperty(name string, min, max, precision int, acceptNeg bool) Property {
	return Property{
		PropertyBase: PropertyBase{
			Name:       name,
			Type:       PropertyTypeFloat,
			IsRequired: true,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyFloat: &PropertyFloatConfig{
			FloatPrecision:        precision,
			IsAcceptNegativeValue: acceptNeg,
		},
	}
}

func makeStringFunnyProperty(name string) Property {
	return Property{
		PropertyBase: PropertyBase{
			Name:       name,
			Type:       PropertyTypeStringFunny,
			IsRequired: true,
		},
	}
}

func makeUUIDProperty(name string, version int) Property {
	return Property{
		PropertyBase: PropertyBase{
			Name:       name,
			Type:       PropertyTypeUUID,
			IsRequired: true,
		},
		PropertyUUID: &PropertyUUIDConfig{Version: version},
	}
}

func validConfig() *ResponseConfig {
	return &ResponseConfig{
		StatusCode: 200,
		Item: ItemConfig{
			IsCollection: true,
			Quantity:     3,
			Properties: []Property{
				makeStringProperty("name", 3, 10, "abcABC"),
				makeIntProperty("version", 0, 99, false),
				makeFloatProperty("value", 0, 100, 2, false),
			},
		},
	}
}

func TestValidateResponseConfig_Valid(t *testing.T) {
	t.Parallel()
	if err := ValidateResponseConfig(validConfig()); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateResponseConfig_StatusCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		// 2xx valid
		{name: "200 ok", statusCode: 200, wantErr: false},
		{name: "201 ok", statusCode: 201, wantErr: false},
		{name: "204 ok", statusCode: 204, wantErr: false},
		// 4xx valid
		{name: "400 ok", statusCode: 400, wantErr: false},
		{name: "401 ok", statusCode: 401, wantErr: false},
		{name: "404 ok", statusCode: 404, wantErr: false},
		{name: "422 ok", statusCode: 422, wantErr: false},
		// 5xx valid
		{name: "500 ok", statusCode: 500, wantErr: false},
		{name: "503 ok", statusCode: 503, wantErr: false},
		// invalid
		{name: "0 rejected", statusCode: 0, wantErr: true},
		{name: "100 rejected", statusCode: 100, wantErr: true},
		{name: "301 rejected", statusCode: 301, wantErr: true},
		{name: "600 rejected", statusCode: 600, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.StatusCode = tc.statusCode
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_NoProperties(t *testing.T) {
	t.Parallel()
	c := validConfig()
	c.Item.Properties = nil
	if err := ValidateResponseConfig(c); err == nil {
		t.Error("expected error for empty properties")
	}
}

func TestValidateResponseConfig_CollectionQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		isCollection bool
		quantity     int
		wantErr      bool
	}{
		{name: "collection quantity 1", isCollection: true, quantity: 1, wantErr: false},
		{name: "collection quantity 0", isCollection: true, quantity: 0, wantErr: true},
		{name: "no collection quantity 0", isCollection: false, quantity: 0, wantErr: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.IsCollection = tc.isCollection
			c.Item.Quantity = tc.quantity
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_PropertyName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		pname   string
		wantErr bool
	}{
		{name: "valid simple", pname: "myField", wantErr: false},
		{name: "valid with underscore", pname: "my_field", wantErr: false},
		{name: "valid with hyphen", pname: "my-field", wantErr: false},
		{name: "valid with digit", pname: "field1", wantErr: false},
		{name: "empty name", pname: "", wantErr: true},
		{name: "space in name", pname: "my field", wantErr: true},
		{name: "dot in name", pname: "my.field", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.Properties = []Property{makeStringProperty(tc.pname, 1, 5, "abc")}
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_PropertyType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ptype   PropertyType
		wantErr bool
	}{
		{name: "string", ptype: PropertyTypeString, wantErr: false},
		{name: "int", ptype: PropertyTypeInt, wantErr: false},
		{name: "float", ptype: PropertyTypeFloat, wantErr: false},
		{name: "uuid", ptype: PropertyTypeUUID, wantErr: false},
		{name: "string-funny", ptype: PropertyTypeStringFunny, wantErr: false},
		{name: "invalid", ptype: "bool", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var prop Property
			switch tc.ptype {
			case PropertyTypeString:
				prop = makeStringProperty("field", 1, 5, "abc")
			case PropertyTypeInt:
				prop = makeIntProperty("field", 0, 10, false)
			case PropertyTypeFloat:
				prop = makeFloatProperty("field", 0, 10, 2, false)
			case PropertyTypeUUID:
				prop = makeUUIDProperty("field", 4)
			case PropertyTypeStringFunny:
				prop = makeStringFunnyProperty("field")
			default:
				prop = Property{PropertyBase: PropertyBase{Name: "field", Type: tc.ptype, MinLength: 0, MaxLength: 10}}
			}
			c := validConfig()
			c.Item.Properties = []Property{prop}
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_MinMaxLength(t *testing.T) {
	t.Parallel()
	c := validConfig()
	c.Item.Properties = []Property{makeStringProperty("field", 10, 5, "abc")}
	if err := ValidateResponseConfig(c); err == nil {
		t.Error("expected error when minLength > maxLength")
	}
}

func TestValidateResponseConfig_StringChars(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		chars   string
		wantErr bool
	}{
		{name: "valid letters", chars: "abcABC", wantErr: false},
		{name: "empty chars", chars: "", wantErr: true},
		{name: "contains digit", chars: "abc1", wantErr: true},
		{name: "contains space", chars: "abc ", wantErr: true},
		{name: "contains special", chars: "abc!", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.Properties = []Property{makeStringProperty("field", 1, 5, tc.chars)}
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_MissingTypeConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		prop Property
	}{
		{
			name: "string without propertyString",
			prop: Property{PropertyBase: PropertyBase{Name: "f", Type: PropertyTypeString, MinLength: 1, MaxLength: 5}},
		},
		{
			name: "int without propertyInt",
			prop: Property{PropertyBase: PropertyBase{Name: "f", Type: PropertyTypeInt, MinLength: 0, MaxLength: 10}},
		},
		{
			name: "float without propertyFloat",
			prop: Property{PropertyBase: PropertyBase{Name: "f", Type: PropertyTypeFloat, MinLength: 0, MaxLength: 10}},
		},
		{
			name: "uuid without propertyUUID",
			prop: Property{PropertyBase: PropertyBase{Name: "f", Type: PropertyTypeUUID}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.Properties = []Property{tc.prop}
			if err := ValidateResponseConfig(c); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestValidateResponseConfig_FloatPrecisionBounds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		precision int
		wantErr   bool
	}{
		{name: "zero ok", precision: 0, wantErr: false},
		{name: "five ok", precision: 5, wantErr: false},
		{name: "negative rejected", precision: -1, wantErr: true},
		{name: "six rejected", precision: 6, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.Properties = []Property{makeFloatProperty("value", 0, 10, tc.precision, false)}
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestValidateResponseConfig_StringFunny(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		c := validConfig()
		c.Item.Properties = []Property{makeStringFunnyProperty("label")}
		if err := ValidateResponseConfig(c); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
	t.Run("ignores minLength greater than maxLength", func(t *testing.T) {
		t.Parallel()
		c := validConfig()
		prop := makeStringFunnyProperty("label")
		prop.MinLength = 100
		prop.MaxLength = 0
		c.Item.Properties = []Property{prop}
		if err := ValidateResponseConfig(c); err != nil {
			t.Errorf("expected no error for string-funny ignoring min/max, got: %v", err)
		}
	})
}

func TestValidateResponseConfig_UUIDVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version int
		wantErr bool
	}{
		{name: "version 1 ok", version: 1, wantErr: false},
		{name: "version 4 ok", version: 4, wantErr: false},
		{name: "version 7 ok", version: 7, wantErr: false},
		{name: "version 0 rejected", version: 0, wantErr: true},
		{name: "version 2 rejected", version: 2, wantErr: true},
		{name: "version 5 rejected", version: 5, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := validConfig()
			c.Item.Properties = []Property{makeUUIDProperty("id", tc.version)}
			err := ValidateResponseConfig(c)
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestParamResp_SetGetClearIsActive(t *testing.T) {
	t.Parallel()
	pr := NewParamResp()

	if pr.IsActive() {
		t.Error("expected inactive before Set")
	}
	if pr.Get() != nil {
		t.Error("expected nil before Set")
	}

	cfg := validConfig()
	pr.Set(cfg)

	if !pr.IsActive() {
		t.Error("expected active after Set")
	}
	if pr.Get() == nil {
		t.Error("expected non-nil after Set")
	}

	pr.Clear()

	if pr.IsActive() {
		t.Error("expected inactive after Clear")
	}
	if pr.Get() != nil {
		t.Error("expected nil after Clear")
	}
}
