//go:build cgo

package renderer

/*
#cgo darwin LDFLAGS: -L${SRCDIR}/../../../Rust渲染 -lformula -Wl,-rpath,${SRCDIR}/../../../Rust渲染
#cgo linux  LDFLAGS: -L${SRCDIR}/../../../Rust渲染 -lformula -Wl,-rpath,${SRCDIR}/../../../Rust渲染
#include <stdlib.h>

char* render_svg(const char* formula);
void  free_svg(char* ptr);
*/
import "C"

import (
	"errors"
	"unsafe"
)

// ErrFFINilResult 表示 Rust FFI 返回了空指针
var ErrFFINilResult = errors.New("Rust 渲染返回空指针")

// ErrFFIMallocFailed 表示 C 字符串分配失败
var ErrFFIMallocFailed = errors.New("无法为公式分配 C 字符串")

type ffiRenderer struct{}

// NewFFIRenderer 创建基于 Rust 共享库的渲染器
func NewFFIRenderer() (Renderer, error) {
	return &ffiRenderer{}, nil
}

// Render 调用 Rust 的 render_svg 生成真实 SVG
func (r *ffiRenderer) Render(tex string) (string, error) {
	cstr := C.CString(tex)
	if cstr == nil {
		return "", ErrFFIMallocFailed
	}
	defer C.free(unsafe.Pointer(cstr))

	out := C.render_svg(cstr)
	if out == nil {
		return "", ErrFFINilResult
	}
	defer C.free_svg(out)

	return C.GoString(out), nil
}
