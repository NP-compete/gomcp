package tools

import (
	"fmt"

	"github.com/NP-compete/gomcp/internal/logger"
)

// MultiplyNumbers multiplies two numbers with comprehensive tool metadata
//
// TOOL_NAME=multiply_numbers
// DISPLAY_NAME=Number Multiplication
// USECASE=Multiply two (floating point) numbers together
// INSTRUCTIONS=1. Provide two numeric values (int or float), 2. Call function, 3. Receive result
// INPUT_DESCRIPTION=Two parameters: a (number), b (number). Examples: (4, 5), (3.14, 2.0), (-1, 10)
// OUTPUT_DESCRIPTION=Dictionary with status, operation, input values (a, b), result, and message
// EXAMPLES=multiply_numbers(4, 5), multiply_numbers(3.14, 2.0)
// PREREQUISITES=None - standalone arithmetic operation
// RELATED_TOOLS=None - basic math operation
func MultiplyNumbers(args map[string]interface{}) (interface{}, error) {
	// Extract arguments
	aVal, aOk := args["a"]
	bVal, bOk := args["b"]

	if !aOk || !bOk {
		return map[string]interface{}{
			"status":  "error",
			"error":   "missing required arguments",
			"message": "Both 'a' and 'b' parameters are required",
		}, nil
	}

	// Convert to float64
	a, aErr := toFloat64(aVal)
	b, bErr := toFloat64(bVal)

	if aErr != nil || bErr != nil {
		return map[string]interface{}{
			"status":  "error",
			"error":   "invalid argument type",
			"message": "Both inputs must be numbers",
		}, nil
	}

	result := a * b

	logger.Info(fmt.Sprintf("Multiply tool called: %v * %v = %v", a, b, result))

	return map[string]interface{}{
		"status":    "success",
		"operation": "multiplication",
		"a":         a,
		"b":         b,
		"result":    result,
		"message":   fmt.Sprintf("Successfully multiplied %v and %v", a, b),
	}, nil
}

// toFloat64 converts various numeric types to float64
func toFloat64(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", val)
	}
}
