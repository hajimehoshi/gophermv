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

const coreClassesSrc = `
Graphics._testCanvasBlendModes = function() {
  this._canUseDifferenceBlend = false;
  this._canUseSaturationBlend = false;
};

Graphics._modifyExistingElements = function() {
  // Do nothing
};

Graphics._createAllElements = function() {
  // Do nothing
};

Graphics._disableTextSelection = function() {
  // Do nothing
};

Graphics._disableContextMenu = function() {
  // Do nothing
};

Graphics._setupEventHandlers = function() {
  // Do nothing
  // TODO: Set input handling
};

WebAudio._detectCodecs = function() {
  this._canPlayOgg = true;
  this._canPlayM4a = false;
};

WebAudio._createMasterGainNode = function() {
  // Do nothing
};

WebAudio._setupEventHandlers = function() {
  // Do nothing
};

Input._setupEventHandlers = function() {
  // Do nothing
  // TODO: Set input handling
};

TouchInput._setupEventHandlers = function() {
  // Do nothing
  // TODO: Set input handling
};

Utils.canReadGameFiles = function() {
  return true;
};

Bitmap.prototype.initialize = function(width, height) {
  this._width = width;
  this._height = height;
};
`

func (vm *VM) overrideCoreClasses() error {
	if _, err := vm.otto.Run(coreClassesSrc); err != nil {
		return err
	}
	return nil
}
