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
	"github.com/robertkrimen/otto"
)

func jsAppendOnLoadCallback(vm *VM, call otto.FunctionCall) (interface{}, error) {
	vm.onLoadCallbacks = append(vm.onLoadCallbacks, call.Argument(0));
	return otto.Value{}, nil
}

func jsAppendScript(vm *VM, call otto.FunctionCall) (interface{}, error) {
	src, err := call.Argument(0).ToString()
	if err != nil {
		return otto.Value{}, err
	}
	vm.Enqueue(src)
	return otto.Value{}, nil
}

func jsRequestAnimationFrame(vm *VM, call otto.FunctionCall) (interface{}, error) {
	vm.requestAnimationFrameCallbacks = append(vm.requestAnimationFrameCallbacks, call.Argument(0))
	return otto.Value{}, nil
}

const documentSrc = `
function Window() {
}

Object.defineProperty(Window.prototype, 'document', {
  get: function() {
    return this._document;
  },
});

Object.defineProperty(Window.prototype, 'onload', {
  set: function(f) {
    _gophermv_appendOnLoadCallback(f);
  },
});

Object.defineProperty(Window.prototype, 'location', {
  get: function() {
    // TODO: Use the flags for hash or search
    return {
      hash:     '',
      host:     '',
      hostname: '',
      href:     '',
      pathname: '',
      port:     0,
      protocol: '',
      search:   '',
    };
  },
});

Object.defineProperty(Window.prototype, 'navigator', {
  get: function() {
    return {
      userAgent: 'gophermv',
    };
  },
});

Object.defineProperty(Window.prototype, 'localStorage', {
  get: function() {
    return this._localStorage;
  },
});

Window.prototype.requestAnimationFrame = function(func) {
  _gophermv_requestAnimationFrame(func);
};

function Document() {
  this.initialize.apply(this, arguments);
}

Document.prototype.initialize = function() {
  this._body = new HTMLBodyElement();
};

Document.prototype.createElement = function(name) {
  switch (name) {
  case 'script':
    return new HTMLScriptElement();
  case 'canvas':
    return new Canvas();
  }
  throw new Error('createElement: not supported element: ' + name);
};

Object.defineProperty(Document.prototype, 'body', {
  get: function() {
    return this._body;
  },
});

function HTMLBodyElement() {
}

HTMLBodyElement.prototype.appendChild = function(child) {
  if (child instanceof HTMLScriptElement) {
    _gophermv_appendScript(child.src);
    return;
  }
  if (child instanceof Canvas) {
    // TODO: Draw canvas elements
    if (!this._canvasElements) {
      this._canvasElements = [];
    }
    this._canvasElements.push(child);
    return;
  }
  throw new Error('appendChild: not supported element: ' + JSON.stringify(child));
};

HTMLBodyElement.prototype._canvasEbitenImages = function() {
  return this._canvasElements.sort(function(a, b) {
    var az = a.style.zIndex || 0;
    var bz = b.style.zIndex || 0;
    return -(az - bz);
  }).map(function(e) {
    return e._ebitenImage;
  });
};

function HTMLScriptElement() {
}

function Image() {
  this.initialize.apply(this, arguments);
}

Image.prototype.initialize = function() {
  this._ebitenImage = null;
};

Object.defineProperty(Image.prototype, 'src', {
  set: function(value) {
    this._ebitenImage = _gophermv_loadEbitenImage(value);
  },
});

Object.defineProperty(Image.prototype, 'width', {
  get: function() {
    var size = _gophermv_ebitenImageSize(this._ebitenImage);
    return size[0];
  },
});

Object.defineProperty(Image.prototype, 'height', {
  get: function() {
    var size = _gophermv_ebitenImageSize(this._ebitenImage);
    return size[1];
  },
});

Object.defineProperty(Image.prototype, 'onload', {
  set: function(f) {
    _gophermv_appendOnLoadCallback(f);
  },
});

function Canvas() {
}

Object.defineProperty(Canvas.prototype, 'width', {
  get: function() {
    var size = _gophermv_ebitenImageSize(this._ebitenImage);
    return size[0];
  },
  set: function(value) {
    if (this._width === value) {
      return;
    }
    this._width = value;
    // TODO: Note that this recreates _ebitenImage. Should we preserve the exisitng image?
    if (0 < this._width && 0 < this._height) {
      this._ebitenImage = _gophermv_newEbitenImage(this._width, this._height);
    }
  },
});

Object.defineProperty(Canvas.prototype, 'height', {
  get: function() {
    var size = _gophermv_ebitenImageSize(this._ebitenImage);
    return size[1];
  },
  set: function(value) {
    if (this._height === value) {
      return;
    }
    this._height = value;
    if (0 < this._width && 0 < this._height) {
      this._ebitenImage = _gophermv_newEbitenImage(this._width, this._height);
    }
  },
});

Object.defineProperty(Canvas.prototype, 'style', {
  get: function() {
    if (!this._style) {
      this._style = {};
    }
    return this._style;
  },
});

Canvas.prototype.getContext = function(mode) {
  if (mode === '2d') {
    return new CanvasRenderingContext2D(this);
  }
  throw new Error('getContext: not supported mode: ' + mode);
};

function CanvasRenderingContext2D() {
  this.initialize.apply(this, arguments);
}

CanvasRenderingContext2D.prototype.initialize = function(canvas) {
  this._canvas = canvas;
  this._stateStack = [{}];
};

(function() {
  function desc(name, defaultValue) {
    return {
      get: function() {
        var state = this._stateStack[this._stateStack.length - 1];
        var val = state[name];
        if (val === undefined) {
          return defaultValue;
        }
        return val;
      },
      set: function(value) {
        var state = this._stateStack[this._stateStack.length - 1];
        state[name] = value;
      },
    };
  }

  Object.defineProperties(CanvasRenderingContext2D.prototype, {
    strokeStyle:              desc('strokeStyle', '#000000'),
    fillStyle:                desc('fillStyle', '#000000'),
    globalAlpha:              desc('globalAlpha', 1.0),
    lineWidth:                desc('lineWidth', 1.0),
    lineCap:                  desc('lineCap', 'butt'),
    lineJoin:                 desc('lineJoin', 'miter'),
    miterLimit:               desc('miterLimit', 10.0),
    shadowOffsetX:            desc('shadowOffsetX', 0),
    shadowOffsetY:            desc('shadowOffsetY', 0),
    shadowBlur:               desc('shadowBlur', 0),
    shadowColor:              desc('shadowColor', '#000000'),
    globalCompositeOperation: desc('globalCompositeOperation', 'source-over'),
    font:                     desc('font', 'normal-weight 10px sans-serif'),
    textAlign:                desc('textAlign', 'start'),
    textBaseline:             desc('textBaseline', 'alphabetic'),
  });
})();

CanvasRenderingContext2D.prototype.clearRect = function(x, y, width, height) {
  if (!this._canvas._ebitenImage) {
    throw new Error('clearRect: canvas is not initialized');
  }
  _gophermv_ebitenImageClearRect(this._canvas._ebitenImage, x, y, width, height);
};

CanvasRenderingContext2D.prototype.setTransform = function(a, b, c, d, tx, ty) {
  var state = this._stateStack[this._stateStack.length - 1];
  state['transform'] = [a, b, c, d, tx, ty];
};

CanvasRenderingContext2D.prototype.strokeText = function(text, tx, ty, maxWidth) {
  // Used at Bitmap.prototype._drawTextOutline
  // TODO: Implement this?
};

CanvasRenderingContext2D.prototype.fillText = function(text, tx, ty, maxWidth) {
  // Used at Bitmap.prototype._drawTextBody
  // TODO: Implement this?
};

CanvasRenderingContext2D.prototype.beginPath = function() {
  // Used at WindowLayer.prototype._renderCanvas
  // TODO: Implement this?
};

CanvasRenderingContext2D.prototype.closePath = function() {
  // Used at WindowLayer.prototype._renderCanvas
  // TODO: Implement this?
};

CanvasRenderingContext2D.prototype.clip = function() {
  // Used at WindowLayer.prototype._renderCanvas
  // TODO: Implement this?
};

CanvasRenderingContext2D.prototype.rect = function(x, y, width, height) {
  // Used at WindowLayer.prototype._renderCanvas
  // TODO: Implement this?
};

(function() {
  function clone(style) {
    var result = {};
    for (var attr in style) {
      if (!style.hasOwnProperty(attr)) {
        continue;
      }
      result[attr] = style[attr];
    }
    return result;
  }

  CanvasRenderingContext2D.prototype.save = function() {
    var newState = clone(this._stateStack[this._stateStack.length - 1]);
    this._stateStack.push(newState);
  };
})();

CanvasRenderingContext2D.prototype.restore = function() {
  this._stateStack.pop();
};

CanvasRenderingContext2D.prototype.drawImage = function(image) {
  if (!this._canvas._ebitenImage) {
    throw new Error('drawImage: canvas is not initialized');
  }
  var sx = 0;
  var sy = 0;
  var sw = image.width;
  var sh = image.height
  var dx = 0;
  var dy = 0;
  var dw = sw;
  var dh = sh;
  switch (arguments.length) {
  case 3:
    dx = arguments[1];
    dy = arguments[2];
    break;
  case 5:
    dx = arguments[1];
    dy = arguments[2];
    dw = arguments[3];
    dh = arguments[4];
    sw = dw
    sh = dh
    break;
  case 9:
    sx = arguments[1];
    sy = arguments[2];
    sw = arguments[3];
    sh = arguments[4];
    dx = arguments[5];
    dy = arguments[6];
    dw = arguments[7];
    dh = arguments[8];
    break;
  default:
    throw new Error('drawImage: invalid argument num: ' + arguments.length);
  }
  var imageParts = [
    {
      src: [sx, sy, sx+sw, sy+sh],
      dst: [dx, dy, dx+dw, dy+dh],
    }
  ];
  var dst = this._canvas._ebitenImage;
  // TODO: What if |image| is a Canvas?
  var src = image._ebitenImage;
  // TODO: Use composition mode
  var state = this._stateStack[this._stateStack.length - 1];
  var op = {
    geom:       (state['transform'] || [1, 0, 0, 1, 0, 0]),
    imageParts: imageParts,
    alpha:      this.globalAlpha,
  };
  _gophermv_ebitenImageDrawImage(dst, src, op);
};

CanvasRenderingContext2D.prototype.fillRect = function(x, y, width, height) {
  if (!this._canvas._ebitenImage) {
    throw new Error('clearRect: canvas is not initialized');
  }
  var dst = this._canvas._ebitenImage;
  var color = 0;
  var m = null;
  var alpha = 0xff;
  if (m = this.fillStyle.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/)) {
    color = (m[1] << 24) | (m[2] << 16) | (m[3] << 8);
  } else if (m = this.fillStyle.match(/^rgba\((\d+),\s*(\d+),\s*(\d+)\s*([\d.]+)\)$/)) {
    alpha = (parseFloat(m[4]) * 255) | 0;
    color = (m[1] << 24) | (m[2] << 16) | (m[3] << 8);
  } else if (m = this.fillStyle.match(/^#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])$/)) {
    var r = parseInt(m[1], 16) * 0x11;
    var g = parseInt(m[2], 16) * 0x11;
    var b = parseInt(m[3], 16) * 0x11;
    color = (r << 24) | (g << 16) | (b << 8);
  } else if (m = this.fillStyle.match(/^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$/)) {
    var r = parseInt(m[1], 16);
    var g = parseInt(m[2], 16);
    var b = parseInt(m[3], 16);
    color = (r << 24) | (g << 16) | (b << 8);
  } else {
    throw new Error('invalid style format: ' + this.fillStyle);
  }
  alpha = (alpha * this.globalAlpha)|0;
  color |= alpha;
  _gophermv_ebitenImageFillRect(dst, x, y, width, height, color);
};

CanvasRenderingContext2D.prototype.getImageData = function(x, y, width, height) {
  var data = _gophermv_ebitenImagePixels(this._canvas._ebitenImage, x, y, width, height);
  return new ImageData(data, width, height);
};

function ImageData(data, width, height) {
  this._data = data;
  this._width = width;
  this._height = height;
}

Object.defineProperty(ImageData.prototype, 'width', {
  get: function() { return this._width; },
});

Object.defineProperty(ImageData.prototype, 'height', {
  get: function() { return this._height; },
});

Object.defineProperty(ImageData.prototype, 'data', {
  get: function() { return this._data; },
});

function AudioContext() {
}

function LocalStorage() {
  this._store = {};
}

LocalStorage.prototype.getItem = function(key) {
  return this._store[key];
};

LocalStorage.prototype.setItem = function(key, value) {
  this._store[key] = value;
};

LocalStorage.prototype.removeItem = function(key) {
  delete this._store[key];
};

(function(global) {
  var names = Object.getOwnPropertyNames(Window.prototype);
  for (var i in names) {
    var name = names[i];
    if (name === 'constructor') {
      continue;
    }
    var desc = Object.getOwnPropertyDescriptor(Window.prototype, name);
    Object.defineProperty(global, name, desc);
  }
  global.window = global;
  global._document = new Document();
  global._localStorage = new LocalStorage();
})(this);
`

func (vm *VM) initDocument() error {
	if err := vm.otto.Set("_gophermv_appendScript", wrapFunc(jsAppendScript, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_appendOnLoadCallback", wrapFunc(jsAppendOnLoadCallback, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_requestAnimationFrame", wrapFunc(jsRequestAnimationFrame, vm)); err != nil {
		return err
	}
	if _, err := vm.otto.Run(documentSrc); err != nil {
		return err
	}
	return nil
}
