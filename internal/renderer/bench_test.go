package renderer

import (
	"testing"
)

func setupFFI(b *testing.B) Renderer {
	r, err := NewFFIRenderer()
	if err != nil {
		b.Skipf("无法加载 Rust 渲染器: %v", err)
	}
	return r
}

func BenchmarkFFISimpleSequential(b *testing.B) {
	r := setupFFI(b)
	formula := "E=mc^2"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Render(formula); err != nil {
			b.Fatalf("渲染失败: %v", err)
		}
	}
}

func BenchmarkFFISimpleParallel(b *testing.B) {
	r := setupFFI(b)
	formula := "E=mc^2"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := r.Render(formula); err != nil {
				b.Fatalf("渲染失败: %v", err)
			}
		}
	})
}

func BenchmarkFFIComplexSequential(b *testing.B) {
	r := setupFFI(b)
	formula := `\displaystyle \int_{0}^{\infty} \frac{\sin(x)}{x} e^{-x^2} \left( \sum_{k=1}^{5} \frac{x^k}{k!} \right) dx`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Render(formula); err != nil {
			b.Fatalf("渲染失败: %v", err)
		}
	}
}

func BenchmarkFFIComplexParallel(b *testing.B) {
	r := setupFFI(b)
	formula := `\displaystyle \int_{0}^{\infty} \frac{\sin(x)}{x} e^{-x^2} \left( \sum_{k=1}^{5} \frac{x^k}{k!} \right) dx`
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := r.Render(formula); err != nil {
				b.Fatalf("渲染失败: %v", err)
			}
		}
	})
}
