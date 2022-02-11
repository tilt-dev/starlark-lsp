// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docstring

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	out := Parse(`An ACL entry: assigns given role (or roles) to given individuals or groups.

  Specifying an empty ACL entry is allowed. It is ignored everywhere. Useful for
  things like:

      luci.project(...)


  Args:
    roles :   a single role (as acl.role) or a list of roles to assign,
        blah-blah multiline.

    groups: a single group name or a list of groups to assign the role to.
    stuff: line1
      line2
      line3
    users: a single user email or a list of emails to assign the role to.


    empty:


  Returns:
    acl.entry struct, consider it opaque.
    Multiline.

  Note:
    blah-blah.

  Empty:
`)

	assert.Equal(t, strings.Join([]string{
		"An ACL entry: assigns given role (or roles) to given individuals or groups.",
		"",
		"Specifying an empty ACL entry is allowed. It is ignored everywhere. Useful for",
		"things like:",
		"",
		"    luci.project(...)",
	}, "\n"), out.Description)

	assert.Equal(t, []FieldsBlock{
		{
			Title: "Args",
			Fields: []Field{
				{"roles", "a single role (as acl.role) or a list of roles to assign, blah-blah multiline."},
				{"groups", "a single group name or a list of groups to assign the role to."},
				{"stuff", "line1 line2 line3"},
				{"users", "a single user email or a list of emails to assign the role to."},
				{"empty", ""},
			},
		},
	}, out.Fields)

	assert.Equal(t, []RemarkBlock{
		{"Returns", "acl.entry struct, consider it opaque.\nMultiline."},
		{"Note", "blah-blah."},
		{"Empty", ""},
	}, out.Remarks)
}

func TestNormalizedLines(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		assert.Len(t, normalizedLines("  \n\n\t\t\n  "), 0)
	})

	t.Run("One line and some space", func(t *testing.T) {
		assert.Equal(t, []string{"Blah"}, normalizedLines("  \n\n  Blah   \n\t\t\n  \n"))
	})

	t.Run("Deindents", func(t *testing.T) {
		actual := normalizedLines(`First paragraph,
		perhaps multiline.

		Second paragraph.

			Deeper indentation.

		Third paragraph.
		`)

		assert.Equal(t, []string{
			"First paragraph,",
			"perhaps multiline.",
			"",
			"Second paragraph.",
			"",
			"\tDeeper indentation.",
			"",
			"Third paragraph.",
		}, actual)
	})
}

func TestDeindent(t *testing.T) {
	t.Parallel()

	t.Run("Space only", func(t *testing.T) {
		assert.Equal(t, []string{"", "", ""}, deindent([]string{"  ", " \t\t  \t", ""}))
	})

	t.Run("Nothing to deindent", func(t *testing.T) {
		assert.Equal(t, []string{"", "a  ", "b", ""}, deindent([]string{"  ", "a  ", "b", "  "}))
	})

	t.Run("Deindentation works", func(t *testing.T) {
		assert.Equal(t, []string{"", "", "a", "b", "  c"}, deindent([]string{"   ", "", "  a", "  b", "    c"}))
	})

	t.Run("Works with tabs too", func(t *testing.T) {
		assert.Equal(t, []string{"", "", "a", "b", "\tc"}, deindent([]string{"\t\t", "", "\ta", "\tb", "\t\tc"}))
	})
}
