package lit

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/emscripten"
)

type Wasm struct {
	Context  context.Context
	Instance api.Module
}

func (w *Wasm) Call(name string, args ...uint64) (interface{}, error) {
	function := w.Instance.ExportedFunction(name)

	result, err := function.Call(context.Background(), args...)
	if err != nil {
		return nil, err
	}

	if len(result) > 1 {
		return result, nil
	} else if len(result) == 1 {
		return result[0], nil
	} else {
		return nil, nil
	}
}

func (w *Wasm) Close() {
	w.Instance.Close(w.Context)
}

func getStringFromMemory(m api.Memory, i, len uint32) (string, error) {
	b, ok := m.Read(context.Background(), i, len)
	if !ok {
		return "", fmt.Errorf("Failed to read memory at %d with length %d. Memory Size: %d", i, len, m.Size(context.Background()))
	}

	return string(b), nil
}

func wbingenThrow(mod api.Module, i, l uint32) {
	s, err := getStringFromMemory(mod.Memory(), i, l)
	if err != nil {
		panic(err)
	} else {
		panic(fmt.Errorf("__wbindgen_throw %s", s))
	}
}

type StringHeap struct {
	Stack []string
}

func (h *StringHeap) wbingenObjectDropRef(mod api.Module, i uint32) {
	// fmt.Printf("Object Drop Ref %d\n", i)

	if i >= uint32(len(h.Stack)) {
		panic(fmt.Errorf("Index %d is out of range for %d", i, len(h.Stack)))
	}

	h.Stack = append(h.Stack[:i], h.Stack[i+i:]...)
}

func (h *StringHeap) wbingenStringNew(mod api.Module, i, l uint32) uint32 {
	// fmt.Printf("String New: memorySize: %d, index: %d, len: %d\n", mod.Memory().Size(context.Background()), i, l)

	s, err := getStringFromMemory(mod.Memory(), i, l)
	if err != nil {
		panic(err)
	}

	index := len(h.Stack)
	h.Stack = append(h.Stack, s)

	return uint32(index)
}

func (h *StringHeap) wbingenLog9a99fb1af846153b(i uint32) {
	// fmt.Printf("%+v\n", h.Stack)
	// TODO: get object from heap by index
	if i >= uint32(len(h.Stack)) {
		panic(fmt.Errorf("Index %d is out of range for %d", i, len(h.Stack)))
	}

	fmt.Printf("WBG: %v\n", h.Stack[i])
}

func NewWasmInstance(ctx context.Context) (*Wasm, error) {
	wasmBytes, err := ioutil.ReadFile("./lit/threshold_crypto_wasm_bridge_bg.wasm")
	if err != nil {
		return nil, err
	}

	r := wazero.NewRuntime(ctx)
	h := &StringHeap{}

	if _, err := r.NewHostModuleBuilder("wbg").
		NewFunctionBuilder().WithFunc(wbingenThrow).Export("__wbindgen_throw").
		NewFunctionBuilder().WithFunc(h.wbingenObjectDropRef).Export("__wbindgen_object_drop_ref").
		NewFunctionBuilder().WithFunc(h.wbingenStringNew).Export("__wbindgen_string_new").
		NewFunctionBuilder().WithFunc(h.wbingenLog9a99fb1af846153b).Export("__wbg_log_9a99fb1af846153b").
		Instantiate(ctx, r); err != nil {
		return nil, err
	}

	if _, err := emscripten.Instantiate(ctx, r); err != nil {
		return nil, err
	}

	mod, err := r.InstantiateModuleFromBinary(ctx, wasmBytes)
	if err != nil {
		return nil, err
	}

	return &Wasm{ctx, mod}, nil
}
