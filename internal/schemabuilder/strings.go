package schemabuilder

import "fmt"

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

func (b *StringField) Pattern(regex string) *StringField {
	b.internal.rules["pattern"] = regex
	return b
}

func (b *StringField) Prefix(s string) *StringField {
	b.internal.rules["prefix"] = s
	return b
}
func (b *StringField) Suffix(s string) *StringField {
	b.internal.rules["suffix"] = s
	return b
}
func (b *StringField) Contains(s string) *StringField {
	b.internal.rules["contains"] = s
	return b
}
func (b *StringField) NotContains(s string) *StringField {
	b.internal.rules["not_contains"] = s
	return b
}

func (b *StringField) In(values ...string) *StringField {
	b.internal.rules["in"] = values
	return b
}
func (b *StringField) NotIn(values ...string) *StringField {
	b.internal.rules["not_in"] = values
	return b
}

func (b *StringField) setWellKnownRule(ruleName string) {
	if b.hasWellKnownRule {
		b.internal.errors = append(b.internal.errors, fmt.Errorf("A string field can only have one well-known rule (e.g., email, hostname, ip, etc.)"))
		return
	}
	b.internal.rules[ruleName] = true
	b.hasWellKnownRule = true
}

func (b *StringField) Email() *StringField {
	b.setWellKnownRule("email")
	return b
}
func (b *StringField) Hostname() *StringField {
	b.setWellKnownRule("hostname")
	return b
}
func (b *StringField) IP() *StringField {
	b.setWellKnownRule("ip")
	return b
}
func (b *StringField) IPv4() *StringField {
	b.setWellKnownRule("ipv4")
	return b
}
func (b *StringField) IPv6() *StringField {
	b.setWellKnownRule("ipv6")
	return b
}
func (b *StringField) URI() *StringField {
	b.setWellKnownRule("uri")
	return b
}
func (b *StringField) URIRef() *StringField {
	b.setWellKnownRule("uri_ref")
	return b
}
func (b *StringField) Address() *StringField {
	b.setWellKnownRule("address")
	return b
}
func (b *StringField) UUID() *StringField {
	b.setWellKnownRule("uuid")
	return b
}
func (b *StringField) TUUID() *StringField {
	b.setWellKnownRule("tuuid")
	return b
}
func (b *StringField) IPWithPrefixLen() *StringField {
	b.setWellKnownRule("ip_with_prefixlen")
	return b
}
func (b *StringField) IPv4WithPrefixLen() *StringField {
	b.setWellKnownRule("ipv4_with_prefixlen")
	return b
}
func (b *StringField) IPv6WithPrefixLen() *StringField {
	b.setWellKnownRule("ipv6_with_prefixlen")
	return b
}
func (b *StringField) IpRange() *StringField {
	b.setWellKnownRule("ip_prefix")
	return b
}
func (b *StringField) Ipv4Range() *StringField {
	b.setWellKnownRule("ipv4_prefix")
	return b
}
func (b *StringField) Ipv6Range() *StringField {
	b.setWellKnownRule("ipv6_prefix")
	return b
}
func (b *StringField) HostAndPort() *StringField {
	b.setWellKnownRule("host_and_port")
	return b
}

// Add known_regex methods

func (b *StringField) Strict(bo bool) *StringField {
	b.internal.rules["strict"] = bo
	return b
}
