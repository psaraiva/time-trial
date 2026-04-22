package generator

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/psaraiva/time-trial/internal/entities"
)

func makeStringProp(name string, min, max int, chars string, required bool) entities.Property {
	return entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       name,
			Type:       entities.PropertyTypeString,
			IsRequired: required,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyString: &entities.PropertyStringConfig{Chars: chars},
	}
}

func makeIntProp(name string, min, max int, acceptNeg bool) entities.Property {
	return entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       name,
			Type:       entities.PropertyTypeInt,
			IsRequired: true,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyInt: &entities.PropertyIntConfig{IsAcceptNegativeValue: acceptNeg},
	}
}

func makeFloatProp(name string, min, max, precision int, acceptNeg bool) entities.Property {
	return entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       name,
			Type:       entities.PropertyTypeFloat,
			IsRequired: true,
			MinLength:  min,
			MaxLength:  max,
		},
		PropertyFloat: &entities.PropertyFloatConfig{
			FloatPrecision:        precision,
			IsAcceptNegativeValue: acceptNeg,
		},
	}
}

func TestGenerate_SingleItem(t *testing.T) {
	t.Parallel()
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item: entities.ItemConfig{
			IsCollection: false,
			Properties:   []entities.Property{makeStringProp("name", 3, 8, "abc", true)},
		},
	}
	result := Generate(config)
	item, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", result)
	}
	if _, exists := item["name"]; !exists {
		t.Error("expected 'name' key in item")
	}
}

func TestGenerate_Collection(t *testing.T) {
	t.Parallel()
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item: entities.ItemConfig{
			IsCollection: true,
			Quantity:     5,
			Properties:   []entities.Property{makeStringProp("name", 1, 5, "abc", true)},
		},
	}
	result := Generate(config)
	items, ok := result.([]map[string]interface{})
	if !ok {
		t.Fatalf("expected []map[string]interface{}, got %T", result)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}
}

func TestGenerate_StringLength(t *testing.T) {
	t.Parallel()
	min, max := 3, 8
	prop := makeStringProp("s", min, max, "abcABC", true)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}

	for i := 0; i < 50; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		s, ok := item["s"].(string)
		if !ok {
			t.Fatalf("expected string value")
		}
		if len(s) < min || len(s) > max {
			t.Errorf("string length %d out of range [%d,%d]", len(s), min, max)
		}
	}
}

func TestGenerate_StringCharsOnly(t *testing.T) {
	t.Parallel()
	chars := "abc"
	prop := makeStringProp("s", 5, 10, chars, true)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}

	for i := 0; i < 30; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		s := item["s"].(string)
		for _, r := range s {
			if !strings.ContainsRune(chars, r) {
				t.Errorf("generated string contains unexpected char %q", r)
			}
		}
	}
}

func TestGenerate_IntRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		min, max  int
		acceptNeg bool
	}{
		{name: "positive range", min: 5, max: 20, acceptNeg: false},
		{name: "zero-based", min: 0, max: 100, acceptNeg: false},
		{name: "negative clamped", min: -50, max: -10, acceptNeg: false},
		{name: "negative allowed", min: -50, max: -10, acceptNeg: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prop := makeIntProp("v", tc.min, tc.max, tc.acceptNeg)
			config := &entities.ResponseConfig{
				StatusCode: 200,
				Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
			}
			for i := 0; i < 30; i++ {
				result := Generate(config)
				item := result.(map[string]interface{})
				v, ok := item["v"].(int)
				if !ok {
					t.Fatalf("expected int value")
				}
				effectiveMin := tc.min
				effectiveMax := tc.max
				if !tc.acceptNeg {
					if effectiveMin < 0 {
						effectiveMin = 0
					}
					if effectiveMax < 0 {
						effectiveMax = 0
					}
				}
				if v < effectiveMin || v > effectiveMax {
					t.Errorf("int value %d out of range [%d,%d]", v, effectiveMin, effectiveMax)
				}
			}
		})
	}
}

func TestGenerate_FloatPrecision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		precision int
	}{
		{name: "zero decimals", precision: 0},
		{name: "two decimals", precision: 2},
		{name: "three decimals", precision: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			prop := makeFloatProp("v", 0, 1000, tc.precision, false)
			config := &entities.ResponseConfig{
				StatusCode: 200,
				Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
			}
			for i := 0; i < 30; i++ {
				result := Generate(config)
				item := result.(map[string]interface{})
				raw, ok := item["v"].(json.RawMessage)
				if !ok {
					t.Fatalf("expected json.RawMessage, got %T", item["v"])
				}
				s := string(raw)
				dot := strings.Index(s, ".")
				if tc.precision == 0 {
					if dot != -1 {
						t.Errorf("precision 0: unexpected decimal point in %q", s)
					}
				} else {
					if dot == -1 {
						t.Errorf("precision %d: expected decimal point in %q", tc.precision, s)
						continue
					}
					if got := len(s) - dot - 1; got != tc.precision {
						t.Errorf("precision %d: expected %d decimal places, got %d in %q", tc.precision, tc.precision, got, s)
					}
				}
			}
		})
	}
}

func TestGenerate_NotRequired_IsNull(t *testing.T) {
	t.Parallel()
	prop := makeStringProp("optional", 1, 5, "abc", false)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	result := Generate(config)
	item := result.(map[string]interface{})
	val, exists := item["optional"]
	if !exists {
		t.Error("expected 'optional' key to exist in item")
	}
	if val != nil {
		t.Errorf("expected nil for non-required field, got %v", val)
	}
}

func TestGenerate_FixedLength_MinEqualsMax(t *testing.T) {
	t.Parallel()
	prop := makeStringProp("s", 5, 5, "abc", true)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	for i := 0; i < 10; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		s := item["s"].(string)
		if len(s) != 5 {
			t.Errorf("expected string length 5, got %d", len(s))
		}
	}
}

func TestGenerateValue_UnknownType_ReturnsNil(t *testing.T) {
	t.Parallel()
	prop := entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       "x",
			Type:       entities.PropertyType("unknown"),
			IsRequired: true,
		},
	}
	result := generateValue(prop)
	if result != nil {
		t.Errorf("expected nil for unknown type, got %v", result)
	}
}

func TestGenerate_Float_NegativeMinClamped(t *testing.T) {
	t.Parallel()
	prop := makeFloatProp("v", -100, 50, 2, false)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	for i := 0; i < 20; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		raw := item["v"].(json.RawMessage)
		var val float64
		if err := json.Unmarshal(raw, &val); err != nil {
			t.Fatalf("failed to unmarshal float: %v", err)
		}
		if val < 0 {
			t.Errorf("expected non-negative value, got %f", val)
		}
	}
}

func TestGenerate_Float_NegativeMaxClamped(t *testing.T) {
	t.Parallel()
	prop := makeFloatProp("v", -100, -10, 2, false)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	result := Generate(config)
	item := result.(map[string]interface{})
	raw := item["v"].(json.RawMessage)
	var val float64
	if err := json.Unmarshal(raw, &val); err != nil {
		t.Fatalf("failed to unmarshal float: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 when both min and max are clamped to 0, got %f", val)
	}
}

func makeUUIDProp(name string, version int) entities.Property {
	return entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       name,
			Type:       entities.PropertyTypeUUID,
			IsRequired: true,
		},
		PropertyUUID: &entities.PropertyUUIDConfig{Version: version},
	}
}

func makeStringFunnyProp(name string) entities.Property {
	return entities.Property{
		PropertyBase: entities.PropertyBase{
			Name:       name,
			Type:       entities.PropertyTypeStringFunny,
			IsRequired: true,
		},
	}
}

var uuidRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var funnyRegexp = regexp.MustCompile(`^[a-z]+_[a-z]+$`)

func TestGenerate_UUID_Format(t *testing.T) {
	t.Parallel()
	for _, version := range []int{1, 4, 7} {
		version := version
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			t.Parallel()
			prop := makeUUIDProp("id", version)
			config := &entities.ResponseConfig{
				StatusCode: 200,
				Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
			}
			for i := 0; i < 20; i++ {
				result := Generate(config)
				item := result.(map[string]interface{})
				s, ok := item["id"].(string)
				if !ok {
					t.Fatalf("expected string value, got %T", item["id"])
				}
				if !uuidRegexp.MatchString(s) {
					t.Errorf("v%d: %q is not a valid UUID format", version, s)
				}
			}
		})
	}
}

func TestGenerate_UUID_Version(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version      int
		expectNibble byte
	}{
		{version: 1, expectNibble: '1'},
		{version: 4, expectNibble: '4'},
		{version: 7, expectNibble: '7'},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("v%d_nibble", tc.version), func(t *testing.T) {
			t.Parallel()
			prop := makeUUIDProp("id", tc.version)
			config := &entities.ResponseConfig{
				StatusCode: 200,
				Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
			}
			for i := 0; i < 20; i++ {
				result := Generate(config)
				item := result.(map[string]interface{})
				s := item["id"].(string)
				// UUID format: xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx
				// version nibble is at index 14
				if len(s) < 15 {
					t.Fatalf("UUID too short: %q", s)
				}
				if s[14] != tc.expectNibble {
					t.Errorf("v%d: expected version nibble %c at index 14, got %c in %q", tc.version, tc.expectNibble, s[14], s)
				}
			}
		})
	}
}

func TestGenerate_StringFunny_Format(t *testing.T) {
	t.Parallel()
	prop := makeStringFunnyProp("label")
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	for i := 0; i < 50; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		s, ok := item["label"].(string)
		if !ok {
			t.Fatalf("expected string value, got %T", item["label"])
		}
		if !funnyRegexp.MatchString(s) {
			t.Errorf("generated value %q does not match <adjective>_<name> pattern", s)
		}
	}
}

func TestGenerate_StringFunny_Variety(t *testing.T) {
	t.Parallel()
	prop := makeStringFunnyProp("label")
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	seen := make(map[string]struct{})
	for i := 0; i < 200; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		seen[item["label"].(string)] = struct{}{}
	}
	if len(seen) < 5 {
		t.Errorf("expected variety in generated names, got only %d distinct values in 200 runs", len(seen))
	}
}

func TestGenerate_Float_MaxEqualMin(t *testing.T) {
	t.Parallel()
	prop := makeFloatProp("v", 7, 7, 2, false)
	config := &entities.ResponseConfig{
		StatusCode: 200,
		Item:       entities.ItemConfig{Properties: []entities.Property{prop}},
	}
	for i := 0; i < 10; i++ {
		result := Generate(config)
		item := result.(map[string]interface{})
		raw := item["v"].(json.RawMessage)
		var val float64
		if err := json.Unmarshal(raw, &val); err != nil {
			t.Fatalf("failed to unmarshal float: %v", err)
		}
		if val != 7.0 {
			t.Errorf("expected 7.0 when min == max, got %f", val)
		}
	}
}
