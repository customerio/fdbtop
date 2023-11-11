package main

import (
	"strings"
)

func RepaintBottomBar(mode DisplayMode) {
	ScreenWidth, ScreenHeight := screen.Size()

	SetBackground("DarkCyan")
	emptyLine := strings.Repeat(" ", ScreenWidth)
	WriteAtS(0, ScreenHeight-1, emptyLine)

	SetColorIf(mode == Metrics, "Black", "White")
	WriteAtS(0, ScreenHeight-1, " [M]etrics ")

	SetColorIf(mode == Transactions, "Black", "White")
	WriteAtS(11, ScreenHeight-1, " [T]ransactions ")

	SetColorIf(mode == Latency, "Black", "White")
	WriteAtS(27, ScreenHeight-1, " [L]atency ")

	SetColorIf(mode == Processes, "Black", "White")
	WriteAtS(38, ScreenHeight-1, " [P]rocesses ")

	SetColorIf(mode == Roles, "Black", "White")
	WriteAtS(51, ScreenHeight-1, " [R]roles ")
	SetBackground("Black")
}
