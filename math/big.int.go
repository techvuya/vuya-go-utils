package mathutils

import (
	"encoding/json"
	"fmt"
	"math/big"
)

type BigInt struct {
	big.Int
}

// Scan implements the sql.Scanner interface
func (bi *BigInt) Scan(src interface{}) error {
	switch src := src.(type) {
	case int64:
		bi.SetInt64(src)
	case []byte:
		bi.SetString(string(src), 10)
	default:
		return fmt.Errorf("not supported")
	}
	return nil
}

func (bi BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(bi.String())
}

func IsZeroString(s string) bool {
	n, success := new(big.Int).SetString(s, 10)
	if !success {
		// If the string cannot be parsed as a number, it's not zero
		return false
	}
	return n.Sign() == 0
}

// AddFloatAsBigInt adds two float64 values by converting them to big.Int (truncating decimals),
// performing addition, and returning the result as float64.
func AddFloatAsBigInt(f1, f2 float64) float64 {
	decimalsPlaces := int64(2)
	// Scale factor (e.g., 100 for 2 decimal places, 1000 for 3, etc.)
	scale := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimalsPlaces), nil))

	// Convert float64 to *big.Float and scale up
	bf1 := new(big.Float).SetFloat64(f1)
	bf2 := new(big.Float).SetFloat64(f2)

	scaled1 := new(big.Float).Mul(bf1, scale)
	scaled2 := new(big.Float).Mul(bf2, scale)

	// Convert to big.Int (now the "decimal" part is in the integer)
	bi1 := new(big.Int)
	scaled1.Int(bi1)

	bi2 := new(big.Int)
	scaled2.Int(bi2)

	// Add the scaled integers
	sum := new(big.Int).Add(bi1, bi2)

	// Convert back to big.Float and scale down
	result := new(big.Float).SetInt(sum)
	result.Quo(result, scale)

	// Convert to float64
	fResult, _ := result.Float64()
	return fResult
}
