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

Graphics._modifyExistingElements = function() {};
Graphics._disableTextSelection = function() {};
Graphics._disableContextMenu = function() {};
Graphics._createErrorPrinter = function() {};
Graphics._createVideo = function() {};
Graphics._createFPSMeter = function() {};
Graphics._createModeBox = function() {};
Graphics._createFontLoader = function() {};

/*Graphics._createRenderer = function() {
  var canvas = this._canvas;
  this._renderer = {
    render: function(stage) {
      stage._render(canvas._ebitenImage);
    },
  };
};*/

Graphics._setupEventHandlers = function() {
  // Do nothing
  // TODO: Set input handling
};

Graphics.isFontLoaded = function(name) {
  return true;
};

WebAudio._detectCodecs = function() {
  this._canPlayOgg = true;
  this._canPlayM4a = false;
};

WebAudio._createMasterGainNode = function() {};
WebAudio._setupEventHandlers = function() {};
WebAudio.prototype._load = function(url) {
  // TODO: Implement this
  this._buffer = [];
  this._totalTime = 0;
  this._loopStart = 0;
  this._loopLength = 0;
  this._onLoad();
};
WebAudio.prototype._startPlaying = function() {};

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
`

func (vm *VM) overrideCoreClasses() error {
	if _, err := vm.otto.Run(coreClassesSrc); err != nil {
		return err
	}
	return nil
}
