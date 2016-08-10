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

package js

import (
	"io/ioutil"
	"path/filepath"

	"github.com/robertkrimen/otto"
)

func jsLoadJSONFile(vm *VM, call otto.FunctionCall) (interface{}, error) {
	path, err := call.Argument(0).ToString()
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(filepath.Join(vm.pwd, path))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

const managerClassesSrc = `
SceneManager.run = function(sceneClass) {
  this.initialize();
  this.goto(sceneClass);
  this.requestUpdate();
};

SceneManager.update = function() {
  this.tickStart();
  this.updateInputData();
  this.updateMain();
  this.tickEnd();
};

SceneManager.setupErrorHandlers = function() {
};

SceneManager.shouldUseCanvasRenderer = function() {
  return true;
};

DataManager.loadDataFile = function(name, src) {
  window[name] = null;
  // TODO: Load file async
  var path = 'data/' + src;
  var content = _gophermv_loadJSONFile(path);
  window[name] = JSON.parse(content);
  DataManager.onLoad(window[name]);
};
`

func (vm *VM) overrideManagerClasses() error {
	if _, err := vm.otto.Run(managerClassesSrc); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_loadJSONFile", wrapFunc(jsLoadJSONFile, vm)); err != nil {
		return err
	}
	return nil
}

