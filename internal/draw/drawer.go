package draw

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	space = "\u2004"
	Width = 150
)

type Char string

type Drawer struct {
	W                           io.Writer
	depth, level                int
	used                        bool
	previousLevel               int
	sectionWritten, lineWritten int
	sections                    int
}

func (d *Drawer) NewEntry(l int, s string) {
	d.EndEntry(l)

	d.level = l
	d.sections = 0

	defer func() {
		d.used = true
		d.previousLevel = l
	}()
	d.depth = 0
	move := 0
	first := 0

	if l == 0 {
		d.previousLevel = 0
	}

	switch {
	case d.previousLevel == l && d.used:
		move = 1
		first = 0
	case d.previousLevel < l:
		move = 2
		first = 1
	case d.previousLevel > l:
		move = 2
		first = 1
	}

	if d.previousLevel > 0 {
		d.depth = l + 1 - move
		if d.depth < 0 {
			d.depth = 0
		}
	}
	d.depth += first

	if !d.used {
		_ = d.fprintf("%s%s\n", repeat("━", d.depth), "━┓")
	}

	d.NewSection()
	d.NewLine(s)
	d.EndLine()
}

func (d *Drawer) EndEntry(l int) {
	d.EndSection(false)

	diff := (d.level - l) + d.level - l - 1
	var sign string
	switch {
	case d.level == 0 && l == 0:
		return
	case l == d.level:
		sign = " ┃"
	case l > d.level:
		sign = " ┗━┓"
	case l < d.level && l == 0:
		d.lineWritten += d.fprintf("%s%s\n", repeat("━", diff), "━━┛")
		return
	case l < d.level:
		sign = " ┏" + repeat("━", diff) + "┛"
		d.depth = d.depth - (d.level - l)
	}
	d.lineWritten += d.fprintf("%s%s\n", d.prefix(), sign)
}

func (d *Drawer) NewSection() {
	d.sectionWritten = 0
	d.lineWritten = 0
	d.EndSection(true)
	if d.sections == 0 {
		d.newLine(" ┠─┬", "")
		d.NewColumn(0, d.straightLineUntilEnd(Width))
		d.EndLine()
	}

	d.sections += 1
}

func (d *Drawer) EndSection(next bool) {
	if d.sections == 0 {
		return
	}
	sign := " ┃ └"
	if next {
		sign = " ┃ ├"
	}
	d.newLine(sign, "")
	d.NewColumn(0, d.straightLineUntilEnd(Width))
	d.EndLine()
}

const newLineDefaultSign = " ┃ │ "

func (d *Drawer) NewLine(s string) {
	d.newLine(newLineDefaultSign, s)
}

func (d *Drawer) newLine(sign, s string) {
	d.lineWritten = 0

	p := detectPadding(s)
	s = withoutPadding(s)
	sc, ec, _ := retrieveColor(s)
	s = withoutColor(s)

	ps, pw := d.sprintf("%s%s%s", d.prefix(), white(sign), repeat(space, p))
	sw := utf8.RuneCountInString(s)

	if pw+sw > Width {
		for len(s) > 0 {
			idx := utf8.RuneCountInString(s)
			if idx > Width-pw {
				idx = Width - pw
			}

			d.lineWritten += d.fprint(ps, sc, s[:idx], ec)
			if len(s) > idx {
				d.EndLine()
			}
			s = s[idx:]
		}
	} else {
		d.lineWritten += d.fprint(ps, sc, s, ec)
	}
}

func (d *Drawer) NewColumn(m int, s string) {
	sc, ec, _ := retrieveColor(s)
	s = withoutColor(s)

	ps, pw := d.sprintf("%s", repeat(space, m-d.lineWritten))
	_, sw := d.sprintf("%s", s)

	if d.lineWritten+pw+sw > Width {
		idx := utf8.RuneCountInString(s)
		left := Width - pw - d.lineWritten
		if idx > left {
			idx = left
		}
		before := d.lineWritten + pw
		d.lineWritten += d.fprintf("%s%s%s%s\n", sc, ps, s[:idx], ec)

		d.NewLine("")
		d.NewColumn(before, sc+s[idx:]+ec)
	} else {
		d.lineWritten += d.fprintf("%s%s%s%s", sc, ps, s, ec)
	}
}

func (d *Drawer) EndLine() {
	d.lineWritten += d.fprintf("\n")
}

func (d *Drawer) sprintf(s string, args ...interface{}) (string, int) {
	str := fmt.Sprintf(s, args...)

	return str, utf8.RuneCountInString(str)
}

func (d *Drawer) fprint(s ...string) int {
	str := strings.Join(s, "")
	_, _ = fmt.Fprint(d.W, str)

	return utf8.RuneCountInString(str)
}

func (d *Drawer) fprintf(s string, args ...interface{}) int {
	str := fmt.Sprintf(s, args...)
	_, _ = fmt.Fprint(d.W, str)

	return utf8.RuneCountInString(str)
}

func (d *Drawer) straightLineUntilEnd(l int) string {
	return repeat("─", l-d.lineWritten)
}

func repeat(r string, l int) string {
	s := ""
	for i := 0; i < l; i++ {
		s += r
	}
	return s
}

func (d *Drawer) prefix() string {
	return repeat(space+" ", d.depth)
}

func white(s string) string {
	return fmt.Sprintf("\033[37m%s\033[0m", s)
}

var (
	colorExpression = regexp.MustCompile(`^(?P<opening>\033\[[0-9]{0,2}m).*(?P<closing>\033\[0m)$`)
)

func retrieveColor(s string) (string, string, bool) {
	if !colorExpression.MatchString(s) {
		return "", "", false
	}

	parts := colorExpression.FindStringSubmatch(s)
	return parts[1], parts[2], true
}

func withoutColor(s string) string {
	if sc, ec, ok := retrieveColor(s); ok {
		return strings.TrimPrefix(strings.TrimSuffix(s, ec), sc)
	}
	return s
}

func withoutPadding(s string) string {
	p := detectPadding(s)
	return s[p:]
}

func detectPadding(s string) (p int) {
	for _, r := range s {
		if r != ' ' {
			return
		}
		p += 1
	}
	return
}
