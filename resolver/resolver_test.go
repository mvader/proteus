package resolver

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/src-d/proteus/scanner"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var gopath = os.Getenv("GOPATH")

const project = "github.com/src-d/proteus"

func TestPackagesEnums(t *testing.T) {
	packages := Packages{
		&scanner.Package{
			Path: "foo",
			Enums: []*scanner.Enum{
				enum("Foo", "A", "B", "C"),
				enum("Bar", "D", "E"),
			},
		},
		&scanner.Package{
			Path: "bar",
			Enums: []*scanner.Enum{
				enum("Cmp", "Lt", "Eq", "Gt"),
			},
		},
	}

	enumSet := packages.Enums()
	require.Equal(t, 3, len(enumSet), "enums size")
	assertStrSet(t, enumSet, "bar.Cmp", "foo.Bar", "foo.Foo")
}

func TestGetPackagesInfo(t *testing.T) {
	packages := Packages{
		&scanner.Package{
			Path: "foo",
			Aliases: map[string]scanner.Type{
				"foo.Foo": scanner.NewBasic("int"),
				"foo.Bar": scanner.NewBasic("int"),
				"foo.Baz": scanner.NewBasic("int"),
			},
			Enums: []*scanner.Enum{
				enum("Foo", "A", "B", "C"),
				enum("Bar", "D", "E"),
			},
		},
		&scanner.Package{
			Path: "bar",
			Aliases: map[string]scanner.Type{
				"bar.Cmp": scanner.NewBasic("int"),
				"bar.Qux": scanner.NewBasic("int"),
			},
			Enums: []*scanner.Enum{
				enum("Cmp", "Lt", "Eq", "Gt"),
			},
		},
	}

	info := packages.Info()
	require.Equal(t, 2, len(info.Packages))
	assertStrSet(t, info.Packages, "bar", "foo")

	require.Equal(t, 2, len(info.Aliases))
	_, ok := info.Aliases["foo.Baz"]
	require.True(t, ok)

	_, ok = info.Aliases["bar.Qux"]
	require.True(t, ok)
}

func TestResolver(t *testing.T) {
	suite.Run(t, new(ResolverSuite))
}

type ResolverSuite struct {
	suite.Suite
	r *Resolver
}

func (s *ResolverSuite) SetupSuite() {
	s.r = New()
}

func (s *ResolverSuite) TestIsCustomType() {
	cases := []struct {
		path   string
		name   string
		result bool
	}{
		{"foo.bar/baz/bar", "Baz", false},
		{"net/url", "URL", false},
		{"time", "Time", true},
		{"time", "Duration", true},
	}

	for _, c := range cases {
		s.Equal(c.result, s.r.isCustomType(&scanner.Named{nil, c.path, c.name}), "%s.%s", c.path, c.name)
	}
}

func (s *ResolverSuite) TestResolve() {
	sc, err := scanner.New(projectPath("fixtures"), projectPath("fixtures/subpkg"))
	s.Nil(err)
	pkgs, err := sc.Scan()
	s.Nil(err)

	s.r.Resolve(Packages(pkgs))

	pkg := pkgs[0]
	s.assertStruct(pkg.Structs[0], "Bar", "Bar", "Baz")
	s.assertStruct(pkg.Structs[1], "Foo", "Bar", "Baz", "IntList", "IntArray", "Map", "Timestamp", "Duration", "Aliased")

	foo := pkg.Structs[1]
	aliasedType := foo.Fields[len(foo.Fields)-1].Type
	s.True(aliasedType.IsRepeated(), "Aliased type should be repeated")
	basic, ok := aliasedType.(*scanner.Basic)
	s.True(ok, "Aliased should be a basic type")
	s.Equal("int", basic.Name)
}

func (s *ResolverSuite) assertStruct(st *scanner.Struct, name string, fields ...string) {
	s.Equal(name, st.Name, "struct name")
	s.Equal(len(fields), len(st.Fields), "should have same struct fields")
	for _, f := range fields {
		s.True(st.HasField(f), "should have struct field %q", f)
	}
}

func assertStrSet(t *testing.T, set map[string]struct{}, expected ...string) {
	var vals []string
	for v := range set {
		vals = append(vals, v)
	}
	sort.Strings(vals)
	require.Equal(t, expected, vals)
}

func enum(name string, values ...string) *scanner.Enum {
	return &scanner.Enum{
		Name:   name,
		Values: values,
	}
}

func projectPath(pkg string) string {
	return filepath.Join(gopath, "src", project, pkg)
}
