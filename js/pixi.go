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
  this._x = x || 0;
  this._y = y || 0;
};
Object.defineProperty(PIXI.Point.prototype, 'x', {
  get: function() { return this._x; },
  set: function(value) { this._x = value; },
});
Object.defineProperty(PIXI.Point.prototype, 'y', {
  get: function() { return this._y; },
  set: function(value) { this._y = value; },
});

PIXI.Rectangle = function(x, y, width, height) {
  this._x = x || 0;
  this._y = y || 0;
  this._width = width || 0;
  this._height = height || 0;
};
Object.defineProperty(PIXI.Rectangle.prototype, 'x', {
  get: function() { return this._x; },
  set: function(value) { this._x = value; },
});
Object.defineProperty(PIXI.Rectangle.prototype, 'y', {
  get: function() { return this._y; },
  set: function(value) { this._y = value; },
});
Object.defineProperty(PIXI.Rectangle.prototype, 'width', {
  get: function() { return this._width; },
  set: function(value) { this._width = value; },
});
Object.defineProperty(PIXI.Rectangle.prototype, 'height', {
  get: function() { return this._height; },
  set: function(value) { this._height = value; },
});

PIXI.DisplayObject = function() {};
PIXI.DisplayObjectContainer = function() {
  PIXI.DisplayObject.call(this);
  this._children = [];
  // TODO: Use this when rendering an image
  this._scale = {x: 1, y: 1};
  this.alpha = 1;
  this.visible = true;
};
PIXI.DisplayObjectContainer.prototype = Object.create(PIXI.DisplayObject.prototype);
PIXI.DisplayObjectContainer.prototype.constructor = PIXI.DisplayObjectContainer;
Object.defineProperty(PIXI.DisplayObjectContainer.prototype, 'scale', {
  get: function() { return this._scale; },
});
Object.defineProperty(PIXI.DisplayObjectContainer.prototype, 'children', {
  get: function() { return this._children; },
});
PIXI.DisplayObjectContainer.prototype.addChild = function(child) {
  return this.addChildAt(child, this._children.length);
};
PIXI.DisplayObjectContainer.prototype.addChildAt = function(obj, idx) {
  this._children.splice(idx, 0, obj);
  return obj;
};
PIXI.DisplayObjectContainer.prototype.removeChild = function(obj) {
  var idx = this._children.indexOf(obj);
  if (idx === -1) {
    return;
  }
  return this.removeChildAt(idx);
};
PIXI.DisplayObjectContainer.prototype.removeChildAt = function(idx) {
  var obj = this._children[idx];
  this._children.splice(idx, 1);
  return obj;
};
PIXI.DisplayObjectContainer.prototype._render = function(screen) {
  for (var i = 0; i < this.children.length; i++) {
    this.children[i]._render(screen);
  }
};

PIXI.Stage = function() {
  PIXI.DisplayObjectContainer.call(this);
};
PIXI.Stage.prototype = Object.create(PIXI.DisplayObjectContainer.prototype);
PIXI.Stage.prototype.constructor = PIXI.Stage;

PIXI.Sprite = function(texture) {
  PIXI.DisplayObjectContainer.call(this);
  this.anchor = new PIXI.Point();
  this.texture = texture;
};
PIXI.Sprite.prototype = Object.create(PIXI.DisplayObjectContainer.prototype);
PIXI.Sprite.prototype.constructor = PIXI.Sprite;
PIXI.Sprite.prototype._render = function(screen) {
  if (this.bitmap) {
    if (!this.bitmap.canvas._ebitenImage) {
      throw new Error('_render: this.bitmap.canvas._ebitenImage is not set');
    }
    // TODO: Use anchor!
    var frame = this.texture.frame;
    var op = {
      x:     frame.x,
      y:     frame.y,
      alpha: this.alpha,
    };
    _gophermv_ebitenImageDrawImage(screen, this.bitmap.canvas._ebitenImage, op);
  }
  for (var i = 0; i < this.children.length; i++) {
    this.children[i]._render(screen);
  }
};

PIXI.TilingSprite = function(texture) {
  PIXI.Sprite.call(this, texture);
};
PIXI.TilingSprite.prototype = Object.create(PIXI.Sprite.prototype);
PIXI.TilingSprite.prototype.constructor = PIXI.TilingSprite;

PIXI.BaseTexture = function() {};
PIXI.BaseTexture.prototype.dirty = function() {};

PIXI.Texture = function() {
  this._frame = new PIXI.Rectangle();
};
Object.defineProperty(PIXI.Texture.prototype, 'frame', {
  get: function() { return this._frame; },
});
PIXI.Texture.prototype.setFrame = function(frame) {
  this._frame = frame;
};

PIXI.AbstractFilter = function() {};

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
