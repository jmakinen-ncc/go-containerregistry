// Copyright 2019 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/name"
)

func TestCatalogPage(t *testing.T) {
	cases := []struct {
		name         string
		responseBody []byte
		wantErr      bool
		wantRepos    []string
	}{{
		name:         "success",
		responseBody: []byte(`{"repositories":["test/test","foo/bar"]}`),
		wantErr:      false,
		wantRepos:    []string{"test/test", "foo/bar"},
	}, {
		name:         "not json",
		responseBody: []byte("notjson"),
		wantErr:      true,
	}}
	// TODO: add test cases for pagination

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalogPath := "/v2/_catalog"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v2/":
					w.WriteHeader(http.StatusOK)
				case catalogPath:
					if r.Method != http.MethodGet {
						t.Errorf("Method; got %v, want %v", r.Method, http.MethodGet)
					}

					w.Write(tc.responseBody)
				default:
					t.Fatalf("Unexpected path: %v", r.URL.Path)
				}
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("url.Parse(%v) = %v", server.URL, err)
			}

			reg, err := name.NewRegistry(u.Host)
			if err != nil {
				t.Fatalf("name.NewRegistry(%v) = %v", u.Host, err)
			}

			repos, err := CatalogPage(reg, "", 100)
			if (err != nil) != tc.wantErr {
				t.Errorf("Catalog() wrong error: %v, want %v: %v\n", (err != nil), tc.wantErr, err)
			}

			if diff := cmp.Diff(tc.wantRepos, repos); diff != "" {
				t.Errorf("Catalog() wrong repos (-want +got) = %s", diff)
			}
		})
	}
}
