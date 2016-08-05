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

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/gophermv/js"
	"golang.org/x/net/html"
)

const (
	rpgprojectFile = "Game.rpgproject"
	indexHTMLFile  = "index.html"
)

type Script struct {
	Src string
}

func (s *Script) Exec(path string, vm *js.VM) error {
	if err := vm.Exec(filepath.Join(path, s.Src)); err != nil {
		return err
	}
	return nil
}

func process(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	f, err := os.Open(filepath.Join(path, indexHTMLFile))
	if err != nil {
		return err
	}
	defer f.Close()
	doc, err := html.Parse(f)
	if err != nil {
		return err
	}
	scriptNodes := []*html.Node{}
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "script" {
			scriptNodes = append(scriptNodes, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	scripts := []*Script{}
	skips := map[string]struct{}{
		filepath.Join("js", "libs", "pixi.js"): struct{}{},
		filepath.Join("js", "libs", "fpsmeter.js"): struct{}{},
	}
	for _, n := range scriptNodes {
		for _, a := range n.Attr {
			if a.Key != "src" {
				continue
			}
			if _, ok := skips[filepath.Clean(a.Val)]; ok {
				continue
			}
			s := &Script{
				Src: a.Val,
			}
			scripts = append(scripts, s)
		}
	}
	vm, err := js.NewVM()
	if err != nil {
		return err
	}
	for _, s := range scripts {
		if err := s.Exec(path, vm); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		return
	}
	if err := process(arg); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
