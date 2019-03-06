package draw_test

import (
	"bytes"
	"testing"

	"github.com/travelaudience/rosie/internal/draw"
)

func TestDrawer_Flinef(t *testing.T) {
	b := bytes.NewBuffer(nil)
	d := &draw.Drawer{W: b}

	instructions := []struct {
		level int
		msg   string
	}{
		{
			level: 0,
			msg:   "level 0.1",
		},
		{
			level: 0,
			msg:   "level 0.2",
		},
		{
			level: 1,
			msg:   "level 1.1",
		},
		{
			level: 1,
			msg:   "level 1.2",
		},
		{
			level: 2,
			msg:   "level 2.1",
		},
		{
			level: 2,
			msg:   "level 2.1",
		},
		{
			level: 1,
			msg:   "level 1.3",
		},
		{
			level: 1,
			msg:   "level 1.4",
		},
		{
			level: 2,
			msg:   "level 2.3",
		},
		{
			level: 2,
			msg:   "level 2.4",
		},
		{
			level: 3,
			msg:   "level 3.1",
		},
		{
			level: 3,
			msg:   "level 3.2",
		},
	}

	for _, instruction := range instructions {
		d.NewEntry(instruction.level, "SECTION "+instruction.msg)

		d.NewSection()
		d.NewLine("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
		d.EndLine()
		d.NewLine("Etiam eu sapien feugiat, semper ligula id, facilisis ex.")
		d.EndLine()
		d.NewLine("Nunc lacinia ultricies lectus, non aliquam elit.")
		d.EndLine()
		d.NewLine("Sed lectus lorem, venenatis a sodales rhoncus, accumsan tincidunt est.")
		d.EndLine()
		d.NewLine("Aliquam rhoncus gravida magna sed dignissim. Vestibulum posuere ante at risus viverra luctus. Suspendisse vestibulum pretium mauris, in ultricies diam molestie vel. Duis egestas lectus id pretium mollis. Morbi eu dolor quis velit egestas consectetur vel vel lectus. Aliquam iaculis urna nec ligula dapibus porta. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Ut varius est sit amet risus lobortis maximus. Duis eget quam tortor.")
		d.EndLine()

	}
	d.EndEntry(0)

	if b.Len() == 0 {
		t.Error("nothing produced")
	}
}
