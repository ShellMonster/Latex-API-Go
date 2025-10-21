package renderer

import (
	"strings"
	"testing"
)

func TestStub_Render(t *testing.T) {
	renderer := NewStub()
	svg, err := renderer.Render(`E=mc^2`)
	if err != nil {
		t.Fatalf("期望成功却返回错误: %v", err)
	}
	if svg == "" {
		t.Fatal("渲染结果不应为空")
	}
	if !strings.Contains(svg, "Rust 渲染引擎开发中") {
		t.Fatalf("占位提示缺失: %s", svg)
	}
	if !strings.Contains(svg, "E=mc^2") {
		t.Fatalf("公式内容未出现在 SVG 中: %s", svg)
	}
}

func TestStub_Render_Empty(t *testing.T) {
	renderer := NewStub()
	if _, err := renderer.Render("   "); err == nil {
		t.Fatal("空字符串应该返回错误")
	}
}
