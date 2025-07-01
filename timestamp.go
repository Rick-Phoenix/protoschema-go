package protoschema

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// An instance of a protobuf timestamp field.
type TimestampField struct {
	*ProtoField[TimestampField]

	hasLtOrLte bool
	hasGtOrGte bool

	lt  *timestamppb.Timestamp
	lte *timestamppb.Timestamp
	gt  *timestamppb.Timestamp
	gte *timestamppb.Timestamp
}

// The constructor for a timestamp field.
func Timestamp(name string) *TimestampField {
	rules := make(map[string]any)
	options := make(map[string]any)

	gf := &TimestampField{}
	gf.ProtoField = &ProtoField[TimestampField]{
		protoFieldInternal: &protoFieldInternal{
			name:          name,
			protoType:     "google.protobuf.Timestamp",
			goType:        "time.Time",
			protoBaseType: "timestamp",
			imports:       []string{"google/protobuf/timestamp.proto"},
			options:       options,
			isNonScalar:   true,
			rules:         rules,
			messageRef: &MessageSchema{
				Name:       "Timestamp",
				ImportPath: "google/protobuf/timestamp.proto",
				Package: &ProtoPackage{
					GoPackagePath: "google.golang.org/protobuf/types/known/timestamppb",
					Name:          "google.protobuf",
				},
			},
		},
		self: gf,
	}
	return gf
}

// Rule: this timestamp field must be within the selected duration.
func (tf *TimestampField) Within(t *durationpb.Duration) *TimestampField {
	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Within()' received a nil pointer."))
		return tf.self
	}

	tf.rules["within"] = t
	return tf.self
}

// Rule: this timestamp must be earlier than the selected timestamp.
func (tf *TimestampField) Lt(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}

	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Lt()' received a nil pointer."))
		return tf.self
	}

	if tf.gt != nil && tf.gt.GetSeconds() >= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lt'."))
	}

	if tf.gte != nil && tf.gte.GetSeconds() >= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gte' cannot be larger than or equal to 'lt'."))
	}

	tf.rules["lt"] = t
	tf.hasLtOrLte = true
	tf.lt = t
	return tf.self
}

// Rule: this timestamp field must be earlier than or equal to the selected timestamp.
func (tf *TimestampField) Lte(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}
	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Lte()' received a nil pointer."))
		return tf.self
	}

	if tf.gt != nil && tf.gt.GetSeconds() >= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lte'."))
	}

	if tf.gte != nil && tf.gte.GetSeconds() > t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gte' cannot be larger than 'lte'."))
	}

	tf.rules["lte"] = t
	tf.hasLtOrLte = true
	tf.lte = t
	return tf.self
}

// Rule: this timestamp must be in the past.
func (tf *TimestampField) LtNow() *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}

	now := &timestamppb.Timestamp{Seconds: time.Now().Unix()}

	if tf.gt != nil && tf.gt.GetSeconds() >= now.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lt_now'."))
	}

	if tf.gte != nil && tf.gte.GetSeconds() >= now.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gte' cannot be larger than or equal to 'lt_now'."))
	}

	tf.rules["lt_now"] = true
	tf.hasLtOrLte = true
	tf.lt = now
	return tf.self
}

// Rule: this timestamp must be later than the selected timestamp.
func (tf *TimestampField) Gt(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}
	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Gt()' received a nil pointer."))
		return tf.self
	}

	if tf.lt != nil && tf.lt.GetSeconds() <= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gt'."))
	}

	if tf.lte != nil && tf.lte.GetSeconds() <= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lte' cannot be smaller than or equal to 'gt'."))
	}

	tf.rules["gt"] = t
	tf.hasGtOrGte = true
	tf.gt = t
	return tf.self
}

// Rule: this timestamp must be later than or equal to the selected timestamp.
func (tf *TimestampField) Gte(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}
	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Gte()' received a nil pointer."))
	}

	if tf.lt != nil && tf.lt.GetSeconds() <= t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gte'."))
	}

	if tf.lte != nil && tf.lte.GetSeconds() < t.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lte' cannot be smaller than 'gte'."))
	}

	tf.rules["gte"] = t
	tf.hasGtOrGte = true
	tf.gte = t
	return tf.self
}

// Rule: this timestamp must be in the future.
func (tf *TimestampField) GtNow() *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}

	now := &timestamppb.Timestamp{Seconds: time.Now().Unix()}

	if tf.lt != nil && tf.lt.GetSeconds() <= now.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gt_now'."))
	}

	if tf.lte != nil && tf.lte.GetSeconds() <= now.GetSeconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lte' cannot be smaller than or equal to 'gt_now'."))
	}

	tf.rules["gt_now"] = true
	tf.hasGtOrGte = true
	tf.gt = now
	return tf.self
}

// An example value for this field. More than one example can be provided by calling this method multiple times.
func (tf *TimestampField) Example(val *timestamppb.Timestamp) *TimestampField {
	if val == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Example()' received a nil pointer."))
		return tf.self
	}
	tf.repeatedOptions = append(tf.repeatedOptions, fmt.Sprintf("(buf.validate.field).timestamp.example = { seconds: %d }", val.GetSeconds()))
	return tf.self
}

// Rule: this field can only be this specific value. This will cause an error if it is used with other rules.
func (tf *TimestampField) Const(val *timestamppb.Timestamp) *TimestampField {
	if val == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Const()' received a nil pointer."))
		return tf.self
	}
	tf.protoFieldInternal.isConst = true
	tf.rules["const"] = val
	return tf.self
}
