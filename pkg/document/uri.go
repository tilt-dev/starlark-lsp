/*
Copyright 2021 The Skaffold Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Forked from https://github.com/GoogleContainerTools/skaffold/blob/1ac7448e79bedc3e36f5f2021223f85bc57d8193/pkg/skaffold/lsp/util.go
	* milas(2022-01-26): return errors instead of logging on invalid URI
*/

package document

import (
	"fmt"
	"net/url"
	"strings"

	"go.lsp.dev/uri"
)

const gitPrefix = "git:/"

func uriToFilename(v uri.URI) (string, error) {
	s := string(v)
	fixed, ok := fixURI(s)

	if !ok {
		unescaped, err := url.PathUnescape(s)
		if err == nil {
			s = unescaped
		}
		return "", fmt.Errorf("uri is not a filepath: %q", s)
	}
	v = uri.URI(fixed)
	return v.Filename(), nil
}

// workaround for unsupported file paths (git + invalid file://-prefix )
func fixURI(s string) (string, bool) {
	if strings.HasPrefix(s, gitPrefix) {
		return "file:///" + s[len(gitPrefix):], true
	}
	if !strings.HasPrefix(s, "file:///") {
		// VS Code sends URLs with only two slashes, which are invalid. golang/go#39789.
		if strings.HasPrefix(s, "file://") {
			return "file:///" + s[len("file://"):], true
		}
		return "", false
	}
	return s, true
}
