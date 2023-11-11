package main

import (
	"math"
	"time"
)

func RepaintTopBar() {
	const (
		TOP_COL0 = 1
		TOP_COL1 = TOP_COL0 + 24
		TOP_COL2 = TOP_COL1 + 26
		TOP_COL3 = TOP_COL2 + 36
		TOP_COL4 = TOP_COL3 + 22

		TOP_ROW0 = 0
		TOP_ROW1 = 1
		TOP_ROW2 = 2
	)

	SetColor("DarkGray")
	WriteAt(TOP_COL0, TOP_ROW0, "Reads  : %8s Hz", "")
	WriteAt(TOP_COL0, TOP_ROW1, "Writes : %8s Hz", "")
	WriteAt(TOP_COL0, TOP_ROW2, "Written: %8s MB/s", "")

	WriteAt(TOP_COL1, TOP_ROW0, "Total K/V: %10s MB", "")
	WriteAt(TOP_COL1, TOP_ROW1, "Disk Used: %10s MB", "")
	WriteAt(TOP_COL1, TOP_ROW2, "Shards: %05s x%6s MB", "", "")

	WriteAt(TOP_COL2, TOP_ROW0, "Server Time : %19s", "")
	WriteAt(TOP_COL2, TOP_ROW1, "Client Time : %19s", "")
	WriteAt(TOP_COL2, TOP_ROW2, "Read Version: %10s", "")

	WriteAt(TOP_COL3, TOP_ROW0, "Coordinat.: %10s", "")
	WriteAt(TOP_COL3, TOP_ROW1, "Storage   : %10s", "")
	WriteAt(TOP_COL3, TOP_ROW2, "Redundancy: %10s", "")

	WriteAt(TOP_COL4, TOP_ROW0, "State: %10s", "")
	WriteAt(TOP_COL4, TOP_ROW1, "Data : %20s", "")
	WriteAt(TOP_COL4, TOP_ROW2, "Perf.: %20s", "")
}

func UpdateTopBar(status FdbStatus, current HistoryMetric) {
	const (
		TOP_COL0 = 1
		TOP_COL1 = TOP_COL0 + 24
		TOP_COL2 = TOP_COL1 + 26
		TOP_COL3 = TOP_COL2 + 36
		TOP_COL4 = TOP_COL3 + 22

		TOP_ROW0 = 0
		TOP_ROW1 = 1
		TOP_ROW2 = 2
	)

	SetColor("White")
	WriteAt(TOP_COL0+9, TOP_ROW0, "%8.0f", current.ReadsPerSecond)
	WriteAt(TOP_COL0+9, TOP_ROW1, "%8.0f", current.WritesPerSecond)
	WriteAt(TOP_COL0+9, TOP_ROW2, "%8.2f", MegaBytes(int64(current.WrittenBytesPerSecond)))

	WriteAt(TOP_COL1+11, TOP_ROW0, "%10.1f", MegaBytes(status.Cluster.Data.TotalKvSizeBytes))
	WriteAt(TOP_COL1+11, TOP_ROW1, "%10.1f", MegaBytes(status.Cluster.Data.TotalDiskUsedBytes))
	WriteAt(TOP_COL1+8, TOP_ROW2, "%5.0d", status.Cluster.Data.PartitionsCount)
	WriteAt(TOP_COL1+15, TOP_ROW2, "%6.1f", MegaBytes(status.Cluster.Data.AveragePartitionSizeBytes))

	serverTime := time.Unix(current.Timestamp, 0).UTC()
	clientTime := time.Unix(status.Client.Timestamp, 0).UTC()

	format := "02 Jan 06 15:04:05"
	WriteAt(TOP_COL2+14, TOP_ROW0, "%-19s", serverTime.Format(format))
	SetColorIf(math.Abs(serverTime.Sub(clientTime).Seconds()) >= 20, "Red", "White")
	WriteAt(TOP_COL2+14, TOP_ROW1, "%-19s", clientTime.Format(format))
	SetColor("White")
	WriteAt(TOP_COL2+14, TOP_ROW2, "%d", current.ReadVersion)

	trimMax := func(s string, max int) string {
		if len(s) > max {
			return s[:max]
		}
		return s
	}

	WriteAt(TOP_COL3+12, TOP_ROW0, "%-10d", status.Cluster.Configuration.CoordinatorsCount)
	WriteAt(TOP_COL3+12, TOP_ROW1, "%-10s", trimMax(status.Cluster.Configuration.StorageEngine, 9))
	WriteAt(TOP_COL3+12, TOP_ROW2, "%-10s", status.Cluster.Configuration.RedundancyMode)

	//if !status.Client.DatabaseAvailable {
	if false {
		SetColor("Red")
		WriteAtS(TOP_COL4+7, TOP_ROW0, "UNAVAILABLE")
	} else {
		SetColor("Green")
		WriteAtS(TOP_COL4+7, TOP_ROW0, "Available  ")
	}
	SetColor("White")
	WriteAt(TOP_COL4+7, TOP_ROW1, "%-40s", status.Cluster.Data.State.Name)
	WriteAt(TOP_COL4+7, TOP_ROW2, "%-40s", status.Cluster.Qos.PerformanceLimitedBy.Name)
}
