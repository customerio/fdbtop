package main

import (
	"math"
	"strings"
	"time"
)

func ShowLatencyScreen() {
	const (
		COL0 = 1
		COL1 = COL0 + 11
		COL2 = COL1 + 12 + MAX_RW_WIDTH
		COL3 = COL2 + 12 + MAX_RW_WIDTH
	)

	SetColor("DarkCyan")
	WriteAtS(COL0, 5, "Elapsed")
	WriteAtS(COL1, 5, "  Commit (ms)")
	WriteAtS(COL2, 5, "    Read (ms)")
	WriteAtS(COL3, 5, "   Start (ms)")

	maxCommit := GetMax(History, func(m HistoryMetric) float64 { return m.LatencyCommit })
	maxRead := GetMax(History, func(m HistoryMetric) float64 { return m.LatencyRead })
	maxStart := GetMax(History, func(m HistoryMetric) float64 { return m.LatencyStart })
	scaleCommit := GetMaxScale(maxCommit)
	scaleRead := GetMaxScale(maxRead)
	scaleStart := GetMaxScale(maxStart)

	SetColor("DarkGreen")
	WriteAt(COL1+14, 5, "%35.3f", maxCommit*1000)
	WriteAt(COL2+14, 5, "%35.3f", maxRead*1000)
	WriteAt(COL3+14, 5, "%18.3f", maxStart*1000)

	y := 7 + len(History) - 1
	_, ScreenHeight := screen.Size()
	for _, metric := range History {
		if y < ScreenHeight {
			SetColor("DarkGray")
			WriteAt(1, y, "%9s | %8s %40s | %8s %40s | %10s %20s |", time.Duration(math.Round(metric.LocalTime.Seconds()))*time.Second, "", "", "", "", "", "")

			if metric.Available {
				isMaxRead := maxCommit > 0 && metric.LatencyCommit == maxCommit
				isMaxWrite := maxRead > 0 && metric.LatencyRead == maxRead
				isMaxSpeed := maxStart > 0 && metric.LatencyStart == maxStart

				SetColorIf(isMaxRead, "Cyan", LatencyColor(metric.LatencyCommit))
				WriteAt(COL1, y, "%8.3f", metric.LatencyCommit*1000)

				SetColorIf(isMaxWrite, "Cyan", LatencyColor(metric.LatencyRead))
				WriteAt(COL2, y, "%8.3f", metric.LatencyRead*1000)

				SetColorIf(isMaxSpeed, "Cyan", LatencyColor(metric.LatencyStart))
				WriteAt(COL3, y, "%10.3f", metric.LatencyStart*1000)

				SetColor("Green")
				WriteAtS(COL1+9, y, strings.Repeat("|", Bar(metric.LatencyCommit, scaleCommit, MAX_RW_WIDTH)))
				WriteAtS(COL2+9, y, strings.Repeat("|", Bar(metric.LatencyRead, scaleRead, MAX_RW_WIDTH)))
				WriteAtS(COL3+11, y, strings.Repeat("|", Bar(metric.LatencyStart, scaleStart, MAX_WS_WIDTH)))
			} else {
				SetColor("Red")
				WriteAt(COL1, y, "%8s", "x")
				WriteAt(COL2, y, "%8s", "x")
				WriteAt(COL3, y, "%8s", "x")
			}
		}
		y--
	}
}

func GetMax(metrics []HistoryMetric, selector func(HistoryMetric) float64) float64 {
	max := math.NaN()
	for _, item := range metrics {
		x := selector(item)
		if math.IsNaN(max) || x > max {
			max = x
		}
	}
	return max
}

func GetMaxScale(max float64) float64 {
	return math.Pow(10, math.Ceil(math.Log10(max)))
}

func FrequencyColor(hz float64) string {
	if hz >= 100 {
		return "White"
	} else if hz >= 10 {
		return "Gray"
	} else {
		return "DarkCyan"
	}
}

func DiskSpeedColor(bps float64) string {
	if bps >= 1048576 {
		return "White"
	} else if bps >= 1024 {
		return "Gray"
	} else {
		return "DarkGray"
	}
}

func LatencyColor(x float64) string {
	if x >= 1 {
		return "Red"
	} else if x >= 0.1 {
		return "Yellow"
	} else if x >= 0.01 {
		return "White"
	} else {
		return "Gray"
	}
}

func GetBarChar(scale float64) string {
	if scale >= 1000000 {
		return "@"
	} else if scale >= 100000 {
		return "#"
	} else if scale >= 10000 {
		return "|"
	} else {
		return ":"
	}
}

func Bar(hz, scale float64, width int) int {
	if hz == 0 {
		return 0
	}

	x := hz * float64(width) / scale
	if x < 0 || math.IsNaN(x) {
		x = 0
	} else if x > float64(width) {
		x = float64(width)
	}

	if x == 0 {
		return 0
	} else if x < 1 {
		return 1
	} else {
		return int(math.Round(x))
	}
}

func BarGraph(hz, scale float64, width int, full rune, half, nonZero string) string {
	x := hz * float64(width*2) / scale
	if x < 0 || math.IsNaN(x) {
		x = 0
	} else if x > float64(width*2) {
		x = float64(width * 2)
	}

	n := int(math.Round(x))

	if n == 0 {
		if hz == 0 {
			return ""
		} else {
			return nonZero
		}
	} else if n == 1 {
		return half
	} else if n%2 == 1 {
		return strings.Repeat(string(full), n/2) + half
	} else {
		return strings.Repeat(string(full), n/2)
	}
}
