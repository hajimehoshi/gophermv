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

const pixiSrc = `
var PIXI = {};

PIXI.Point = function(x, y) {
  this._x = x;
  this._y = y;
};

Object.defineProperty(PIXI.Point.prototype, 'x', {
  get: function() { return this._x; },
});

Object.defineProperty(PIXI.Point.prototype, 'y', {
  get: function() { return this._y; },
});

PIXI.Rectangle = function(x, y, width, height) {
  this._x = x;
  this._y = y;
  this._width = width;
  this._height = height;
};

Object.defineProperty(PIXI.Rectangle.prototype, 'x', {
  get: function() { return this._x; },
});

Object.defineProperty(PIXI.Rectangle.prototype, 'y', {
  get: function() { return this._y; },
});

Object.defineProperty(PIXI.Rectangle.prototype, 'width', {
  get: function() { return this._width; },
});

Object.defineProperty(PIXI.Rectangle.prototype, 'height', {
  get: function() { return this._height; },
});

PIXI.Sprite = function() {};
PIXI.DisplayObjectContainer = function() {};
PIXI.TilingSprite = function() {};
PIXI.AbstractFilter = function() {};
PIXI.DisplayObject = function() {};
PIXI.Stage = function() {};

// Called at Bitmap
PIXI.BaseTexture = function() {};
PIXI.BaseTexture.prototype.dirty = function() {
};
PIXI.scaleModes = {
  NEAREST: 0,
  LINEAR: 1,
};
`

func (vm *VM) initPixi() error {
	if _, err := vm.otto.Run(pixiSrc); err != nil {
		return err
	}
	return nil
}
