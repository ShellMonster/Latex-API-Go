//go:build !cgo

package renderer

import "errors"

// ErrCGODisabled 表示当前构建未启用 cgo
var ErrCGODisabled = errors.New("未启用 cgo，无法加载 Rust 渲染引擎")

// NewFFIRenderer 在未启用 cgo 时返回错误
func NewFFIRenderer() (Renderer, error) {
	return nil, ErrCGODisabled
}
