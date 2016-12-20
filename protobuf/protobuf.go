package protobuf

import (
	"fmt"
	"path/filepath"
	"sort"
)

// Package represents an unique .proto file with its own package definition.
type Package struct {
	Name     string
	Path     string
	Imports  []string
	Options  Options
	Messages []*Message
	Enums    []*Enum
}

// Import tries to import the given protobuf type to the current package.
// If the type requires no import at all, nothing will be done.
func (p *Package) Import(typ *ProtoType) {
	if typ.Import != "" && !p.isImported(typ.Import) {
		p.Imports = append(p.Imports, typ.Import)
	}
}

// ImportFromPath adds a new import from a Go path.
func (p *Package) ImportFromPath(path string) {
	file := filepath.Join(path, "generated.proto")
	if path != p.Path && !p.isImported(file) {
		p.Imports = append(p.Imports, filepath.Join(path, "generated.proto"))
	}
}

func (p *Package) isImported(file string) bool {
	for _, i := range p.Imports {
		if i == file {
			return true
		}
	}
	return false
}

// Message is the representation of a Protobuf message.
type Message struct {
	Name     string
	Reserved []uint
	Options  Options
	Fields   []*Field
}

// Reserve reserves a position in the message.
func (m *Message) Reserve(pos uint) {
	if !m.isReserved(pos) {
		m.Reserved = append(m.Reserved, pos)
	}
}

func (m *Message) isReserved(pos uint) bool {
	for _, r := range m.Reserved {
		if r == pos {
			return true
		}
	}
	return false
}

// Field is the representation of a protobuf message field.
type Field struct {
	Name     string
	Pos      int
	Repeated bool
	Type     Type
	Options  Options
}

// Options are the set of options given to a field, message or enum value.
type Options map[string]OptionValue

// Option name and value pair.
type Option struct {
	Name  string
	Value OptionValue
}

// Sorted returns a sorted set of options.
func (o Options) Sorted() []*Option {
	var names = make([]string, 0, len(o))
	for k := range o {
		names = append(names, k)
	}

	sort.Stable(sort.StringSlice(names))
	var opts = make([]*Option, len(o))
	for i, n := range names {
		opts[i] = &Option{Name: n, Value: o[n]}
	}

	return opts
}

// OptionValue is the common interface for the value of an option, which can be
// a literal value (a number, true, etc) or a string value ("foo").
type OptionValue interface {
	fmt.Stringer
	isOptionValue()
}

// LiteralValue is a literal option value like true, false or a number.
type LiteralValue struct {
	val string
}

// NewLiteralValue creates a new literal option value.
func NewLiteralValue(val string) LiteralValue {
	return LiteralValue{val}
}

func (LiteralValue) isOptionValue() {}
func (v LiteralValue) String() string {
	return v.val
}

// StringValue is a string option value.
type StringValue struct {
	val string
}

// NewStringValue creates a new string option value.
func NewStringValue(val string) StringValue {
	return StringValue{val}
}

func (StringValue) isOptionValue() {}
func (v StringValue) String() string {
	return fmt.Sprintf("%q", v.val)
}

// Type is the common interface of all possible types, which are named types,
// maps and basic types.
type Type interface {
	fmt.Stringer
	isType()
}

// Named is a type which has a name and is defined somewhere else, maybe even
// in another package.
type Named struct {
	Package string
	Name    string
}

// NewNamed creates a new Named type given its package and name.
func NewNamed(pkg, name string) *Named {
	return &Named{pkg, name}
}

func (n Named) String() string {
	return fmt.Sprintf("%s.%s", n.Package, n.Name)
}

// Basic is one of the basic types of protobuf.
type Basic string

// NewBasic creates a new basic type given its name.
func NewBasic(name string) *Basic {
	b := Basic(name)
	return &b
}

func (b Basic) String() string {
	return string(b)
}

// Map is a key-value map type.
type Map struct {
	Key   Type
	Value Type
}

// NewMap creates a new Map type with the key and value types given.
func NewMap(k, v Type) *Map {
	return &Map{k, v}
}

func (m Map) String() string {
	return fmt.Sprintf("map<%s, %s>", m.Key, m.Value)
}

func (*Named) isType() {}
func (*Basic) isType() {}
func (*Map) isType()   {}

// Enum is the representation of a protobuf enumeration.
type Enum struct {
	Name    string
	Options Options
	Values  []*EnumValue
}

// EnumValue is a single value in an enumeration.
type EnumValue struct {
	Name    string
	Value   uint
	Options Options
}
