package main

import (
	"math"
	"sort"
	"strings"
	"time"
)

func ShowProcessesScreen(status FdbStatus) {
	const (
		CPU_BARSZ     = 10
		MEM_BARSZ     = 5
		HDD_BARSZ     = 10
		SP            = 1
		BAR           = SP + 1 + SP
		COL_HOST      = 1
		LEN_HOST      = 16
		COL_NET       = COL_HOST + LEN_HOST + BAR
		LEN_NET       = 4 + SP + 8 + SP + 8
		COL_CPU       = COL_NET + LEN_NET + BAR
		LEN_CPU       = 5 + 1 + SP + CPU_BARSZ
		COL_MEM_USED  = COL_CPU + LEN_CPU + BAR
		LEN_MEM_USED  = 7 + SP
		COL_MEM_TOTAL = COL_MEM_USED + LEN_MEM_USED
		LEN_MEM_TOTAL = 9 + MEM_BARSZ
		COL_DISK      = COL_MEM_TOTAL + LEN_MEM_TOTAL
		LEN_DISK      = 8 + SP + 7 + SP + 7
		COL_HDD       = COL_DISK + LEN_DISK + BAR
		LEN_HDD       = 5 + 1 + SP + HDD_BARSZ
		COL_ROLES     = COL_HDD + LEN_HDD + BAR
		LEN_ROLES     = 11
		COL9          = COL_ROLES + LEN_ROLES + BAR
	)

	SetColor("DarkCyan")
	WriteAtS(COL_HOST, 5, "Address (port)")
	WriteAtS(COL_NET, 5, "Network (Mbps)")
	WriteAtS(COL_NET, 6, " Cnx     Recv     Sent")
	WriteAtS(COL_CPU, 5, "CPU Activity")
	WriteAtS(COL_CPU, 6, " %core")

	WriteAtS(COL_MEM_USED, 5, "Memory Activity (GB)")
	WriteAtS(COL_MEM_USED, 6, " Used / Total")
	WriteAtS(COL_DISK, 5, "Disk Activity (MB/s)")
	WriteAtS(COL_DISK, 6, "   Queue Queried Mutated   HDD Busy")
	WriteAtS(COL_ROLES, 5, "Roles")
	WriteAtS(COL9, 5, "Uptime")

	LastProcessYMax = 0

	if debugLayout {
		SetColor("DarkGray")
		WriteAtS(COL_HOST, 4, "0 - - - - - -")
		WriteAtS(COL_NET, 4, "1 - - - - - -")
		WriteAtS(COL_CPU, 4, "3 - - - - - -")
		WriteAtS(COL_MEM_USED, 4, "4 - - - - - -")
		WriteAtS(COL_MEM_TOTAL, 4, "5 - - - - - -")
		WriteAtS(COL_DISK, 4, "6 - - - - - -")
		WriteAtS(COL_HDD, 4, "7 - - - - - -")
		WriteAtS(COL_ROLES, 4, "8 - - - - - -")
	}

	if len(status.Cluster.Machines) == 0 {
		SetColor("Red")
		WriteAt(COL_HOST, 4, "No machines found!")
		//TODO display error message?
		return
	} else {
		WriteAt(COL_HOST, 4, "                  ")
	}

	var maxVersion string
	for _, p := range status.Cluster.Processes {
		if p.Version > maxVersion {
			maxVersion = p.Version
		}
	}

	ScreenWidth, ScreenHeight := screen.Size()
	emptyLine := strings.Repeat(" ", ScreenWidth)

	roleMap := &RoleMap{}

	y := 7
	//machines := status.Cluster.Machines
	var machines []FdbMachine
	for id, m := range status.Cluster.Machines {
		m.Id = id
		machines = append(machines, m)
	}

	sort.Slice(machines, func(i, j int) bool {
		return machines[i].Address < machines[j].Address
	})

	for _, machine := range machines {
		var procs []FdbProcess
		for _, p := range status.Cluster.Processes {
			if p.MachineId == machine.Id {
				procs = append(procs, p)
			}
		}

		sort.Slice(procs, func(i, j int) bool {
			return procs[i].Address < procs[j].Address
		})

		var storageBytes, queueDiskBytes, totalCnx, totalQueueSize int64
		var totalDiskBusy, totalMutationBytes, totalQueriedBytes float64

		roleMap.Reset()
		for _, proc := range procs {
			totalDiskBusy += proc.Disk.Busy
			totalCnx += proc.Network.CurrentConnections
			for _, role := range proc.Roles {
				roleMap.Add(role.Role)
				switch role.Role {
				case StorageRoleMetrics:
					totalMutationBytes += role.MutationBytes.Hz
					totalQueriedBytes += role.BytesQueried.Hz
					totalQueueSize += int64(role.InputBytes.Counter - role.DurableBytes.Counter)
					storageBytes += role.StoredBytes
				case LogRoleMetrics:
					queueDiskBytes += role.QueueDiskUsedBytes
				}
			}
		}

		totalDiskBusy /= float64(len(procs))

		SetColor("DarkGray")
		WriteAt(1, y, "%s", strings.ReplaceAll("                 | ____ ________ ________ | _____% __________ | _____ / _____ _____ | ________ _______ _______ | _____% __________ | ___________ | ", "_", " "))

		SetColor("White")
		WriteAt(COL_HOST, y, "%s", machine.Address)
		SetColor(MapConnectionsToColor(totalCnx))
		WriteAt(COL_NET, y, "%4.0d", totalCnx)
		SetColor(MapMegabitsToColor(machine.Network.MegabitsReceived.Hz))
		WriteAt(COL_NET+5, y, "%8.2f", machine.Network.MegabitsReceived.Hz)
		SetColor(MapMegabitsToColor(machine.Network.MegabitsSent.Hz))
		WriteAt(COL_NET+14, y, "%8.2f", machine.Network.MegabitsSent.Hz)

		WriteAt(COL_CPU, y, "%5.1f", machine.Cpu.LogicalCoreUtilization*100)

		WriteAt(COL_MEM_USED, y, "%5.1f", GigaBytes(machine.Memory.CommittedBytes))
		WriteAt(COL_MEM_TOTAL, y, "%5.1f", GigaBytes(machine.Memory.TotalBytes))

		SetColor("DarkGray")
		WriteAt(COL_ROLES, y, "%11s", roleMap.String())

		if machine.Cpu.LogicalCoreUtilization >= 0.9 {
			SetColor("DarkRed")
		} else {
			SetColor("DarkGreen")
		}
		WriteAt(COL_CPU+7, y, "%-10s", BarGraph(machine.Cpu.LogicalCoreUtilization, 1, CPU_BARSZ, '=', ":", ".")) // 1 = all the (logical) cores

		memRatio := float64(machine.Memory.CommittedBytes) / float64(machine.Memory.TotalBytes)
		if memRatio >= 0.95 {
			SetColor("Red")
		} else if memRatio >= 0.79 {
			SetColor("DarkYellow")
		} else {
			SetColor("Green")
		}
		WriteAt(COL_MEM_TOTAL+6, y, "%-5s", BarGraph(float64(machine.Memory.CommittedBytes), float64(machine.Memory.TotalBytes), MEM_BARSZ, '=', ":", "."))

		if roleMap.Log || roleMap.Storage {
			SetColor(MapQueueSizeToColor(float64(totalQueueSize)))
			WriteAt(COL_DISK, y, "%8s", FriendlyBytes(totalQueueSize))
		}
		if roleMap.Storage {
			SetColor(MapDiskOpsToColor(totalQueriedBytes))
			WriteAt(COL_DISK+9, y, "%7.1f", MegaBytes(int64(totalQueriedBytes)))
			SetColor(MapDiskOpsToColor(totalMutationBytes))
			WriteAt(COL_DISK+17, y, "%7.1f", MegaBytes(int64(totalMutationBytes)))
		}

		SetColor("Gray")
		WriteAt(COL_HDD, y, "%5.1f", totalDiskBusy*100)
		if totalDiskBusy == 0.0 {
			SetColor("DarkGray")
		} else if totalDiskBusy >= 0.95 {
			SetColor("DarkRed")
		} else {
			SetColor("DarkGreen")
		}
		WriteAt(COL_HDD+7, y, "%-10s", BarGraph(totalDiskBusy, 1, 10, '=', ":", "."))

		y++

		for _, proc := range procs {
			p := strings.Index(proc.Address, ":")
			port := ""
			if p >= 0 {
				port = proc.Address[p+1:]
			} else {
				port = proc.Address
			}

			roleMap.Reset()
			var mutationBytes, queriedBytes float64
			var queueSize int64
			for _, role := range proc.Roles {
				roleMap.Add(role.Role)
				if role.Role == StorageRoleMetrics {
					mutationBytes += role.MutationBytes.Hz
					queriedBytes += role.BytesQueried.Hz
					queueSize += int64(role.InputBytes.Counter - role.DurableBytes.Counter)
				} else if role.Role == LogRoleMetrics {
					queueSize += int64(role.InputBytes.Counter - role.DurableBytes.Counter)
				}
			}

			if y < ScreenHeight {
				SetColor("DarkGray")
				WriteAtS(COL_HOST, y, strings.ReplaceAll("_______ | ______ | ____ ________ ________ | _____% __________ | _____ / _____ _____ | ________ _______ _______ | _____% __________ | ___________ |", "_", " "))

				SetColorIf(proc.Version != maxVersion, "DarkRed", "DarkGray")
				WriteAt(COL_HOST+10, y, "%6s", proc.Version)

				SetColorIf(proc.Excluded, "DarkRed", "Gray")
				WriteAt(COL_HOST, y, "%7s", port)

				SetColor(MapConnectionsToColor(proc.Network.CurrentConnections))
				WriteAt(COL_NET, y, "%4d", proc.Network.CurrentConnections)
				SetColor(MapMegabitsToColor(proc.Network.MegabitsReceived.Hz))
				//WriteAt( COL_NET+5, y, "%8.2f", Nice(proc.Network.MegabitsReceived.Hz, "-", 0.005, "~"))
				WriteAt(COL_NET+5, y, "%8s", Nice(proc.Network.MegabitsReceived.Hz, "-", 0.005, "~"))
				SetColor(MapMegabitsToColor(proc.Network.MegabitsSent.Hz))
				//WriteAt( COL_NET+14, y, "%8.2f", Nice(proc.Network.MegabitsSent.Hz, "-", 0.005, "~"))
				WriteAt(COL_NET+14, y, "%8s", Nice(proc.Network.MegabitsSent.Hz, "-", 0.005, "~"))

				cpuUsage := proc.Cpu.UsageCores
				if cpuUsage >= 0.95 {
					SetColor("DarkRed")
				} else if cpuUsage >= 0.75 {
					SetColor("DarkYellow")
				} else if cpuUsage >= 0.2 {
					SetColor("Gray")
				} else {
					SetColor("DarkGray")
				}
				WriteAt(COL_CPU, y, "%5.1f", cpuUsage*100)
				if cpuUsage >= 0.95 {
					SetColor("DarkRed")
				} else if cpuUsage >= 0.75 {
					SetColor("DarkYellow")
				} else {
					SetColor("DarkGreen")
				}
				WriteAt(COL_CPU+7, y, "%-10s", BarGraph(proc.Cpu.UsageCores, 1, CPU_BARSZ, '|', ":", "."))

				memoryUsed := proc.Memory.UsedBytes - proc.Memory.UnusedAllocatedMemory
				memoryAllocated := proc.Memory.UsedBytes
				SetColor(MapMemoryToColor(memoryUsed))
				WriteAt(COL_MEM_USED, y, "%5.1f", GigaBytes(memoryUsed))
				SetColor(MapMemoryToColor(memoryAllocated))
				WriteAt(COL_MEM_TOTAL, y, "%5.1f", GigaBytes(memoryAllocated))
				if float64(memoryUsed) >= 0.9*float64(proc.Memory.LimitBytes) {
					SetColor("DarkRed")
				} else if float64(memoryUsed) >= 0.75*float64(proc.Memory.LimitBytes) {
					SetColor("DarkYellow")
				} else {
					SetColor("DarkGreen")
				}
				WriteAt(COL_MEM_TOTAL+6, y, "%-5s", BarGraph(float64(memoryUsed), float64(machine.Memory.CommittedBytes), MEM_BARSZ, '|', ":", "."))

				if roleMap.Log || roleMap.Storage {
					SetColor(MapQueueSizeToColor(float64(queueSize)))
					WriteAt(COL_DISK, y, "%8s", FriendlyBytes(queueSize))
				}
				if roleMap.Storage {
					SetColor(MapDiskOpsToColor(queriedBytes))
					WriteAt(COL_DISK+9, y, "%7.1f", MegaBytes(int64(queriedBytes)))
					SetColor(MapDiskOpsToColor(mutationBytes))
					WriteAt(COL_DISK+17, y, "%7.1f", MegaBytes(int64(mutationBytes)))
				}

				SetColor("Gray")
				WriteAt(COL_HDD, y, "%5.1f", proc.Disk.Busy*100)
				if proc.Disk.Busy == 0.0 {
					SetColor("DarkGray")
				} else if proc.Disk.Busy >= 0.95 {
					SetColor("DarkRed")
				} else {
					SetColor("DarkGreen")
				}
				WriteAt(COL_HDD+7, y, "%-10s", strings.Repeat("|", Bar(proc.Disk.Busy, 1, 10)))

				SetColor("Gray")
				WriteAt(COL_ROLES, y, "%11s", roleMap.String())

				SetColor("DarkGray")
				WriteAt(COL9, y, "%11s", time.Duration(proc.UptimeSeconds)*time.Second)
			}

			y++
		}
		WriteAtS(COL_HOST, y, emptyLine)
		y++
	}
	y = int(math.Min(float64(y), float64(ScreenHeight)))
	for ; y < LastProcessYMax; y++ {
		WriteAtS(COL_HOST, y, emptyLine)
	}
	LastProcessYMax = y
}
