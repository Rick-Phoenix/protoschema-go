package schemabuilder

import (
	"errors"
	"fmt"
)

type ProtoStringField struct {
	*ProtoFieldExternal[ProtoStringField]
	*ByteOrStringField[ProtoStringField, string]
	*ProtoConstField[ProtoStringField, string, string]
	minBytes *uint
	maxBytes *uint
}

type ByteOrStringField[BuilderT any, ValueT string | []byte] struct {
	internal         *protoFieldInternal
	self             *BuilderT
	hasWellKnownRule bool
	minLen           *uint
	maxLen           *uint
	*ProtoOptionalField[BuilderT]
}

func String(name string) *ProtoStringField {
	rules := make(map[string]any)
	options := make(map[string]any)
	internal := &protoFieldInternal{name: name, protoType: "string", goType: "string", options: options, rules: rules}

	sf := &ProtoStringField{}
	sf.ProtoFieldExternal = &ProtoFieldExternal[ProtoStringField]{
		protoFieldInternal: internal,
		self:               sf,
	}
	sf.ByteOrStringField = &ByteOrStringField[ProtoStringField, string]{
		internal: internal,
		self:     sf,
	}
	sf.ProtoConstField = &ProtoConstField[ProtoStringField, string, string]{
		constInternal: internal,
		self:          sf,
	}
	sf.ProtoOptionalField = &ProtoOptionalField[ProtoStringField]{
		optionalInternal: internal,
		self:             sf,
	}

	return sf
}

func (b *ByteOrStringField[BuilderT, ValueT]) setWellKnownRule(ruleName string, ruleValue any) {
	if b.hasWellKnownRule {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("A string field can only have one well-known rule (e.g., email, hostname, ip, etc.)"))
		return
	}
	b.internal.rules[ruleName] = ruleValue
	b.hasWellKnownRule = true
}

func (b *ByteOrStringField[BuilderT, ValueT]) Prefix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["prefix"] = protoVal
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Suffix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["suffix"] = protoVal
	return b.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Contains(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
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

func (l *ByteOrStringField[BuilderT, ValueT]) MinLen(n uint) *BuilderT {
	if _, exists := l.internal.rules["len"]; exists {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("Cannot use min_len and len together."))
	}
	if l.maxLen != nil && *l.maxLen < n {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("max_len cannot be smaller than min_len."))
	}
	l.minLen = &n
	l.internal.rules["min_len"] = n
	return l.self
}

func (l *ByteOrStringField[BuilderT, ValueT]) MaxLen(n uint) *BuilderT {
	if _, exists := l.internal.rules["len"]; exists {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("Cannot use max_len and len together."))
	}
	if l.minLen != nil && *l.minLen > n {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("max_len cannot be smaller than min_len."))
	}
	l.maxLen = &n
	l.internal.rules["max_len"] = n
	return l.self
}

func (l *ByteOrStringField[BuilderT, ValueT]) Len(n uint) *BuilderT {
	if l.minLen != nil {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("Cannot use min_len and len together."))
	}
	if l.maxLen != nil {
		l.internal.errors = errors.Join(l.internal.errors, fmt.Errorf("Cannot use max_len and len together."))
	}
	l.internal.rules["len"] = n
	return l.self
}

func (b *ByteOrStringField[BuilderT, ValueT]) Pattern(regex string) *BuilderT {
	b.internal.rules["pattern"] = regex
	return b.self
}

func (b *ProtoStringField) LenBytes(n uint) *ProtoStringField {
	if b.minBytes != nil {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use min_bytes and len_bytes together."))
	}
	if b.maxBytes != nil {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use max_bytes and len_bytes together."))
	}
	b.protoFieldInternal.rules["len_bytes"] = n
	return b
}

func (b *ProtoStringField) MinBytes(n uint) *ProtoStringField {
	if _, exists := b.internal.rules["len_bytes"]; exists {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use min_bytes and len_bytes together."))
	}
	if b.maxBytes != nil && *b.maxBytes < n {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("min_bytes cannot be larger than max_bytes."))
	}
	b.minBytes = &n
	b.protoFieldInternal.rules["min_bytes"] = n
	return b
}

func (b *ProtoStringField) MaxBytes(n uint) *ProtoStringField {
	if _, exists := b.internal.rules["len_bytes"]; exists {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use max_bytes and len_bytes together."))
	}
	if b.minBytes != nil && *b.minBytes > n {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("min_bytes cannot be larger than max_bytes."))
	}
	b.maxBytes = &n
	b.protoFieldInternal.rules["max_bytes"] = n
	return b
}

func (b *ProtoStringField) NotContains(s string) *ProtoStringField {
	b.protoFieldInternal.rules["not_contains"] = s
	return b
}

func (b *ProtoStringField) Email() *ProtoStringField {
	b.setWellKnownRule("email", true)
	return b
}

func (b *ProtoStringField) Hostname() *ProtoStringField {
	b.setWellKnownRule("hostname", true)
	return b
}

func (b *ProtoStringField) URI() *ProtoStringField {
	b.setWellKnownRule("uri", true)
	return b
}

func (b *ProtoStringField) URIRef() *ProtoStringField {
	b.setWellKnownRule("uri_ref", true)
	return b
}

func (b *ProtoStringField) Address() *ProtoStringField {
	b.setWellKnownRule("address", true)
	return b
}

func (b *ProtoStringField) UUID() *ProtoStringField {
	b.setWellKnownRule("uuid", true)
	return b
}

func (b *ProtoStringField) TUUID() *ProtoStringField {
	b.setWellKnownRule("tuuid", true)
	return b
}

func (b *ProtoStringField) IpWithMask() *ProtoStringField {
	b.setWellKnownRule("ip_with_prefixlen", true)
	return b
}

func (b *ProtoStringField) Ipv4WithMask() *ProtoStringField {
	b.setWellKnownRule("ipv4_with_prefixlen", true)
	return b
}

func (b *ProtoStringField) Ipv6WithMask() *ProtoStringField {
	b.setWellKnownRule("ipv6_with_prefixlen", true)
	return b
}

func (b *ProtoStringField) IpPrefix() *ProtoStringField {
	b.setWellKnownRule("ip_prefix", true)
	return b
}

func (b *ProtoStringField) Ipv4Prefix() *ProtoStringField {
	b.setWellKnownRule("ipv4_prefix", true)
	return b
}

func (b *ProtoStringField) Ipv6Prefix() *ProtoStringField {
	b.setWellKnownRule("ipv6_prefix", true)
	return b
}

func (b *ProtoStringField) HostAndPort() *ProtoStringField {
	b.setWellKnownRule("host_and_port", true)
	return b
}

func (b *ProtoStringField) HttpHeaderNameStrict() *ProtoStringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	return b
}

func (b *ProtoStringField) HttpHeaderName() *ProtoStringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	b.rules["strict"] = false
	return b
}

func (b *ProtoStringField) HttpHeaderValueStrict() *ProtoStringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	return b
}

func (b *ProtoStringField) HttpHeaderValue() *ProtoStringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	b.rules["strict"] = false
	return b
}

type BytesField struct {
	*ProtoFieldExternal[BytesField]
	*ByteOrStringField[BytesField, []byte]
	*ProtoConstField[BytesField, []byte, byte]
}

func Bytes(name string) *BytesField {
	rules := make(map[string]any)
	options := make(map[string]any)
	internal := &protoFieldInternal{name: name, protoType: "bytes", goType: "[]byte", options: options, rules: rules}

	bf := &BytesField{}
	bf.ProtoFieldExternal = &ProtoFieldExternal[BytesField]{
		protoFieldInternal: internal,
		self:               bf,
	}
	bf.ByteOrStringField = &ByteOrStringField[BytesField, []byte]{
		internal: internal,
		self:     bf,
	}
	bf.ProtoConstField = &ProtoConstField[BytesField, []byte, byte]{
		constInternal: internal,
		self:          bf,
	}
	bf.ProtoOptionalField = &ProtoOptionalField[BytesField]{
		optionalInternal: internal,
		self:             bf,
	}
	return bf
}
