package viewport

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("default values on create by New", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)

		if !m.initialized {
			t.Errorf("on create by New Model should be initialized")
		}

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("default horizontalStep should be %d, got %d", defaultHorizontalStep, m.horizontalStep)
		}

		if m.MouseWheelDelta != 3 {
			t.Errorf("default MouseWheelDelta should be 3, got %d", m.MouseWheelDelta)
		}

		if !m.MouseWheelEnabled {
			t.Error("mouse wheel should be enabled by default")
		}
	})
}

func TestSetInitialValues(t *testing.T) {
	t.Parallel()

	t.Run("default horizontalStep", func(t *testing.T) {
		t.Parallel()

		m := Model{}
		m.setInitialValues()

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("default horizontalStep should be %d, got %d", defaultHorizontalStep, m.horizontalStep)
		}
	})
}

func TestSetHorizontalStep(t *testing.T) {
	t.Parallel()

	t.Run("change default", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("default horizontalStep should be %d, got %d", defaultHorizontalStep, m.horizontalStep)
		}

		newStep := 8
		m.SetHorizontalStep(newStep)
		if m.horizontalStep != newStep {
			t.Errorf("horizontalStep should be %d, got %d", newStep, m.horizontalStep)
		}
	})

	t.Run("no negative", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)

		if m.horizontalStep != defaultHorizontalStep {
			t.Errorf("default horizontalStep should be %d, got %d", defaultHorizontalStep, m.horizontalStep)
		}

		zero := 0
		m.SetHorizontalStep(-1)
		if m.horizontalStep != zero {
			t.Errorf("horizontalStep should be %d, got %d", zero, m.horizontalStep)
		}
	})
}

func TestMoveLeft(t *testing.T) {
	t.Parallel()

	zeroPosition := 0

	t.Run("zero position", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		if m.indent != zeroPosition {
			t.Errorf("default indent should be %d, got %d", zeroPosition, m.indent)
		}

		m.MoveLeft()
		if m.indent != zeroPosition {
			t.Errorf("indent should be %d, got %d", zeroPosition, m.indent)
		}
	})

	t.Run("move", func(t *testing.T) {
		t.Parallel()
		m := New(10, 10)
		if m.indent != zeroPosition {
			t.Errorf("default indent should be %d, got %d", zeroPosition, m.indent)
		}

		m.indent = defaultHorizontalStep * 2
		m.MoveLeft()
		newIndent := defaultHorizontalStep
		if m.indent != newIndent {
			t.Errorf("indent should be %d, got %d", newIndent, m.indent)
		}
	})
}

func TestMoveRight(t *testing.T) {
	t.Parallel()

	t.Run("move", func(t *testing.T) {
		t.Parallel()

		zeroPosition := 0

		m := New(10, 10)
		if m.indent != zeroPosition {
			t.Errorf("default indent should be %d, got %d", zeroPosition, m.indent)
		}

		m.MoveRight()
		newIndent := defaultHorizontalStep
		if m.indent != newIndent {
			t.Errorf("indent should be %d, got %d", newIndent, m.indent)
		}
	})
}

func TestResetIndent(t *testing.T) {
	t.Parallel()

	t.Run("reset", func(t *testing.T) {
		t.Parallel()

		zeroPosition := 0

		m := New(10, 10)
		m.indent = 500

		m.ResetIndent()
		if m.indent != zeroPosition {
			t.Errorf("indent should be %d, got %d", zeroPosition, m.indent)
		}
	})
}

func TestVisibleLines(t *testing.T) {
	t.Parallel()

	defaultList := []string{
		`57 Precepts of narcissistic comedy character Zote from an awesome "Hollow knight" game (https://store.steampowered.com/app/367520/Hollow_Knight/).`,
		`Precept One: 'Always Win Your Battles'. Losing a battle earns you nothing and teaches you nothing. Win your battles, or don't engage in them at all!`,
		`Precept Two: 'Never Let Them Laugh at You'. Fools laugh at everything, even at their superiors. But beware, laughter isn't harmless! Laughter spreads like a disease, and soon everyone is laughing at you. You need to strike at the source of this perverse merriment quickly to stop it from spreading.`,
		`Precept Three: 'Always Be Rested'. Fighting and adventuring take their toll on your body. When you rest, your body strengthens and repairs itself. The longer you rest, the stronger you become.`,
		`Precept Four: 'Forget Your Past'. The past is painful, and thinking about your past can only bring you misery. Think about something else instead, such as the future, or some food.`,
		`Precept Five: 'Strength Beats Strength'. Is your opponent strong? No matter! Simply overcome their strength with even more strength, and they'll soon be defeated.`,
		`Precept Six: 'Choose Your Own Fate'. Our elders teach that our fate is chosen for us before we are even born. I disagree.`,
		`Precept Seven: 'Mourn Not the Dead'. When we die, do things get better for us or worse? There's no way to tell, so we shouldn't bother mourning. Or celebrating for that matter.`,
		`Precept Eight: 'Travel Alone'. You can rely on nobody, and nobody will always be loyal. Therefore, nobody should be your constant companion.`,
		`Precept Nine: 'Keep Your Home Tidy'. Your home is where you keep your most prized possession - yourself. Therefore, you should make an effort to keep it nice and clean.`,
		`Precept Ten: 'Keep Your Weapon Sharp'. I make sure that my weapon, 'Life Ender', is kept well-sharpened at all times. This makes it much easier to cut things.`,
		`Precept Eleven: 'Mothers Will Always Betray You'. This Precept explains itself.`,
		`Precept Twelve: 'Keep Your Cloak Dry'. If your cloak gets wet, dry it as soon as you can. Wearing wet cloaks is unpleasant, and can lead to illness.`,
		`Precept Thirteen: 'Never Be Afraid'. Fear can only hold you back. Facing your fears can be a tremendous effort. Therefore, you should just not be afraid in the first place.`,
		`Precept Fourteen: 'Respect Your Superiors'. If someone is your superior in strength or intellect or both, you need to show them your respect. Don't ignore them or laugh at them.`,
		`Precept Fifteen: 'One Foe, One Blow'. You should only use a single blow to defeat an enemy. Any more is a waste. Also, by counting your blows as you fight, you'll know how many foes you've defeated.`,
		`...`,
	}

	t.Run("empty list", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		list := m.visibleLines()

		if len(list) != 0 {
			t.Errorf("list should be empty, got %d", len(list))
		}
	})

	t.Run("empty list: with indent", func(t *testing.T) {
		t.Parallel()

		m := New(10, 10)
		list := m.visibleLines()
		m.indent = 5

		if len(list) != 0 {
			t.Errorf("list should be empty, got %d", len(list))
		}
	})

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.lines = defaultList

		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("list should have %d lines, got %d", numberOfLines, len(list))
		}

		lastItem := numberOfLines - 1
		if list[lastItem] != defaultList[lastItem] {
			t.Errorf("%dth list item should the the same as %dth default list item", lastItem, lastItem)
		}
	})

	t.Run("list: with y offset", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.lines = defaultList
		m.YOffset = 5

		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("list should have %d lines, got %d", numberOfLines, len(list))
		}

		if list[0] == defaultList[0] {
			t.Error("first item of list should not be the first item of initial list because of Y offset")
		}

		lastItem := numberOfLines - 1
		if list[lastItem] != defaultList[m.YOffset+lastItem] {
			t.Errorf("%dth list item should the the same as %dth default list item", lastItem, lastItem)
		}
	})

	t.Run("list: with y offset: horizontal scroll", func(t *testing.T) {
		t.Parallel()
		numberOfLines := 10

		m := New(10, numberOfLines)
		m.lines = defaultList
		m.YOffset = 7

		// default list
		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("list should have %d lines, got %d", numberOfLines, len(list))
		}

		lastItem := numberOfLines - 1
		defaultLastItem := len(defaultList) - 1
		if list[lastItem] != defaultList[defaultLastItem] {
			t.Errorf("%dth list item should the the same as %dth default list item", lastItem, defaultLastItem)
		}

		perceptPrefix := "Precept"
		if !strings.HasPrefix(list[0], perceptPrefix) {
			t.Errorf("first list item has to have prefix %s", perceptPrefix)
		}

		// move right
		m.MoveRight()
		list = m.visibleLines()

		newPrefix := perceptPrefix[m.indent:]
		if !strings.HasPrefix(list[0], newPrefix) {
			t.Errorf("first list item has to have prefix %s, get %s", newPrefix, list[0])
		}

		if list[lastItem] != "" {
			t.Errorf("last item should be empty, got %s", list[lastItem])
		}

		// move left
		m.MoveLeft()
		list = m.visibleLines()
		if !strings.HasPrefix(list[0], perceptPrefix) {
			t.Errorf("first list item has to have prefix %s", perceptPrefix)
		}

		if list[lastItem] != defaultList[defaultLastItem] {
			t.Errorf("%dth list item should the the same as %dth default list item", lastItem, defaultLastItem)
		}
	})

	t.Run("list: with 2 cells symbols: horizontal scroll", func(t *testing.T) {
		t.Parallel()

		initList := []string{
			"あいうえお",
			"Aあいうえお",
			"あいうえお",
			"Aあいうえお",
		}
		numberOfLines := len(initList)

		m := New(10, numberOfLines)
		m.lines = initList

		// default list
		list := m.visibleLines()
		if len(list) != numberOfLines {
			t.Errorf("list should have %d lines, got %d", numberOfLines, len(list))
		}

		lastItem := numberOfLines - 1
		initLastItem := len(initList) - 1
		if list[lastItem] != initList[initLastItem] {
			t.Errorf("%dth list item should the the same as %dth default list item", lastItem, initLastItem)
		}

		// move right
		m.MoveRight()
		list = m.visibleLines()

		for i := range list {
			if i == 0 || i == 2 {
				cutLine := " えお"
				if list[i] != cutLine {
					t.Errorf("line must be `%s`, get `%s`", cutLine, list[i])
				}

				continue
			}
			cutLine := "うえお"
			if list[i] != cutLine {
				t.Errorf("line must be `%s`, get `%s`", cutLine, list[i])
			}
		}

		// move left
		m.MoveLeft()
		list = m.visibleLines()
		for i := range list {
			if list[i] != initList[i] {
				t.Errorf("line must be `%s`, get `%s`", list[i], initList[i])
			}
		}

		// move left second times do not change lites if indent == 0
		m.indent = 0
		m.MoveLeft()
		list = m.visibleLines()
		for i := range list {
			if list[i] != initList[i] {
				t.Errorf("line must be `%s`, get `%s`", list[i], initList[i])
			}
		}
	})
}
