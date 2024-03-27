package key

import (
	"math/bits"
	"testing"
)

type optConfig uint

const (
	optWithKeys = optConfig(1) << iota
	optWithHelp
	optWithDisabled
)
const optNone = optConfig(0)

type bindingTest struct {
	desc string
	bin  Binding
	ok   bool
}

func (o *optConfig) next() optConfig {
	c := optConfig(1) << bits.TrailingZeros(uint(*o))
	*o &^= c // clear optConfig at LSB
	return c
}

func makeBinding(o ...optConfig) Binding {
	opt := []BindingOpt{}
	for _, c := range o {
		for c > 0 {
			switch c.next() {
			case optWithKeys:
				opt = append(opt, WithKeys("k", "up"))
			case optWithHelp:
				opt = append(opt, WithHelp("↑/k", "move up"))
			case optWithDisabled:
				opt = append(opt, WithDisabled())
			}
		}
	}
	return NewBinding(opt...)
}

func TestBinding_Enabled(t *testing.T) {
	for _, tc := range []bindingTest{
		{
			desc: "WithKeys",
			bin: makeBinding(optWithKeys),
			ok: true,
		},
		{
			desc: "WithKeys, WithHelp",
			bin: makeBinding(optWithKeys | optWithHelp),
			ok: true,
		},
		{
			desc: "WithHelp",
			bin: makeBinding(optWithHelp),
			ok: true,
		},
		{
			desc: "",
			bin: makeBinding(),
			ok: true,
		},
		{
			desc: "WithDisabled, WithKeys",
			bin: makeBinding(optWithDisabled | optWithKeys),
			ok: false,
		},
		{
			desc: "WithDisabled, WithKeys, WithHelp",
			bin: makeBinding(optWithDisabled | optWithKeys | optWithHelp),
			ok: false,
		},
		{
			desc: "WithDisabled, WithHelp",
			bin: makeBinding(optWithDisabled | optWithHelp),
			ok: false,
		},
		{
			desc: "WithDisabled",
			bin: makeBinding(optWithDisabled),
			ok: false,
		},
	} {
		if tc.bin.Enabled() != tc.ok {
			t.Errorf("bin.Enabled(%s) != %t", tc.desc, tc.ok)
		}
		tc.bin.SetEnabled(!tc.ok)
		if tc.bin.Enabled() == tc.ok {
			t.Errorf("bin.Enabled(%s) != %t", tc.desc, !tc.ok)
		}
		tc.bin.SetEnabled(true)
		tc.bin.Unbind()
		if tc.bin.Enabled() {
			t.Errorf("bin.Enabled(%s) != %t", tc.desc, false)
		}
	}
}

func TestBinding_HelpAvailable(t *testing.T) {
	for _, tc := range []bindingTest{
		{
			desc: "WithKeys",
			bin: makeBinding(optWithKeys),
			ok: false,
		},
		{
			desc: "WithKeys, WithHelp",
			bin: makeBinding(optWithKeys | optWithHelp),
			ok: true,
		},
		{
			desc: "WithHelp",
			bin: makeBinding(optWithHelp),
			ok: true,
		},
		{
			desc: "",
			bin: makeBinding(),
			ok: false,
		},
		{
			desc: "WithDisabled, WithKeys",
			bin: makeBinding(optWithDisabled | optWithKeys),
			ok: false,
		},
		{
			desc: "WithDisabled, WithKeys, WithHelp",
			bin: makeBinding(optWithDisabled | optWithKeys | optWithHelp),
			ok: true,
		},
		{
			desc: "WithDisabled, WithHelp",
			bin: makeBinding(optWithDisabled | optWithHelp),
			ok: true,
		},
		{
			desc: "WithDisabled",
			bin: makeBinding(optWithDisabled),
			ok: false,
		},
	} {
		if tc.bin.HelpAvailable() != tc.ok {
			t.Errorf("bin.HelpAvailable(%s) != %t", tc.desc, tc.ok)
		}
		tc.bin.SetEnabled(false)
		if tc.bin.HelpAvailable() != tc.ok {
			t.Errorf("bin.HelpAvailable(%s) != %t", tc.desc, tc.ok)
		}
		tc.bin.SetEnabled(true)
		tc.bin.Unbind()
		if tc.bin.HelpAvailable() {
			t.Errorf("bin.HelpAvailable(%s) != %t", tc.desc, false)
		}
	}
}
