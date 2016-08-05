// Copyright 2016 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package js_test

import (
	"testing"

	. "github.com/hajimehoshi/gophermv/js"
)

func TestStringFormat(t *testing.T) {
	tests := []struct {
		this string
		args []string
		out  string
		err  error
	}{
		{
			this: "",
			args: []string{""},
			out:  "",
			err:  nil,
		},
		{
			this: "foo%1bar%2baz",
			args: []string{"hello", "world"},
			out:  "foohellobarworldbaz",
			err:  nil,
		},
		{
			this: "%2%1%2",
			args: []string{"hello", "world"},
			out:  "worldhelloworld",
			err:  nil,
		},
	}
	for _, test := range tests {
		out, err := StringFormat(test.this, test.args...)
		if test.out != out || test.err != err {
			t.Errorf("StringFormat(%v, %v) = %v, %v want %v, %v", test.this, test.args, out, err, test.out, test.err)
		}
	}
}
