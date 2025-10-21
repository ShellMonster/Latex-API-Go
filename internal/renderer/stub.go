package renderer

import (
	"errors"
	"fmt"
	"html"
	"strings"
)

// Renderer 定义渲染器应当具备的最小接口
type Renderer interface {
	Render(tex string) (string, error)
}

// Stub 用于在 Rust 模块尚未接入时提供占位 SVG
type Stub struct {
	defaultWidth  int
	defaultHeight int
}

// NewStub 构建占位渲染器
func NewStub() *Stub {
	return &Stub{
		defaultWidth:  400,
		defaultHeight: 80,
	}
}

// Render 将 LaTeX 文本包裹在提示信息中，方便前端联调
func (s *Stub) Render(tex string) (string, error) {
	trimmed := strings.TrimSpace(tex)
	if trimmed == "" {
		return "", errors.New("公式内容为空")
	}

	escaped := html.EscapeString(trimmed)
	width := s.defaultWidth + len([]rune(trimmed))*6
	height := s.defaultHeight
	viewBox := fmt.Sprintf("0 0 %d %d", width, height)

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="%s" preserveAspectRatio="xMinYMin meet"><rect width="100%%" height="100%%" fill="#f8fafc"/><text x="20" y="%d" font-size="20" font-family="monospace" fill="#0f172a">%s</text><text x="20" y="%d" font-size="14" font-family="monospace" fill="#64748b">Rust 渲染引擎开发中，当前为占位结果</text></svg>`,
		viewBox,
		height/2,
		escaped,
		height/2+24,
	), nil
}
