// Package progress provides a simple progress bar for CLI.
package progress

import (
	"fmt"
	"os"
	"strings"
)

type Bar struct {
	total    int64
	current  int64
	width    int
	prefix   string
	complete rune
	incomplete rune
}

func NewBar(total int64, prefix string) *Bar {
	return &Bar{
		total:    total,
		width:    40,
		prefix:   prefix,
		complete: '█',
		incomplete: '░',
	}
}

func (b *Bar) Set(n int64) {
	b.current = n
	b.render()
}

func (b *Bar) Inc() {
	b.current++
	b.render()
}

func (b *Bar) render() {
	if b.total <= 0 {
		return
	}
	pct := float64(b.current) / float64(b.total)
	done := int(pct * float64(b.width))
	bar := strings.Repeat(string(b.complete), done) + strings.Repeat(string(b.incomplete), b.width-done)
	fmt.Fprintf(os.Stderr, "\r%s [%s] %3.0f%% (%d/%d)", b.prefix, bar, pct*100, b.current, b.total)
}

func (b *Bar) Done() {
	b.current = b.total
	b.render()
	fmt.Fprintln(os.Stderr)
}
