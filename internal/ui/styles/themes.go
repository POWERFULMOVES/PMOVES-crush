package styles

import (
	"image/color"

	"github.com/charmbracelet/x/exp/charmtone"
)

// PMOVES agent signature colors (from pmoves/config/agent_signatures.yaml).
// These mirror the --pm-* design tokens in pmoves/design/.
var (
	pmovesSkyBlue   = color.RGBA{R: 0x0E, G: 0xA5, B: 0xE9, A: 0xFF} // ◇ crush
	pmovesViolet    = color.RGBA{R: 0x7C, G: 0x3A, B: 0xED, A: 0xFF} // ◆ claude-opus
	pmovesCrimson   = color.RGBA{R: 0xE1, G: 0x1D, B: 0x48, A: 0xFF} // ✦ darkxside
	pmovesEmerald   = color.RGBA{R: 0x05, G: 0x96, B: 0x69, A: 0xFF} // ▲ kilocode
	pmovesTeal      = color.RGBA{R: 0x0D, G: 0x94, B: 0x88, A: 0xFF} // ◉ 4090-claude
	pmovesVoid      = color.RGBA{R: 0x05, G: 0x05, B: 0x08, A: 0xFF} // --pm-bg
	pmovesSurface   = color.RGBA{R: 0x12, G: 0x12, B: 0x1A, A: 0xFF} // --pm-surface
	pmovesInk       = color.RGBA{R: 0xF8, G: 0xF8, B: 0xF8, A: 0xFF} // --pm-ink
	pmovesInkSubtle = color.RGBA{R: 0xA0, G: 0xA0, B: 0xA8, A: 0xFF} // --pm-ink-subtle
)

// ThemeKeyForProvider returns a stable identifier for the theme
// associated with the given provider ID. Providers that share a theme
// yield the same key, so callers can cheaply detect when switching
// providers would not actually change the active theme and skip the
// expensive style rebuild. This is the single source of truth for the
// provider-to-theme mapping; [ThemeForProvider] builds on it.
func ThemeKeyForProvider(providerID string) string {
	switch providerID {
	case "hyper":
		return "hyper"
	case "zai", "tensorzero", "ollama":
		return "pmoves"
	default:
		return "default"
	}
}

// ThemeForProvider returns the Styles associated with the given provider
// ID. Unknown or empty provider IDs yield the default Charmtone Pantera
// theme.
func ThemeForProvider(providerID string) Styles {
	switch ThemeKeyForProvider(providerID) {
	case "hyper":
		return HypercrushObsidiana()
	case "pmoves":
		return PMOVESArmor()
	default:
		return CharmtonePantera()
	}
}

// CharmtonePantera returns the Charmtone dark theme. It's the default style
// for the UI.
func CharmtonePantera() Styles {
	s := quickStyle(quickStyleOpts{
		primary:   charmtone.Charple,
		secondary: charmtone.Dolly,
		accent:    charmtone.Bok,
		keyword:   charmtone.Blush,

		fgBase:       charmtone.Sash,
		fgMoreSubtle: charmtone.Squid,
		fgSubtle:     charmtone.Smoke,
		fgMostSubtle: charmtone.Oyster,

		onPrimary: charmtone.Butter,

		bgBase:         charmtone.Pepper,
		bgLeastVisible: charmtone.BBQ,
		bgLessVisible:  charmtone.Char,
		bgMostVisible:  charmtone.Iron,

		separator: charmtone.Char,

		destructive:       charmtone.Coral,
		error:             charmtone.Sriracha,
		warningSubtle:     charmtone.Zest,
		warning:           charmtone.Mustard,
		attention:         charmtone.Tang,
		busy:              charmtone.Citron,
		info:              charmtone.Malibu,
		infoMoreSubtle:    charmtone.Sardine,
		infoMostSubtle:    charmtone.Damson,
		success:           charmtone.Julep,
		successMoreSubtle: charmtone.Bok,
		successMostSubtle: charmtone.Guac,

		// ANSI 16-color palette for remapping raw terminal output
		// (e.g. bang-mode shell commands) onto legible Charmtone colors.
		ansiBlack:   charmtone.BBQ,
		ansiRed:     charmtone.Coral,
		ansiGreen:   charmtone.Guac,
		ansiYellow:  charmtone.Mustard,
		ansiBlue:    charmtone.Charple,
		ansiMagenta: charmtone.Dolly,
		ansiCyan:    charmtone.Malibu,
		ansiWhite:   charmtone.Smoke,

		ansiBrightBlack:   charmtone.Iron,
		ansiBrightRed:     charmtone.Tuna,
		ansiBrightGreen:   charmtone.Julep,
		ansiBrightYellow:  charmtone.Zest,
		ansiBrightBlue:    charmtone.Guppy,
		ansiBrightMagenta: charmtone.Blush,
		ansiBrightCyan:    charmtone.Sardine,
		ansiBrightWhite:   charmtone.Salt,
	})

	// Bang ! prompt overrides - use Salt/Hazy/Larple colors.
	s.Editor.PromptBangIconFocused = s.Editor.PromptBangIconFocused.
		Foreground(charmtone.Salt).
		Background(charmtone.Hazy)
	s.Editor.PromptBangDotsFocused = s.Editor.PromptBangDotsFocused.
		Foreground(charmtone.Hazy)
	s.Editor.PromptBangDotsBlurred = s.Editor.PromptBangDotsBlurred.
		Foreground(charmtone.Larple)

	// Shell bar/prompt overrides - use Charple/Iron/Hazy colors.
	s.Messages.ShellBarFocused = s.Messages.ShellBarFocused.
		BorderForeground(charmtone.Charple)
	s.Messages.ShellBarBlurred = s.Messages.ShellBarBlurred.
		BorderForeground(charmtone.Iron)
	s.Messages.ShellPrompt = s.Messages.ShellPrompt.
		Foreground(charmtone.Hazy)
	s.Messages.ShellPromptBlurred = s.Messages.ShellPromptBlurred.
		Foreground(charmtone.Hazy)

	return s
}

// HypercrushObsidiana returns the Hypercrush dark theme.
func HypercrushObsidiana() Styles {
	return CharmtonePantera()
}

// PMOVESArmor returns the PMOVES dark theme using agent signature colors.
// Primary = crush sky blue (#0EA5E9), secondary = claude-opus violet
// (#7C3AED), accent = darkxside crimson (#E11D48). Matches the --pm-*
// design tokens from pmoves/design/themes/pmoves-armor.json.
func PMOVESArmor() Styles {
	s := quickStyle(quickStyleOpts{
		primary:   pmovesSkyBlue,
		secondary: pmovesViolet,
		accent:    pmovesCrimson,
		keyword:   pmovesEmerald,

		fgBase:       pmovesInk,
		fgMoreSubtle: pmovesInkSubtle,
		fgSubtle:     color.RGBA{R: 0x80, G: 0x80, B: 0x88, A: 0xFF},
		fgMostSubtle: color.RGBA{R: 0x50, G: 0x50, B: 0x58, A: 0xFF},

		onPrimary: pmovesInk,

		bgBase:         pmovesVoid,
		bgLeastVisible: pmovesSurface,
		bgLessVisible:  color.RGBA{R: 0x0A, G: 0x0A, B: 0x0F, A: 0xFF},
		bgMostVisible:  color.RGBA{R: 0x1A, G: 0x1A, B: 0x24, A: 0xFF},

		separator: color.RGBA{R: 0x20, G: 0x20, B: 0x28, A: 0xFF},

		destructive:       pmovesCrimson,
		error:             color.RGBA{R: 0xDC, G: 0x26, B: 0x26, A: 0xFF},
		warningSubtle:     color.RGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF},
		warning:           color.RGBA{R: 0xD9, G: 0x77, B: 0x06, A: 0xFF},
		attention:         pmovesSkyBlue,
		busy:              pmovesTeal,
		info:              pmovesSkyBlue,
		infoMoreSubtle:    color.RGBA{R: 0x7D, G: 0xD3, B: 0xFC, A: 0xFF},
		infoMostSubtle:    color.RGBA{R: 0x38, G: 0xBD, B: 0xF8, A: 0xFF},
		success:           pmovesEmerald,
		successMoreSubtle: pmovesTeal,
		successMostSubtle: color.RGBA{R: 0x2D, G: 0xD4, B: 0xBF, A: 0xFF},

		ansiBlack:   pmovesSurface,
		ansiRed:     pmovesCrimson,
		ansiGreen:   pmovesEmerald,
		ansiYellow:  color.RGBA{R: 0xF5, G: 0x9E, B: 0x0B, A: 0xFF},
		ansiBlue:    pmovesSkyBlue,
		ansiMagenta: pmovesViolet,
		ansiCyan:    pmovesTeal,
		ansiWhite:   pmovesInkSubtle,

		ansiBrightBlack:   color.RGBA{R: 0x1A, G: 0x1A, B: 0x24, A: 0xFF},
		ansiBrightRed:     color.RGBA{R: 0xF8, G: 0x71, B: 0x71, A: 0xFF},
		ansiBrightGreen:   color.RGBA{R: 0x34, G: 0xD3, B: 0x99, A: 0xFF},
		ansiBrightYellow:  color.RGBA{R: 0xFB, G: 0xBF, B: 0x24, A: 0xFF},
		ansiBrightBlue:    color.RGBA{R: 0x38, G: 0xBD, B: 0xF8, A: 0xFF},
		ansiBrightMagenta: color.RGBA{R: 0xA7, G: 0x8B, B: 0xFA, A: 0xFF},
		ansiBrightCyan:    color.RGBA{R: 0x2D, G: 0xD4, B: 0xBF, A: 0xFF},
		ansiBrightWhite:   pmovesInk,
	})

	return s
}
