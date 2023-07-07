package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type pgpPublicKeyType uint8

const (
	PGPPublicKeyType pgpPublicKeyType = iota
)

var (
	_ xattr.TypeWithValidate  = PGPPublicKeyType
	_ basetypes.StringTypable = PGPPublicKeyType

	_ basetypes.StringValuable                   = PGPPublicKey{}
	_ basetypes.StringValuableWithSemanticEquals = PGPPublicKey{}
)

// TerraformType returns the tftypes.Type that should be used to represent this
// framework type.
func (t pgpPublicKeyType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t pgpPublicKeyType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsUnknown() {
		return PGPPublicKeyUnknown(), nil
	}

	if in.IsNull() {
		return PGPPublicKeyNull(), nil
	}

	return PGPPublicKey{
		state: attr.ValueStateKnown,
		value: in.ValueString(),
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to
// convert the tftypes.Value into a more convenient Go type for the provider to
// consume the data with.
func (t pgpPublicKeyType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return PGPPublicKeyUnknown(), nil
	}

	if in.IsNull() {
		return PGPPublicKeyNull(), nil
	}

	var s string
	err := in.As(&s)

	if err != nil {
		return nil, err
	}

	return PGPPublicKey{
		state: attr.ValueStateKnown,
		value: s,
	}, nil
}

// ValueType returns the Value type.
func (t pgpPublicKeyType) ValueType(context.Context) attr.Value {
	return PGPPublicKey{}
}

// Equal returns true if `o` is also a PGPPublicKeyType.
func (t pgpPublicKeyType) Equal(o attr.Type) bool {
	_, ok := o.(pgpPublicKeyType)
	return ok
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t pgpPublicKeyType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

// String returns a human-friendly description of the PGPPublicKeyType.
func (t pgpPublicKeyType) String() string {
	return "types.PGPPublicKeyType"
}

// Validate implements type validation.
func (t pgpPublicKeyType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"PGPPublicKey Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected String value, received %T with value: %v", in, in),
		)

		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string

	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"PGPPublicKey Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Error: %s", err),
		)

		return diags
	}

	_, err = crypto.NewKeyFromArmored(value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid PGP Public Key",
			err.Error())

		return diags
	}

	return diags
}

func (t pgpPublicKeyType) Description() string {
	return `PGP Public key in ASCII-armor base64 encoded format.`
}

func PGPPublicKeyNull() PGPPublicKey {
	return PGPPublicKey{
		state: attr.ValueStateNull,
	}
}

func PGPPublicKeyUnknown() PGPPublicKey {
	return PGPPublicKey{
		state: attr.ValueStateUnknown,
	}
}

func PGPPublicKeyValue(value string) PGPPublicKey {
	return PGPPublicKey{
		state: attr.ValueStateKnown,
		value: value,
	}
}

type PGPPublicKey struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the original string representation.
	value string
}

// Type returns a PGPPublicKeyType.
func (k PGPPublicKey) Type(_ context.Context) attr.Type {
	return PGPPublicKeyType
}

// ToStringValue should convert the value type to a String.
func (k PGPPublicKey) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	switch k.state {
	case attr.ValueStateKnown:
		return types.StringValue(k.value), nil
	case attr.ValueStateNull:
		return types.StringNull(), nil
	case attr.ValueStateUnknown:
		return types.StringUnknown(), nil
	default:
		return types.StringUnknown(), diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled PGPPublicKey state in ToStringValue: %s", k.state), ""),
		}
	}
}

// ToTerraformValue returns the data contained in the *String as a string. If
// Unknown is true, it returns a tftypes.UnknownValue. If Null is true, it
// returns nil.
func (k PGPPublicKey) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := PGPPublicKeyType.TerraformType(ctx)

	switch k.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(t, k.value); err != nil {
			return tftypes.NewValue(t, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(t, k.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(t, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	default:
		return tftypes.NewValue(t, tftypes.UnknownValue), fmt.Errorf("unhandled PGPPublicKey state in ToTerraformValue: %s", k.state)
	}
}

// Equal returns true if `other` is a *PGPPublicKey and has the same value as `d`.
func (k PGPPublicKey) Equal(other attr.Value) bool {
	o, ok := other.(PGPPublicKey)

	if !ok {
		return false
	}

	if k.state != o.state {
		return false
	}

	if k.state != attr.ValueStateKnown {
		return true
	}

	return k.value == o.value
}

// IsNull returns true if the Value is not set, or is explicitly set to null.
func (k PGPPublicKey) IsNull() bool {
	return k.state == attr.ValueStateNull
}

// IsUnknown returns true if the Value is not yet known.
func (k PGPPublicKey) IsUnknown() bool {
	return k.state == attr.ValueStateUnknown
}

// String returns a summary representation of either the underlying Value,
// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
// or NullValueString (`<null>`) when IsNull() return true.
//
// This is an intentionally lossy representation, that are best suited for
// logging and error reporting, as they are not protected by
// compatibility guarantees within the framework.
func (k PGPPublicKey) String() string {
	if k.IsUnknown() {
		return attr.UnknownValueString
	}

	if k.IsNull() {
		return attr.NullValueString
	}

	return k.value
}

// ValuePGPPublicKey returns the known string value. If PGPPublicKey is null or unknown, returns "".
func (k PGPPublicKey) ValuePGPPublicKey() string {
	return k.value
}

// StringSemanticEquals should return true if the given value is
// semantically equal to the current value. This logic is used to prevent
// Terraform data consistency errors and resource drift where a value change
// may have inconsequential differences, such as spacing character removal
// in JSON formatted strings.
//
// Only known values are compared with this method as changing a value's
// state implicitly represents a different value.
func (k PGPPublicKey) StringSemanticEquals(ctx context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
	return strings.TrimSpace(k.value) == strings.TrimSpace(other.String()), nil
}
