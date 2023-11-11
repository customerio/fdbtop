package main

import (
	"strings"
)

func ShowTransactionsScreen() {
	const (
		COL0 = 1
		COL1 = COL0 + 11
		COL2 = COL1 + 12 + MAX_RW_WIDTH
		COL3 = COL2 + 12 + MAX_RW_WIDTH
	)

	SetColor("DarkCyan")
	WriteAt(COL0, 5, "Elapsed")
	WriteAt(COL1, 5, "Started (tps)")
	WriteAt(COL2, 5, "Committed (tps)")
	WriteAt(COL3, 5, "Conflicted (tps)")

	maxStarted := GetMax(History, func(m HistoryMetric) float64 { return m.TransStarted })
	maxCommitted := GetMax(History, func(m HistoryMetric) float64 { return m.TransCommitted })
	maxConflicted := GetMax(History, func(m HistoryMetric) float64 { return m.TransConflicted })
	scaleStarted := GetMaxScale(maxStarted)
	scaleCommitted := GetMaxScale(maxCommitted)
	scaleConflicted := GetMaxScale(maxConflicted)

	SetColor("DarkGreen")
	WriteAt(COL1+14, 5, "%35.0f", maxStarted)
	WriteAt(COL2+16, 5, "%33.0f", maxCommitted)
	WriteAt(COL3+16, 5, "%15.0f", maxConflicted)

	_, ScreenHeight := screen.Size()

	y := 7 + len(History) - 1
	for _, metric := range History {
		if y < ScreenHeight {
			SetColor("DarkGray")
			WriteAt(1, y, "%9s | %8s %40s | %8s %40s | %10s %20s |", TimeSpanInSecondsWithRounding(metric.LocalTime.Seconds()), "", "", "", "", "", "")

			if metric.Available {
				isMaxRead := maxStarted > 0 && metric.LatencyCommit == maxStarted
				isMaxWrite := maxCommitted > 0 && metric.LatencyRead == maxCommitted
				isMaxSpeed := maxConflicted > 0 && metric.LatencyStart == maxConflicted

				SetColorIf(isMaxRead, "Cyan", FrequencyColor(metric.TransStarted))
				WriteAt(COL1, y, "%8.0f", metric.TransStarted)
				SetColorIf(isMaxWrite, "Cyan", FrequencyColor(metric.TransCommitted))
				WriteAt(COL2, y, "%8.0f", metric.TransCommitted)
				SetColorIf(isMaxSpeed, "Cyan", FrequencyColor(metric.TransConflicted))
				WriteAt(COL3, y, "%8.1f", metric.TransConflicted)

				SetColorIf(metric.TransStarted > 10, "Green", "DarkGreen")
				if metric.TransStarted == 0 {
					WriteAt(COL1+9, y, "-")
				} else {
					WriteAt(COL1+9, y, strings.Repeat("|", Bar(metric.TransStarted, scaleStarted, MAX_RW_WIDTH)))
				}
				SetColorIf(metric.TransCommitted > 10, "Green", "DarkGreen")
				if metric.TransCommitted == 0 {
					WriteAt(COL2+9, y, "-")
				} else {
					WriteAt(COL2+9, y, strings.Repeat("|", Bar(metric.TransCommitted, scaleCommitted, MAX_RW_WIDTH)))
				}
				if metric.TransConflicted > 1000 {
					SetColor("Red")
				} else if metric.TransConflicted > 10 {
					SetColor("Green")
				} else {
					SetColor("DarkGreen")
				}
				if metric.TransConflicted == 0 {
					WriteAt(COL3+9, y, "-")
				} else {
					WriteAt(COL3+9, y, strings.Repeat("|", Bar(metric.TransConflicted, scaleConflicted, MAX_WS_WIDTH)))
				}
			} else {
				SetColor("DarkRed")
				WriteAt(COL1, y, "%8s", "x")
				WriteAt(COL2, y, "%8s", "x")
				WriteAt(COL3, y, "%8s", "x")
			}
		}

		y--
	}

	SetColor("Gray")
}
