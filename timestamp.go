package schemabuilder

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProtoTimestampField struct {
	*ProtoFieldExternal[ProtoTimestampField]

	hasLtOrLte bool
	hasGtOrGte bool

	lt  *timestamppb.Timestamp
	lte *timestamppb.Timestamp
	gt  *timestamppb.Timestamp
	gte *timestamppb.Timestamp
}

func Timestamp(name string) *ProtoTimestampField {
	rules := make(map[string]any)
	options := make(map[string]any)

	gf := &ProtoTimestampField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[ProtoTimestampField]{
		&protoFieldInternal{name: name, protoType: "google.protobuf.Timestamp", goType: "time.Time", protoBaseType: "timestamp", imports: []string{"google/protobuf/timestamp.proto"}, options: options, isNonScalar: true, rules: rules}, gf,
	}
	return gf
}

func (tf *ProtoTimestampField) Within(t *durationpb.Duration) *ProtoTimestampField {
	if t == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Within()' received a nil pointer."))
		return tf.self
	}

	tf.rules["within"] = t
	return tf.self
}

func (tf *ProtoTimestampField) Lt(t *timestamppb.Timestamp) *ProtoTimestampField {
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

func (tf *ProtoTimestampField) Lte(t *timestamppb.Timestamp) *ProtoTimestampField {
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

func (tf *ProtoTimestampField) LtNow() *ProtoTimestampField {
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

func (tf *ProtoTimestampField) Gt(t *timestamppb.Timestamp) *ProtoTimestampField {
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

func (tf *ProtoTimestampField) Gte(t *timestamppb.Timestamp) *ProtoTimestampField {
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

func (tf *ProtoTimestampField) GtNow() *ProtoTimestampField {
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

func (tf *ProtoTimestampField) Example(val *timestamppb.Timestamp) *ProtoTimestampField {
	if val == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Example()' received a nil pointer."))
		return tf.self
	}
	tf.repeatedOptions = append(tf.repeatedOptions, fmt.Sprintf("(buf.validate.field).timestamp.example = { seconds: %d }", val.GetSeconds()))
	return tf.self
}

func (tf *ProtoTimestampField) Const(val *timestamppb.Timestamp) *ProtoTimestampField {
	if val == nil {
		tf.errors = errors.Join(tf.errors, fmt.Errorf("'Const()' received a nil pointer."))
		return tf.self
	}
	tf.protoFieldInternal.isConst = true
	tf.rules["const"] = val
	return tf.self
}
