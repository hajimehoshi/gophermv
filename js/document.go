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

func jsSetOnLoadCallback(vm *VM, call otto.FunctionCall) (interface{}, error) {
	vm.onLoadCallback = call.Argument(0)
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
  set: function(func) {
    _gophermv_setOnLoadCallback(func);
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

function Canvas() {
}

Object.defineProperty(Canvas.prototype, 'width', {
  set: function(value) {
    this._width = value;
    if (!this._ebitenImage && 0 < this._width && 0 < this._height) {
      this._ebitenImage = _gophermv_newEbitenImage(this._width, this._height);
    }
  },
});

Object.defineProperty(Canvas.prototype, 'height', {
  set: function(value) {
    this._height = value;
    if (!this._ebitenImage && 0 < this._width && 0 < this._height) {
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
};

CanvasRenderingContext2D.prototype.clearRect = function(x, y, width, height) {
  if (!this._canvas._ebitenImage) {
    throw new Error('clearRect: canvas is not initialized');
  }
  _gophermv_ebitenImageClearRect(this._canvas._ebitenImage, x, y, width, height);
};

CanvasRenderingContext2D.prototype.save = function() {
  // TODO: Implement this
};

CanvasRenderingContext2D.prototype.restore = function() {
  // TODO: Implement this
};

CanvasRenderingContext2D.prototype.drawImage = function(image, x, y) {
  // TODO: Implement this
};

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
	if err := vm.otto.Set("_gophermv_setOnLoadCallback", wrapFunc(jsSetOnLoadCallback, vm)); err != nil {
		return err
	}
	if err := vm.otto.Set("_gophermv_requestAnimationFrame", wrapFunc(jsRequestAnimationFrame, vm)); err != nil {
		return err
	}
	if _, err := vm.otto.Run(documentSrc); err != nil {
		println(err.(*otto.Error).String())
		return err
	}
	return nil
}
