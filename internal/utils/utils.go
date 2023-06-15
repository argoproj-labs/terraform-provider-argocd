package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func OptionalInt64(value *int64) basetypes.Int64Value {
	if value == nil {
		return types.Int64Null()
	}

	return types.Int64Value(*value)
}

func OptionalString(value *string) basetypes.StringValue {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(*value)
}

func OptionalTimeString(value *metav1.Time) basetypes.StringValue {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(value.String())
}

// MapMap will return a new map where each element has been mapped (transformed).
// The number of elements returned will always be the same as the input.
//
// Be careful when using this with maps of pointers. If you modify the input
// value it will affect the original slice. Be sure to return a new allocated
// object or deep copy the existing one.
//
// Based on pie.Map (which only works on slices) at https://github.com/elliotchance/pie/blob/a9ee294da00683bd3f44e8b35bc1deb1dad8fbda/v2/map.go#L3-L20
func MapMap[T comparable, U any, V any](ss map[T]U, fn func(U) V) map[T]V {
	if ss == nil {
		return nil
	}

	ss2 := make(map[T]V)
	for k, v := range ss {
		ss2[k] = fn(v)
	}

	return ss2
}
