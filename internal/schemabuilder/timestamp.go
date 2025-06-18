package schemabuilder

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimestampField struct {
	*ProtoFieldExternal[TimestampField, *timestamppb.Timestamp]

	hasLtOrLte bool
	hasGtOrGte bool
}

func ProtoTimestamp(fieldNr uint) *TimestampField {
	options := make(map[string]string)

	gf := &TimestampField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[TimestampField, *timestamppb.Timestamp]{
		&protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.Timestamp", goType: "timestamp", imports: []string{"google/protobuf/timestamp.proto"}, options: options, isNonScalar: true}, gf,
	}
	return gf
}

func (tf *TimestampField) Within(t *timestamppb.Timestamp) *TimestampField {
	if t == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Within()' received a nil pointer."))
		return tf.self
	}
	tf.rules["within"] = t
	return tf.self
}

func (tf *TimestampField) Lt(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}
	if t == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Lt()' received a nil pointer."))
		return tf.self
	}
	tf.rules["lt"] = t
	tf.hasLtOrLte = true
	return tf.self
}

func (tf *TimestampField) Lte(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}
	if t == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Lte()' received a nil pointer."))
		return tf.self
	}
	tf.rules["lte"] = t
	tf.hasLtOrLte = true
	return tf.self
}

func (tf *TimestampField) LtNow() *TimestampField {
	if tf.hasLtOrLte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'lt', 'lt_now' and 'lte'."))
	}

	tf.rules["lt_now"] = true
	tf.hasLtOrLte = true
	return tf.self
}

func (tf *TimestampField) Gt(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}
	if t == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Gt()' received a nil pointer."))
		return tf.self
	}
	tf.rules["gt"] = t
	tf.hasGtOrGte = true
	return tf.self
}

func (tf *TimestampField) Gte(t *timestamppb.Timestamp) *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}
	if t == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Gte()' received a nil pointer."))
	}
	tf.rules["gte"] = t
	tf.hasGtOrGte = true
	return tf.self
}

func (tf *TimestampField) GtNow() *TimestampField {
	if tf.hasGtOrGte {
		tf.errors = append(tf.errors, fmt.Errorf("A timestamp field cannot have more than one rule between 'gt', 'gt_now' and 'gte'."))
	}

	tf.rules["gt_now"] = true
	tf.hasGtOrGte = true
	return tf.self
}

func (tf *TimestampField) Example(val *timestamppb.Timestamp) *TimestampField {
	if val == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Example()' received a nil pointer."))
		return tf.self
	}
	tf.repeatedOptions = append(tf.repeatedOptions, fmt.Sprintf("(buf.validate.field).timestamp.example = { seconds: %d }", val.GetSeconds()))
	return tf.self
}

func (tf *TimestampField) Const(val *timestamppb.Timestamp) *TimestampField {
	if val == nil {
		tf.errors = append(tf.errors, fmt.Errorf("'Const()' received a nil pointer."))
		return tf.self
	}
	tf.repeatedOptions = append(tf.repeatedOptions, fmt.Sprintf("(buf.validate.field).timestamp.const = { seconds: %d }", val.GetSeconds()))
	return tf.self
}
