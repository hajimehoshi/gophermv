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
  }
  throw new Error('not supported element: ' + name);
};

Object.defineProperty(Document.prototype, "body", {
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
  throw new Error('not supported element: ' + JSON.stringify(child));
};

function HTMLScriptElement() {
}

function Image() {
}

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
		return err
	}
	return nil
}
