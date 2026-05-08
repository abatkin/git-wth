package ui

import (
	"os"
)

type ColorManager struct {
	Enabled bool
}

func (c ColorManager) Red(s string) string    { return c.Colorize("31", s) }
func (c ColorManager) Green(s string) string  { return c.Colorize("32", s) }
func (c ColorManager) Yellow(s string) string { return c.Colorize("33", s) }
func (c ColorManager) Cyan(s string) string   { return c.Colorize("36", s) }
func (c ColorManager) Grey(s string) string   { return c.Colorize("1;30", s) }
func (c ColorManager) Purple(s string) string { return c.Colorize("35", s) }

func (c ColorManager) Colorize(code, s string) string {
	if !c.Enabled {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func StdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}
