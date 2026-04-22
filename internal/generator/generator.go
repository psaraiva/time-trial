package generator

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"

	"github.com/google/uuid"
	"github.com/psaraiva/time-trial/internal/entities"
)

func Generate(config *entities.ResponseConfig) interface{} {
	if config.Item.IsCollection {
		items := make([]map[string]interface{}, config.Item.Quantity)
		for i := range items {
			items[i] = generateItem(config.Item.Properties)
		}
		return items
	}
	return generateItem(config.Item.Properties)
}

func generateItem(properties []entities.Property) map[string]interface{} {
	item := make(map[string]interface{}, len(properties))
	for _, prop := range properties {
		if !prop.IsRequired {
			item[prop.Name] = nil
			continue
		}
		item[prop.Name] = generateValue(prop)
	}
	return item
}

func generateValue(prop entities.Property) interface{} {
	switch prop.Type {
	case entities.PropertyTypeString:
		return generateString(prop)
	case entities.PropertyTypeInt:
		return generateInt(prop)
	case entities.PropertyTypeFloat:
		return generateFloat(prop)
	case entities.PropertyTypeUUID:
		return generateUUID(prop)
	case entities.PropertyTypeStringFunny:
		return generateStringFunny()
	}
	return nil
}

func generateStringFunny() string {
	adj := adjectives[rand.Intn(len(adjectives))]
	obj := objects[rand.Intn(len(objects))]
	return adj + "_" + obj
}

func generateUUID(prop entities.Property) string {
	switch prop.PropertyUUID.Version {
	case 1:
		id, err := uuid.NewUUID()
		if err != nil {
			return uuid.NewString()
		}
		return id.String()
	case 7:
		id, err := uuid.NewV7()
		if err != nil {
			return uuid.NewString()
		}
		return id.String()
	default:
		return uuid.NewString()
	}
}

func generateString(prop entities.Property) string {
	chars := []rune(prop.PropertyString.Chars)
	length := prop.MinLength

	if prop.MaxLength > prop.MinLength {
		length = prop.MinLength + rand.Intn(prop.MaxLength-prop.MinLength+1)
	}

	result := make([]rune, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}

func generateInt(prop entities.Property) int {
	min := prop.MinLength
	max := prop.MaxLength

	if !prop.PropertyInt.IsAcceptNegativeValue {
		if min < 0 {
			min = 0
		}
		if max < 0 {
			max = 0
		}
	}

	if max <= min {
		return min
	}

	return min + rand.Intn(max-min+1)
}

func generateFloat(prop entities.Property) json.RawMessage {
	min := float64(prop.MinLength)
	max := float64(prop.MaxLength)
	if !prop.PropertyFloat.IsAcceptNegativeValue {
		if min < 0 {
			min = 0
		}
		if max < 0 {
			max = 0
		}
	}

	var val float64
	if max <= min {
		val = min
	} else {
		val = min + rand.Float64()*(max-min)
	}

	precision := prop.PropertyFloat.FloatPrecision
	factor := math.Pow(10, float64(precision))
	val = math.Round(val*factor) / factor

	return json.RawMessage(fmt.Sprintf("%.*f", precision, val))
}
