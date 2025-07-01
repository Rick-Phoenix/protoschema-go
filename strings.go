package schemabuilder

import (
	"errors"
	"fmt"
)

// An instance of a protobuf string field.
type StringField struct {
	*ProtoField[StringField]
	*ByteOrStringField[StringField, string]
	*ConstField[StringField, string, string]
	minBytes *uint
	maxBytes *uint
}

// A subtype of protobuf field, implemented by other field types.
type ByteOrStringField[BuilderT any, ValueT string | []byte] struct {
	internal         *protoFieldInternal
	self             *BuilderT
	hasWellKnownRule bool
	minLen           *uint
	maxLen           *uint
	*OptionalField[BuilderT]
}

// The constructor for a protobuf string field.
func String(name string) *StringField {
	rules := make(map[string]any)
	options := make(map[string]any)
	internal := &protoFieldInternal{name: name, protoType: "string", goType: "string", options: options, rules: rules}

	sf := &StringField{}
	sf.ProtoField = &ProtoField[StringField]{
		protoFieldInternal: internal,
		self:               sf,
	}
	sf.ByteOrStringField = &ByteOrStringField[StringField, string]{
		internal: internal,
		self:     sf,
	}
	sf.ConstField = &ConstField[StringField, string, string]{
		constInternal: internal,
		self:          sf,
	}
	sf.OptionalField = &OptionalField[StringField]{
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

// Rule: this string or bytes field must contain the indicated prefix.
func (b *ByteOrStringField[BuilderT, ValueT]) Prefix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["prefix"] = protoVal
	return b.self
}

// Rule: this string or bytes field must contain the indicated suffix.
func (b *ByteOrStringField[BuilderT, ValueT]) Suffix(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["suffix"] = protoVal
	return b.self
}

// Rule: this string or bytes field must contain the indicated value.
func (b *ByteOrStringField[BuilderT, ValueT]) Contains(s ValueT) *BuilderT {
	protoVal, err := formatProtoValue(s)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["contains"] = protoVal
	return b.self
}

// Rule: this string or bytes field must be a valid Ipv4 or Ipv6 address.
func (b *ByteOrStringField[BuilderT, ValueT]) Ip() *BuilderT {
	b.setWellKnownRule("ip", true)
	return b.self
}

// Rule: this string or bytes field must be a valid Ipv4 address.
func (b *ByteOrStringField[BuilderT, ValueT]) Ipv4() *BuilderT {
	b.setWellKnownRule("ipv4", true)
	return b.self
}

// Rule: this string or bytes field must be a valid Ipv6 address.
func (b *ByteOrStringField[BuilderT, ValueT]) Ipv6() *BuilderT {
	b.setWellKnownRule("ipv6", true)
	return b.self
}

// Rule: this string or bytes field must be of the minimum specified length.
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

// Rule: this string or bytes field must have a smaller length than the specified value.
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

// Rule: this string or bytes field must be of the exact specified length.
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

// Rule: this string or bytes field must match the specified regex.
func (b *ByteOrStringField[BuilderT, ValueT]) Pattern(regex string) *BuilderT {
	b.internal.rules["pattern"] = regex
	return b.self
}

// Rule: this string must have the exact specified byte length.
func (b *StringField) LenBytes(n uint) *StringField {
	if b.minBytes != nil {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use min_bytes and len_bytes together."))
	}
	if b.maxBytes != nil {
		b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("Cannot use max_bytes and len_bytes together."))
	}
	b.protoFieldInternal.rules["len_bytes"] = n
	return b
}

// Rule: this string must have a byte length that is larger than the indicated value.
func (b *StringField) MinBytes(n uint) *StringField {
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

// Rule: this string must have a byte length that is smaller than the indicated value.
func (b *StringField) MaxBytes(n uint) *StringField {
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

// Rule: this string must not contain the indicated string.
func (b *StringField) NotContains(s string) *StringField {
	b.protoFieldInternal.rules["not_contains"] = s
	return b
}

// Rule: this string must be an email address.
func (b *StringField) Email() *StringField {
	b.setWellKnownRule("email", true)
	return b
}

// Rule: this string must be a valid hostname.
func (b *StringField) Hostname() *StringField {
	b.setWellKnownRule("hostname", true)
	return b
}

// Rule: this string must be a valid uri.
func (b *StringField) URI() *StringField {
	b.setWellKnownRule("uri", true)
	return b
}

// Rule: this string must be either a uri or a relative filesystem path.
func (b *StringField) URIRef() *StringField {
	b.setWellKnownRule("uri_ref", true)
	return b
}

// Rule: this string must be an ip address or a hostname.
func (b *StringField) Address() *StringField {
	b.setWellKnownRule("address", true)
	return b
}

// Rule: this string must be a valid uuid.
func (b *StringField) UUID() *StringField {
	b.setWellKnownRule("uuid", true)
	return b
}

// Rule: this string must be a valid tuuid (trimmed uuid).
func (b *StringField) TUUID() *StringField {
	b.setWellKnownRule("tuuid", true)
	return b
}

// Rule: this string must be a valid ipv4 or ipv6 address with a prefix length (i.e. 10.0.0.1/24)
func (b *StringField) IpWithPrefixLen() *StringField {
	b.setWellKnownRule("ip_with_prefixlen", true)
	return b
}

// Rule: this string must be a valid ipv4 address with a prefix length (i.e. 10.0.0.1/24)
func (b *StringField) Ipv4WithPrefixLen() *StringField {
	b.setWellKnownRule("ipv4_with_prefixlen", true)
	return b
}

// Rule: this string must be a valid ipv6 address with a prefix length (i.e. 2001:0DB8:ABCD:0012::F1/64)
func (b *StringField) Ipv6WithPrefixLen() *StringField {
	b.setWellKnownRule("ipv6_with_prefixlen", true)
	return b
}

// Rule: this string must be an ipv4 or ipv6 prefix. All unmasked bits must be set to zero (i.e. 10.0.0.0/24)
func (b *StringField) IpPrefix() *StringField {
	b.setWellKnownRule("ip_prefix", true)
	return b
}

// Rule: this string must be an ipv4 prefix. All unmasked bits must be set to zero (i.e. 10.0.0.0/24)
func (b *StringField) Ipv4Prefix() *StringField {
	b.setWellKnownRule("ipv4_prefix", true)
	return b
}

// Rule: this string must be an ipv6 prefix. All unmasked bits must be set to zero (i.e. 2001:0DB8:ABCD:0012::0/64)
func (b *StringField) Ipv6Prefix() *StringField {
	b.setWellKnownRule("ipv6_prefix", true)
	return b
}

// Rule: this string must be a valid host with port (i.e. golang.org:443 or 127.0.0.1:3000)
func (b *StringField) HostAndPort() *StringField {
	b.setWellKnownRule("host_and_port", true)
	return b
}

// Rule: this string must be a valid HTTP header name as defined by [RFC 7230](https://datatracker.ietf.org/doc/html/rfc7230#section-3.2). While HTTPHeaderName accepts custom header names, this does not.
func (b *StringField) HTTPHeaderNameStrict() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	return b
}

// Rule: this string must be a valid HTTP header name. Unlike HTTPHeaderNameStrict, this also allows custom header names, as long as they don't contain the `\r\n\0` characters.
func (b *StringField) HTTPHeaderName() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_NAME")
	b.rules["strict"] = false
	return b
}

// Rule: this string must be a valid HTTP header value as defined by [RFC 7230](https://datatracker.ietf.org/doc/html/rfc7230#section-3.2.4). While HTTPHeaderValue accepts custom header values, this does not.
func (b *StringField) HTTPHeaderValueStrict() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	return b
}

// Rule: this string must be a valid HTTP header value. Unlike HTTPHeaderValueStrict, this also allows custom header values, as long as they don't contain the `\r\n\0` characters.
func (b *StringField) HTTPHeaderValue() *StringField {
	b.setWellKnownRule("well_known_regex", "KNOWN_REGEX_HTTP_HEADER_VALUE")
	b.rules["strict"] = false
	return b
}

// An instance of a protobuf bytes field.
type BytesField struct {
	*ProtoField[BytesField]
	*ByteOrStringField[BytesField, []byte]
	*ConstField[BytesField, []byte, byte]
}

// The constructor for a protobuf bytes field.
func Bytes(name string) *BytesField {
	rules := make(map[string]any)
	options := make(map[string]any)
	internal := &protoFieldInternal{name: name, protoType: "bytes", goType: "[]byte", options: options, rules: rules}

	bf := &BytesField{}
	bf.ProtoField = &ProtoField[BytesField]{
		protoFieldInternal: internal,
		self:               bf,
	}
	bf.ByteOrStringField = &ByteOrStringField[BytesField, []byte]{
		internal: internal,
		self:     bf,
	}
	bf.ConstField = &ConstField[BytesField, []byte, byte]{
		constInternal: internal,
		self:          bf,
	}
	bf.OptionalField = &OptionalField[BytesField]{
		optionalInternal: internal,
		self:             bf,
	}
	return bf
}
