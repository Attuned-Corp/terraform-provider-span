package dynamic

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This is an internal mapping of json to tf attributes.
// Attempts to resolve objects in depth.
func mapToJSON(b []byte) (attr.Type, attr.Value, error) {
	if string(b) == "null" {
		return types.DynamicType, types.DynamicNull(), nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(b, &object); err == nil {
		attrTypes := map[string]attr.Type{}
		attrVals := map[string]attr.Value{}
		for k, v := range object {
			attrTypes[k], attrVals[k], err = mapToJSON(v)
			if err != nil {
				return nil, nil, err
			}
		}
		typ := types.ObjectType{AttrTypes: attrTypes}
		val, diags := types.ObjectValue(attrTypes, attrVals)
		if diags.HasError() {
			return nil, nil, fmt.Errorf("Unexpected error mapping object from content: [%s]", string(b))
		}
		return typ, val, nil
	}

	var array []json.RawMessage
	if err := json.Unmarshal(b, &array); err == nil {
		eTypes := []attr.Type{}
		eVals := []attr.Value{}
		for _, e := range array {
			eType, eVal, err := mapToJSON(e)
			if err != nil {
				return nil, nil, err
			}
			eTypes = append(eTypes, eType)
			eVals = append(eVals, eVal)
		}
		typ := types.TupleType{ElemTypes: eTypes}
		val, diags := types.TupleValue(eTypes, eVals)
		if diags.HasError() {
			return nil, nil, fmt.Errorf("Unexpected error mapping tuple from content: [%s]", string(b))
		}
		return typ, val, nil
	}

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, nil, fmt.Errorf("Unexpected error unmarshalling content: [%s] -> [%v]", string(b), err)
	}

	switch v := v.(type) {
	case bool:
		return types.BoolType, types.BoolValue(v), nil
	case float64:
		return types.NumberType, types.NumberValue(big.NewFloat(v)), nil
	case string:
		return types.StringType, types.StringValue(v), nil
	case nil:
		return types.DynamicType, types.DynamicNull(), nil
	default:
		return nil, nil, fmt.Errorf("Encountered unknown JSON type: %T", v)
	}
}

// FromJSON maps serialized json string to TF dynamic type ad-hoc.
func FromJSON(b []byte) (types.Dynamic, error) {
	_, v, err := mapToJSON(b)
	if err != nil {
		return types.Dynamic{}, err
	}
	return types.DynamicValue(v), nil
}
