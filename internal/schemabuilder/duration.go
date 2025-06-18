package schemabuilder

import (
	"fmt"
	"slices"

	"google.golang.org/protobuf/types/known/durationpb"
)

type DurationField struct {
	*ProtoFieldExternal[DurationField, *durationpb.Duration]

	hasLtOrLte bool
	hasGtOrGte bool
	in         []string
	notIn      []string
}

func ProtoDuration(fieldNr uint) *DurationField {
	options := make(map[string]string)

	gf := &DurationField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[DurationField, *durationpb.Duration]{
		&protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.Duration", goType: "durationpb.Duration", imports: []string{"google/protobuf/duration.proto"}, options: options, isNonScalar: true}, gf,
	}
	return gf
}

func (tf *DurationField) Lt(d string) *DurationField {
	if tf.hasLtOrLte {
		tf.errors = append(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'lt' and 'lte'."))
	}
	err := ValidateDurationString(d)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["lt"] = d
	tf.hasLtOrLte = true
	return tf.self
}

func (tf *DurationField) Lte(d string) *DurationField {
	if tf.hasLtOrLte {
		tf.errors = append(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'lt' and 'lte'."))
	}
	err := ValidateDurationString(d)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["lte"] = d
	tf.hasLtOrLte = true
	return tf.self
}

func (tf *DurationField) Gt(d string) *DurationField {
	if tf.hasGtOrGte {
		tf.errors = append(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'gt' and 'gte'."))
	}
	err := ValidateDurationString(d)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["gt"] = d
	tf.hasGtOrGte = true
	return tf.self
}

func (tf *DurationField) Gte(d string) *DurationField {
	if tf.hasGtOrGte {
		tf.errors = append(tf.errors, fmt.Errorf("A duration field cannot have more than one rule between 'gt' and 'gte'."))
	}
	err := ValidateDurationString(d)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["gte"] = d
	tf.hasGtOrGte = true
	return tf.self
}

func (tf *DurationField) In(values ...string) *DurationField {
	for _, v := range values {
		err := ValidateDurationString(v)
		if err != nil {
			tf.errors = append(tf.errors, err)
		}
		if slices.Contains(tf.notIn, v) {
			tf.errors = append(tf.errors, fmt.Errorf("field %s cannot be inside of 'in' and 'not_in' at the same time.", v))
			return tf.self
		}
	}
	list, err := formatProtoList(values)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["in"] = list
	return tf.self
}

func (tf *DurationField) NotIn(values ...string) *DurationField {
	for _, v := range values {
		err := ValidateDurationString(v)
		if err != nil {
			tf.errors = append(tf.errors, err)
		}
		if slices.Contains(tf.in, v) {
			tf.errors = append(tf.errors, fmt.Errorf("field %s cannot be inside of 'in' and 'not_in' at the same time.", v))
			return tf.self
		}
	}
	list, err := formatProtoList(values)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["not_in"] = list
	return tf.self
}

func (tf *DurationField) Const(d string) *DurationField {
	err := ValidateDurationString(d)
	if err != nil {
		tf.errors = append(tf.errors, err)
	}
	tf.rules["const"] = d
	return tf.self
}
