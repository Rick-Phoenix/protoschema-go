package schemabuilder

import (
	"fmt"
	"strconv"
)

type StringField struct {
	*ProtoFieldExternal[StringField, string]
	*ByteOrStringField[StringField, string]
}

type ByteOrStringField[BuilderT any, ValueT string | []byte] struct {
	internal         *protoFieldInternal
	self             *BuilderT
	hasWellKnownRule bool
}

func (b *ByteOrStringField[BuilderT, ValueT]) setWellKnownRule(ruleName string, ruleValue any) {
	if b.hasWellKnownRule {
		b.internal.errors = append(b.internal.errors, fmt.Errorf("A string field can only have one well-known rule (e.g., email, hostname, ip, etc.)"))
		return
	}
	b.internal.rules[ruleName] = ruleValue
	b.hasWellKnownRule = true
}

func (b *ByteOrStringField[BuilderT, ValueT]) Prefix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["prefix"] = protoVal
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Suffix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["suffix"] = protoVal
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Contains(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["contains"] = protoVal
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Ip() *BuilderT {
	b.setWellKnownRule("ip", true)
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Ipv4() *BuilderT {
	b.setWellKnownRule("ipv4", true)
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Ipv6() *BuilderT {
	b.setWellKnownRule("ipv6", true)
	return b.self
}

func (l *ByteOrStringField[BuilderT, ValueT]) MinLen(n int) *BuilderT {
	l.internal.options["(buf.validate.field)."+l.internal.protoType+".min_len"] = strconv.Itoa(n)
	l.internal.rules["min_len"] = n
	return l.self
}

func (l *ByteOrStringField[BuilderT, ValueT]) MaxLen(n int) *BuilderT {
	l.internal.options["(buf.validate.field)."+l.internal.protoType+".max_len"] = strconv.Itoa(n)
	l.internal.rules["max_len"] = n
	return l.self
}

func (l *ByteOrStringField[BuilderT, ValueT]) Len(n int) *BuilderT {
	l.internal.options["(buf.validate.field)."+l.internal.protoType+".len"] = strconv.Itoa(n)
	l.internal.rules["len"] = n
	return l.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Pattern(regex string) *BuilderT {
	b.internal.rules["pattern"] = regex
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) In(values ...string) *BuilderT {
	list, err := formatProtoList(values)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["in"] = list
	return b.self
}
func (b *ByteOrStringField[BuilderT, ValueT]) NotIn(values ...string) *BuilderT {
	list, err := formatProtoList(values)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["not_in"] = list
	return b.self
}

func ProtoString(fieldNumber uint) *StringField {
	imports := make(Set)
	rules := make(map[string]any)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, protoType: "string", goType: "string", imports: imports, options: options, rules: rules}

	sf := &StringField{}
	sf.ProtoFieldExternal = &ProtoFieldExternal[StringField, string]{
		protoFieldInternal: internal,
		self:               sf,
	}
	sf.ByteOrStringField = &ByteOrStringField[StringField, string]{
		internal: internal,
		self:     sf,
	}
	return sf
}

type BytesField struct {
	*ProtoFieldExternal[BytesField, []byte]
	*ByteOrStringField[BytesField, []byte]
}

func ProtoBytes(fieldNumber uint) *BytesField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, protoType: "bytes", goType: "[]byte", imports: imports, options: options}

	bf := &BytesField{}
	bf.ProtoFieldExternal = &ProtoFieldExternal[BytesField, []byte]{
		protoFieldInternal: internal,
		self:               bf,
	}
	bf.ByteOrStringField = &ByteOrStringField[BytesField, []byte]{
		internal: internal,
		self:     bf,
	}
	return bf
}

func (b *StringField) LenBytes(n int) *StringField {
	b.internal.rules["len_bytes"] = n
	return b
}
func (b *StringField) MinBytes(n int) *StringField {
	b.internal.rules["min_bytes"] = n
	return b
}

func (b *StringField) MaxBytes(n int) *StringField {
	b.internal.rules["max_bytes"] = n
	return b
}

func (b *StringField) NotContains(s string) *StringField {
	b.internal.rules["not_contains"] = s
	return b
}

func (b *StringField) Email() *StringField {
	b.setWellKnownRule("email", true)
	return b
}
func (b *StringField) Hostname() *StringField {
	b.setWellKnownRule("hostname", true)
	return b
}

func (b *StringField) URI() *StringField {
	b.setWellKnownRule("uri", true)
	return b
}
func (b *StringField) URIRef() *StringField {
	b.setWellKnownRule("uri_ref", true)
	return b
}
func (b *StringField) Address() *StringField {
	b.setWellKnownRule("address", true)
	return b
}
func (b *StringField) UUID() *StringField {
	b.setWellKnownRule("uuid", true)
	return b
}
func (b *StringField) TUUID() *StringField {
	b.setWellKnownRule("tuuid", true)
	return b
}
func (b *StringField) IpWithMask() *StringField {
	b.setWellKnownRule("ip_with_prefixlen", true)
	return b
}
func (b *StringField) Ipv4WithMask() *StringField {
	b.setWellKnownRule("ipv4_with_prefixlen", true)
	return b
}
func (b *StringField) Ipv6WithMask() *StringField {
	b.setWellKnownRule("ipv6_with_prefixlen", true)
	return b
}
func (b *StringField) IpPrefix() *StringField {
	b.setWellKnownRule("ip_prefix", true)
	return b
}
func (b *StringField) Ipv4Prefix() *StringField {
	b.setWellKnownRule("ipv4_prefix", true)
	return b
}
func (b *StringField) Ipv6Prefix() *StringField {
	b.setWellKnownRule("ipv6_prefix", true)
	return b
}
func (b *StringField) HostAndPort() *StringField {
	b.setWellKnownRule("host_and_port", true)
	return b
}

func (b *StringField) HttpHeaderNameStrict() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	return b
}

func (b *StringField) HttpHeaderName() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	b.rules["strict"] = false
	return b
}

func (b *StringField) HttpHeaderValueStrict() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	return b
}

func (b *StringField) HttpHeaderValue() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	b.rules["strict"] = false
	return b
}
