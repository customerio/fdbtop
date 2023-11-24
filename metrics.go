package main

import (
	"math"
	"strings"
	"time"
)

const (
	MAX_RW_WIDTH = 40
	MAX_WS_WIDTH = 20
	maxHistory   = 100
)

type HistoryMetric struct {
	Available             bool
	LocalTime             time.Duration
	ReadVersion           int64
	Timestamp             int64
	ReadsPerSecond        float64
	WritesPerSecond       float64
	WrittenBytesPerSecond float64
	TransStarted          float64
	TransCommitted        float64
	TransConflicted       float64
	LatencyCommit         float64
	LatencyRead           float64
	LatencyStart          float64
}

var History []HistoryMetric

func ShowMetricsScreen() {
	const (
		COL0 = 1
		COL1 = COL0 + 11
		COL2 = COL1 + 12 + MAX_RW_WIDTH
		COL3 = COL2 + 12 + MAX_RW_WIDTH
	)

	SetColor("DarkCyan")
	WriteAt(COL0, 5, "Elapsed")
	WriteAt(COL1, 5, "   Reads (Hz)")
	WriteAt(COL2, 5, "  Writes (Hz)")
	WriteAt(COL3, 5, "Disk Speed (MB/s)")

	maxRead := GetMax(History, func(m HistoryMetric) float64 { return m.ReadsPerSecond })
	maxWrite := GetMax(History, func(m HistoryMetric) float64 { return m.WritesPerSecond })
	maxSpeed := GetMax(History, func(m HistoryMetric) float64 { return m.WrittenBytesPerSecond })
	scaleRead := GetMaxScale(maxRead)
	scaleWrite := GetMaxScale(maxWrite)
	scaleSpeed := GetMaxScale(maxSpeed)

	SetColor("Green")
	WriteAt(COL1+14, 5, "%35.0f", maxRead)
	WriteAt(COL2+14, 5, "%35.0f", maxWrite)
	WriteAt(COL3+18, 5, "%13.3f", MegaBytes(int64(maxSpeed)))

	y := 7 + len(History) - 1
	_, ScreenHeight := screen.Size()
	for _, metric := range History {
		if y < ScreenHeight-1 {
			SetColor("DarkGray")
			WriteAt(1, y, "%9s | %8s %40s | %8s %40s | %10s %20s |", time.Duration(math.Round(float64(metric.LocalTime.Seconds())))*time.Second, "", "", "", "", "", "")

			if metric.Available {
				isMaxRead := maxRead > 0 && metric.ReadsPerSecond == maxRead
				isMaxWrite := maxWrite > 0 && metric.WritesPerSecond == maxWrite
				isMaxSpeed := maxSpeed > 0 && metric.WrittenBytesPerSecond == maxSpeed

				SetColorIf(isMaxRead, "Cyan", FrequencyColor(metric.ReadsPerSecond))
				WriteAt(COL1, y, "%8.0f", metric.ReadsPerSecond)
				SetColorIf(isMaxWrite, "Cyan", FrequencyColor(metric.WritesPerSecond))
				WriteAt(COL2, y, "%8.0f", metric.WritesPerSecond)
				SetColorIf(isMaxSpeed, "Cyan", DiskSpeedColor(metric.WrittenBytesPerSecond))
				WriteAt(COL3, y, "%10.3f", MegaBytes(int64(metric.WrittenBytesPerSecond)))

				SetColorIf(metric.ReadsPerSecond > 10, "Green", "DarkCyan")
				if metric.ReadsPerSecond == 0 {
					WriteAt(COL1+9, y, "-")
				} else {
					WriteAt(COL1+9, y, strings.Repeat(GetBarChar(metric.ReadsPerSecond), Bar(metric.ReadsPerSecond, scaleRead, MAX_RW_WIDTH)))
				}
				SetColorIf(metric.WritesPerSecond > 10, "Green", "DarkCyan")
				if metric.WritesPerSecond == 0 {
					WriteAt(COL2+9, y, "-")
				} else {
					WriteAt(COL2+9, y, strings.Repeat(GetBarChar(metric.WritesPerSecond), Bar(metric.WritesPerSecond, scaleWrite, MAX_RW_WIDTH)))
				}
				SetColorIf(metric.WrittenBytesPerSecond > 1000, "Green", "DarkCyan")
				if metric.WrittenBytesPerSecond == 0 {
					WriteAt(COL3+11, y, "-")
				} else {
					WriteAt(COL3+11, y, strings.Repeat(GetBarChar(metric.WrittenBytesPerSecond/1000), Bar(metric.WrittenBytesPerSecond, scaleSpeed, MAX_WS_WIDTH)))
				}
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
