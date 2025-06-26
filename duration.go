package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
	"time"
)

type ProtoDurationField struct {
	*ProtoFieldExternal[ProtoDurationField]

	hasLtOrLte bool
	hasGtOrGte bool
	in         []string
	notIn      []string

	gt  *time.Duration
	gte *time.Duration
	lt  *time.Duration
	lte *time.Duration
}

func Duration(name string) *ProtoDurationField {
	options := make(map[string]any)
	rules := make(map[string]any)

	gf := &ProtoDurationField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[ProtoDurationField]{
		&protoFieldInternal{name: name, protoType: "google.protobuf.Duration", protoBaseType: "duration", goType: "*durationpb.Duration", imports: []string{"google/protobuf/duration.proto"}, options: options, isNonScalar: true, rules: rules}, gf,
	}
	return gf
}

func (tf *ProtoDurationField) Lt(d string) *ProtoDurationField {
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

func (tf *ProtoDurationField) Lte(d string) *ProtoDurationField {
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

func (tf *ProtoDurationField) Gt(d string) *ProtoDurationField {
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

func (tf *ProtoDurationField) Gte(d string) *ProtoDurationField {
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

func (tf *ProtoDurationField) In(values ...string) *ProtoDurationField {
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

func (tf *ProtoDurationField) NotIn(values ...string) *ProtoDurationField {
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

func (tf *ProtoDurationField) Const(d string) *ProtoDurationField {
	err := validateDurationString(d)
	if err != nil {
		tf.errors = errors.Join(tf.errors, err)
	}
	tf.protoFieldInternal.isConst = true
	tf.rules["const"] = d
	return tf.self
}
