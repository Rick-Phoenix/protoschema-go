package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

// An instance of a google.protobuf.Duration protobuf field.
type DurationField struct {
	*ProtoField[DurationField]

	hasLtOrLte bool
	hasGtOrGte bool
	in         []string
	notIn      []string

	gt  *time.Duration
	gte *time.Duration
	lt  *time.Duration
	lte *time.Duration
}

// Constructor for a google.protobuf.Duration protobuf field.
func Duration(name string) *DurationField {
	options := make(map[string]any)
	rules := make(map[string]any)

	gf := &DurationField{}
	gf.ProtoField = &ProtoField[DurationField]{
		protoFieldInternal: &protoFieldInternal{
			name:          name,
			protoType:     "google.protobuf.Duration",
			protoBaseType: "duration",
			goType:        "*durationpb.Duration",
			imports:       []string{"google/protobuf/duration.proto"},
			options:       options,
			isNonScalar:   true,
			rules:         rules,
			messageRef: &MessageSchema{
				ImportPath: "google/protobuf/duration.proto",
				Package: &ProtoPackage{
					GoPackagePath: "google.golang.org/protobuf/types/known/durationpb",
					GoPackageName: "durationpb",
					Name:          "Duration",
				},
			},
		},
		self: gf,
	}
	return gf
}

// Rule: this duration must be lower than the value indicated.
func (tf *DurationField) Lt(d string) *DurationField {
	if tf.hasLtOrLte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'lt' and 'lte'."))
	}
	duration, err := time.ParseDuration(d)
	if err != nil {
		tf.errors = errors.Join(tf.errors, err)
	}

	if tf.gt != nil && tf.gt.Seconds() >= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lt'."))
	}

	if tf.gte != nil && tf.gte.Seconds() >= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gte' cannot be larger than or equal to 'lt'."))
	}

	tf.rules["lt"] = d
	tf.hasLtOrLte = true
	tf.lt = &duration
	return tf.self
}

// Rule: this duration must be lower than or equal to the value indicated.
func (tf *DurationField) Lte(d string) *DurationField {
	if tf.hasLtOrLte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'lt' and 'lte'."))
	}

	duration, err := time.ParseDuration(d)
	if err != nil {
		tf.errors = errors.Join(tf.errors, err)
	}

	if tf.gt != nil && tf.gt.Seconds() >= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lte'."))
	}

	if tf.gte != nil && tf.gte.Seconds() > duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'gte' cannot be larger than 'lte'."))
	}

	tf.rules["lte"] = d
	tf.hasLtOrLte = true
	tf.lte = &duration
	return tf.self
}

// Rule: this duration must be higher than the value indicated.
func (tf *DurationField) Gt(d string) *DurationField {
	if tf.hasGtOrGte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'gt' and 'gte'."))
	}

	duration, err := time.ParseDuration(d)
	if err != nil {
		tf.errors = errors.Join(tf.errors, err)
	}

	if tf.lt != nil && tf.lt.Seconds() <= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gt'."))
	}

	if tf.lte != nil && tf.lte.Seconds() <= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lte' cannot be smaller than or equal to 'gt'."))
	}

	tf.rules["gt"] = d
	tf.hasGtOrGte = true
	tf.gt = &duration
	return tf.self
}

// Rule: this duration must be higher than or equal to the value indicated.
func (tf *DurationField) Gte(d string) *DurationField {
	if tf.hasGtOrGte {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'gt' and 'gte'."))
	}

	duration, err := time.ParseDuration(d)
	if err != nil {
		tf.errors = errors.Join(tf.errors, err)
	}

	if tf.lt != nil && tf.lt.Seconds() <= duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gte'."))
	}

	if tf.lte != nil && tf.lte.Seconds() < duration.Seconds() {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'lte' cannot be smaller than 'gte'."))
	}

	tf.rules["gte"] = d
	tf.hasGtOrGte = true
	tf.gte = &duration
	return tf.self
}

// Rule: the field's value must be among those listed in order to be accepted.
func (tf *DurationField) In(values ...string) *DurationField {
	for _, v := range values {
		err := validateDurationString(v)
		if err != nil {
			tf.errors = errors.Join(tf.errors, err)
		}
		if slices.Contains(tf.notIn, v) {
			tf.errors = errors.Join(tf.errors, fmt.Errorf("field %s cannot be inside of 'in' and 'not_in' at the same time.", v))
		}
	}

	tf.in = values
	tf.rules["in"] = values
	return tf.self
}

// Rule: the field's value must not be present among those listed in order to be accepted.
func (tf *DurationField) NotIn(values ...string) *DurationField {
	for _, v := range values {
		err := validateDurationString(v)
		if err != nil {
			tf.errors = errors.Join(tf.errors, err)
		}
		if slices.Contains(tf.in, v) {
			tf.errors = errors.Join(tf.errors, fmt.Errorf("field %s cannot be inside of 'in' and 'not_in' at the same time.", v))
		}
	}

	tf.notIn = values
	tf.rules["not_in"] = values
	return tf.self
}

// Rule: this field can only be this specific value. This will cause an error if it is used with other rules.
func (df *DurationField) Const(d string) *DurationField {
	err := validateDurationString(d)
	if err != nil {
		df.errors = errors.Join(df.errors, err)
	}
	df.protoFieldInternal.isConst = true
	df.rules["const"] = d
	return df.self
}

// An example value for this field. More than one example can be provided by calling this method multiple times.
func (df *DurationField) Example(val *durationpb.Duration) *DurationField {
	if val == nil {
		df.errors = errors.Join(df.errors, fmt.Errorf("'Example()' received a nil pointer."))
		return df.self
	}
	df.repeatedOptions = append(df.repeatedOptions, fmt.Sprintf("(buf.validate.field).duration.example = { seconds: %d }", val.GetSeconds()))
	return df.self
}
