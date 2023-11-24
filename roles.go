package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

var debugLayout = true

func ShowRolesScreen(status FdbStatus) {
	const (
		CPU_BARSZ       = 10
		HDD_BARSZ       = 10
		SP              = 1
		BAR             = 3
		COL0            = 1
		COL_NET         = COL0 + 27
		COL2            = COL_NET + 8
		COL_CPU         = COL2 + 8 + 2
		LEN_CPU         = 10 + CPU_BARSZ
		COL_MEMORY      = COL_CPU + LEN_CPU
		LEN_MEMORY      = 8
		COL_HDD         = COL_MEMORY + LEN_MEMORY + BAR
		LEN_HDD         = 5 + 1 + SP + HDD_BARSZ
		COL_STORAGE     = COL_HDD + LEN_HDD + BAR
		LEN_STORAGE     = 8 + SP + 7 + SP + 7 + SP + 8
		COL_DATAVERSION = COL_STORAGE + LEN_STORAGE + BAR
		LEN_DATAVERSION = 14
		COL_KVSTORE     = COL_DATAVERSION + LEN_DATAVERSION + BAR
	)

	var y = 5

	LastProcessYMax = 0

	// Debug layout
	if debugLayout {
		SetColor("DarkGray")
		WriteAtS(COL0, 4, "0 - - - - - -")
		WriteAtS(COL_NET, 4, "1 - - - - - -")
		WriteAtS(COL2, 4, "2 - - - - - -")
		WriteAtS(COL_CPU, 4, "3 - - - - - -")
		WriteAtS(COL_MEMORY, 4, "4 - - - - - -")
		WriteAtS(COL_STORAGE, 4, "6 - - - - - -")
		WriteAtS(COL_HDD, 4, "7 - - - - - -")
		WriteAtS(COL_DATAVERSION, 4, "8 - - - - - -")
	}

	if len(status.Cluster.Machines) == 0 {
		SetColor("Red")
		WriteAtS(COL0, y, "No machines found!")
		return
	}

	ScreenWidth, ScreenHeight := screen.Size()
	emptyLine := strings.Repeat(" ", ScreenWidth)

	type Role struct {
		Process   FdbProcess
		Role      FdbRole
		MachineId string
	}
	byRoles := make(map[string][]Role)
	for _, process := range status.Cluster.Processes {
		for _, role := range process.Roles {
			byRoles[role.Role] = append(byRoles[role.Role], Role{Process: process, Role: role, MachineId: process.MachineId})
		}
	}

	maxDataVersion := func(kv []Role) int64 {
		max := int64(0)
		for _, role := range kv {
			if role.Role.Role == LogRoleMetrics {
				if role.Role.DataVersion > max {
					max = role.Role.DataVersion
				}
			}
		}
		return max
	}
	for _, roleId := range []string{"log", "storage", "proxy", "commit_proxy", "grv_proxy", "resolver", "master", "cluster_controller", "data_distributor", "ratekeeper"} {
		kv := byRoles[roleId]

		hasDisk := roleId == StorageRoleMetrics || roleId == LogRoleMetrics

		SetColor("Cyan")
		WriteAtS(0, y, emptyLine)
		WriteAtS(COL0, y, roleId)
		SetColor("DarkCyan")
		WriteAtS(COL_NET, y, "Network (Mbps)")
		WriteAtS(COL_CPU, y, "Processor Activity")
		WriteAtS(COL_MEMORY, y, "Memory")
		if hasDisk {
			WriteAtS(COL_HDD, y, "Disk Activity")
			WriteAtS(COL_STORAGE, y, "Storage Activity")
			WriteAtS(COL_DATAVERSION, y, "Data Version")
			WriteAtS(COL_KVSTORE, y, "KV Store")
		}
		y++

		WriteAtS(0, y, emptyLine)

		// WriteAt(COL1, y, "%8.3f", MegaBytes(machine.Network.MegabitsReceived.Hz * 125000))
		// WriteAt(COL2, y, "%8.3f", MegaBytes(machine.Network.MegabitsSent.Hz * 125000))
		// WriteAt(COL3, y, "%5.1f", machine.Cpu.LogicalCoreUtilization * 100)
		// WriteAt(COL4, y, "%5.1f", GigaBytes(machine.Memory.CommittedBytes))
		// WriteAt(COL5, y, "%5.1f", GigaBytes(machine.Memory.TotalBytes))
		// SetColor("DarkGray")
		// WriteAt(COL8, y, "%11s", map)

		// if machine.Cpu.LogicalCoreUtilization >= 0.9 {
		//     SetColor("Red")
		// } else {
		//     SetColor("Green")
		// }
		// WriteAt(COL3+7, y, "%-10s", BarGraph(machine.Cpu.LogicalCoreUtilization, 1, CPU_BARSZ, '|', ':')) // 1 = all the (logical) cores

		// memRatio := float64(machine.Memory.CommittedBytes) / float64(machine.Memory.TotalBytes)
		// if memRatio >= 0.95 {
		//     SetColor("Red")
		// } else if memRatio >= 0.79 {
		//     SetColor("DarkYellow")
		// } else {
		//     SetColor("Green")
		// }
		// WriteAt(COL5+6, y, "%-10s", BarGraph(machine.Memory.CommittedBytes, machine.Memory.TotalBytes, MEM_BARSZ, '|', ':'))

		// SetColor("DarkGray")
		// WriteAt(COL6, y, "S: %s; Q: %s, D: %s", FriendlyBytes(storageBytes), FriendlyBytes(queueDiskBytes), FriendlyBytes(int64(blahBytes)))

		maxLogTransaction := int64(0)
		SetColor("DarkCyan")
		WriteAtS(COL0, y, "         Address:Port")
		WriteAtS(COL_NET, y, "   Recv    Sent")
		WriteAtS(COL_MEMORY, y, " VM Size")
		WriteAtS(COL_CPU, y, "% CPU Core")
		if roleId == StorageRoleMetrics {
			WriteAtS(COL_HDD, y, "% Busy")
			WriteAtS(COL_STORAGE, y, "Queue Sz")
			WriteAtS(COL_STORAGE+9, y, "Queried")
			WriteAtS(COL_STORAGE+17, y, "Mutation")
			WriteAtS(COL_STORAGE+25, y, "  Stored")
			WriteAtS(COL_DATAVERSION, y, "Data/Dura. Lag")
			WriteAtS(COL_KVSTORE, y, "    Used")
		} else if roleId == LogRoleMetrics {
			WriteAtS(COL_HDD, y, "% Busy")
			WriteAtS(COL_STORAGE, y, "Queue Sz")
			WriteAtS(COL_STORAGE+9, y, "  Input")
			WriteAtS(COL_STORAGE+17, y, "Durable")
			WriteAtS(COL_STORAGE+25, y, "    Used")
			maxLogTransaction = maxDataVersion(kv)
			WriteAtS(COL_DATAVERSION, y, "        Delta")
			WriteAtS(COL_KVSTORE, y, "    Used")
		}

		y++

		prevHost := ""
		sort.SliceStable(kv, func(i, j int) bool {
			return kv[i].Process.Address < kv[j].Process.Address
		})
		for _, item := range kv {
			role := item.Role
			proc := item.Process
			//machineID := v.MachineId
			if y < ScreenHeight {
				SetColor("DarkGray")
				WriteAtS(COL0, y, strings.ReplaceAll(" _______________:_____ | ________ ________ | _____% __________ | ________ | ", "_", " "))
				if roleId == "storage" {
					WriteAtS(COL_HDD, y, strings.ReplaceAll("_____% __________ | ________ _______ _______ ________ | _____s _____s | ________", "_", " "))
				} else if roleId == "log" {
					WriteAtS(COL_HDD, y, strings.ReplaceAll("_____% __________ | ________ _______ _______ ________ | _____________ | ________", "_", " "))
				}

				host := GetHostFromAddress(proc.Address)
				if host != prevHost {
					SetColor("Gray")
					WriteAt(COL0+1, y, "%15s", host)
				}
				if proc.Excluded {
					SetColor("DarkRed")
				} else {
					SetColor("Gray")
				}
				WriteAt(COL0+1+16, y, GetPortFromAddress(proc.Address))
				prevHost = host

				SetColor(MapMegabitsToColor(proc.Network.MegabitsReceived.Hz))
				//WriteAt( COL_NET, y, "%7.1f", Nice(proc.Network.MegabitsReceived.Hz, "-", 0.05, "~"))
				WriteAt(COL_NET, y, "%7s", Nice(proc.Network.MegabitsReceived.Hz, "-", 0.05, "~"))
				SetColor(MapMegabitsToColor(proc.Network.MegabitsSent.Hz))
				//WriteAt( COL2, y, "%7.1f", Nice(proc.Network.MegabitsSent.Hz, "-", 0.05, "~"))
				WriteAt(COL2, y, "%7s", Nice(proc.Network.MegabitsSent.Hz, "-", 0.05, "~"))

				SetColor("Gray")
				WriteAt(COL_CPU, y, "%5.1f", proc.Cpu.UsageCores*100)
				if proc.Cpu.UsageCores >= 0.95 {
					SetColor("DarkRed")
				} else if proc.Cpu.UsageCores >= 0.75 {
					SetColor("DarkYellow")
				} else {
					SetColor("DarkGreen")
				}
				WriteAt(COL_CPU+7, y, "%-10s", BarGraph(proc.Cpu.UsageCores, 1, CPU_BARSZ, '|', ":", "."))

				memoryUsed := proc.Memory.UsedBytes

				if memoryUsed >= int64(0.75*float64(proc.Memory.LimitBytes)) {
					SetColor("White")
				} else if memoryUsed >= GIBIBYTE {
					SetColor("Gray")
				} else {
					SetColor("DarkGray")
				}
				WriteAt(COL_MEMORY, y, "%8s", FriendlyBytes(memoryUsed))

				if role.Role == StorageRoleMetrics {
					storage := role
					// Queue Size
					SetColor(MapQueueSizeToColor(float64(storage.InputBytes.Counter - storage.DurableBytes.Counter)))
					WriteAt(COL_STORAGE, y, "%8s", FriendlyBytes(int64(storage.InputBytes.Counter-storage.DurableBytes.Counter)))

					// Bytes Queried
					SetColor(MapDiskOpsToColor(storage.BytesQueried.Hz))
					//WriteAt( COL_STORAGE+9, y, "%7.2f", Nice(MegaBytes(int64(storage.BytesQueried.Hz)), "-", 0.005, "~"))
					WriteAt(COL_STORAGE+9, y, "%7s", Nice(MegaBytes(int64(storage.BytesQueried.Hz)), "-", 0.005, "~"))

					// Mutation Bytes
					SetColor(MapDiskOpsToColor(storage.MutationBytes.Hz))
					//WriteAt( COL_STORAGE+17, y, "%7.2f", Nice(MegaBytes(int64(storage.MutationBytes.Hz)), "-", 0.005, "~"))
					WriteAt(COL_STORAGE+17, y, "%7s", Nice(MegaBytes(int64(storage.MutationBytes.Hz)), "-", 0.005, "~"))

					SetColor("Gray")
					WriteAt(COL_STORAGE+25, y, "%8s", FriendlyBytes(storage.StoredBytes))

					dataLag := storage.DataLag.Seconds
					SetColor(MapDataLagToColor(dataLag))
					//WriteAt( COL_DATAVERSION, y, "%5.1f", Nice(dataLag, "-", 0.005, "~"))
					WriteAt(COL_DATAVERSION, y, "%5s", Nice(dataLag, "-", 0.005, "~"))
					durLag := storage.DurabilityLag.Seconds
					SetColor(MapDurLagToColor(durLag))
					WriteAt(COL_DATAVERSION+7, y, "%5.1f", durLag)

					WriteAt(COL_KVSTORE, y, "%8s", FriendlyBytes(storage.KvstoreUsedBytes))
				} else if role.Role == LogRoleMetrics {
					log := role
					// Queue Size
					SetColor(MapQueueSizeToColor(float64(log.InputBytes.Counter - log.DurableBytes.Counter)))
					WriteAt(COL_STORAGE, y, "%8s", FriendlyBytes(int64(log.InputBytes.Counter-log.DurableBytes.Counter)))

					// Durable Bytes
					SetColor(MapDiskOpsToColor(log.InputBytes.Hz))
					//WriteAt( COL_STORAGE+9, y, "%7.1f", Nice(MegaBytes(int64(log.InputBytes.Hz)), "-", 0.005, "~"))
					WriteAt(COL_STORAGE+9, y, "%7s", Nice(MegaBytes(int64(log.InputBytes.Hz)), "-", 0.005, "~"))

					SetColor(MapDiskOpsToColor(log.InputBytes.Hz))
					//WriteAt( COL_STORAGE+17, y, "%7.1f", Nice(MegaBytes(int64(log.DurableBytes.Hz)), "-", 0.005, "~"))
					WriteAt(COL_STORAGE+17, y, "%7s", Nice(MegaBytes(int64(log.DurableBytes.Hz)), "-", 0.005, "~"))

					SetColor("Gray")
					WriteAt(COL_STORAGE+25, y, "%8s", FriendlyBytes(log.QueueDiskUsedBytes))

					delta := log.DataVersion - maxLogTransaction
					if delta >= -500_000 {
						SetColor("DarkGray")
					} else if delta >= -1_000_000 {
						SetColor("Gray")
					} else if delta >= -2_000_000 {
						SetColor("White")
					} else if delta >= -5_000_000 {
						SetColor("Cyan")
					} else if delta >= -10_000_000 {
						SetColor("DarkYellow")
					} else {
						SetColor("DarkRed")
					}
					//WriteAt( COL_DATAVERSION, y, "%13.0f", Nice(float64(delta), "-", 0.005, "~"))
					WriteAt(COL_DATAVERSION, y, "%13s", Nice(float64(delta), "-", 0.005, "~"))

					WriteAt(COL_KVSTORE, y, "%8s", FriendlyBytes(log.KvstoreUsedBytes))
				}

				if hasDisk {
					SetColor("Gray")
					WriteAt(COL_HDD, y, "%5.1f", proc.Disk.Busy*100)
					if proc.Disk.Busy == 0.0 {
						SetColor("DarkGray")
					} else if proc.Disk.Busy >= 0.95 {
						SetColor("DarkRed")
					} else {
						SetColor("DarkGreen")
					}
					WriteAt(COL_HDD+7, y, "%-5s", BarGraph(proc.Disk.Busy, 1, HDD_BARSZ, '|', ":", "."))
				}

			}
			y++
		}
		WriteAt(COL0, y, emptyLine)
		y++
	}

	min := func(x, y int) int {
		if x < y {
			return x
		}
		return y
	}
	y = min(y, ScreenHeight)
	for ; y < LastProcessYMax; y++ {
		WriteAt(COL0, y, emptyLine)
	}
	LastProcessYMax = y
}

func MapDataLagToColor(dataLag float64) string {
	if dataLag < 0.5 {
		return "DarkGray"
	} else if dataLag < 1 {
		return "Gray"
	} else if dataLag < 2 {
		return "White"
	} else if dataLag < 6 {
		return "Cyan"
	} else if dataLag < 11 {
		return "DarkYellow"
	} else {
		return "DarkRed"
	}
}

func MapDurLagToColor(durLag float64) string {
	if durLag < 6 {
		return "DarkGray"
	} else if durLag < 8 {
		return "Gray"
	} else if durLag < 11 {
		return "White"
	} else if durLag < 16 {
		return "Cyan"
	} else if durLag < 26 {
		return "DarkYellow"
	} else {
		return "DarkRed"
	}
}

func GetHostFromAddress(address string) string {
	p := strings.Index(address, ":")
	if p < 0 {
		return address
	}
	return address[:p]
}

func GetPortFromAddress(address string) string {
	p := strings.Index(address, ":")
	if p < 0 {
		return ""
	}
	return address[p+1:]
}

type RoleMap struct {
	Master            bool
	ClusterController bool
	Proxy             bool
	Log               bool
	Storage           bool
	Resolver          bool
	RateKeeper        bool
	DataDistributor   bool
	CommitProxy       bool
	GrvProxy          bool
	Other             bool
}

func (r *RoleMap) Add(role string) {
	switch role {
	case "master":
		r.Master = true
	case "cluster_controller":
		r.ClusterController = true
	case "proxy":
		r.Proxy = true
	case "commit_proxy":
		r.Proxy = true
		r.CommitProxy = true
	case "grv_proxy":
		r.Proxy = true
		r.GrvProxy = true
	case "log":
		r.Log = true
	case "storage":
		r.Storage = true
	case "resolver":
		r.Resolver = true
	case "ratekeeper":
		r.Other = true
		r.RateKeeper = true
	case "data_distributor":
		r.Other = true
		r.DataDistributor = true
	default:
		r.Other = true
	}
}

func (r *RoleMap) Reset() {
	r.Master = false
	r.ClusterController = false
	r.Proxy = false
	r.Log = false
	r.Storage = false
	r.Resolver = false
	r.RateKeeper = false
	r.DataDistributor = false
	r.CommitProxy = false
	r.GrvProxy = false
	r.Other = false
}

func (r *RoleMap) String() string {
	var sb strings.Builder
	if r.Master {
		sb.WriteString("M")
	} else {
		sb.WriteString("-")
	}
	if r.ClusterController {
		sb.WriteString("C")
	} else {
		sb.WriteString("-")
	}
	if r.Proxy {
		sb.WriteString("P")
	} else {
		sb.WriteString("-")
	}
	if r.CommitProxy {
		sb.WriteString("c")
	} else {
		sb.WriteString("-")
	}
	if r.GrvProxy {
		sb.WriteString("g")
	} else {
		sb.WriteString("-")
	}
	if r.Log {
		sb.WriteString("L")
	} else {
		sb.WriteString("-")
	}
	if r.Storage {
		sb.WriteString("S")
	} else {
		sb.WriteString("-")
	}
	if r.Resolver {
		sb.WriteString("R")
	} else {
		sb.WriteString("-")
	}
	if r.Other {
		sb.WriteString("O")
	} else {
		sb.WriteString("-")
	}
	if r.RateKeeper {
		sb.WriteString("r")
	} else {
		sb.WriteString("-")
	}
	if r.DataDistributor {
		sb.WriteString("d")
	} else {
		sb.WriteString("-")
	}
	return sb.String()
}

var LastProcessYMax = 0

const (
	KIBIBYTE = 1024.0
	MEBIBYTE = 1024.0 * 1024.0
	GIBIBYTE = 1024.0 * 1024.0 * 1024.0
	TEBIBYTE = 1024.0 * 1024.0 * 1024.0 * 1024.0
)

func KiloBytes(x int64) float64 {
	return float64(x) / KIBIBYTE
}

func GigaBytes(x int64) float64 {
	return float64(x) / GIBIBYTE
}

func MegaBytes(x int64) float64 {
	return float64(x) / MEBIBYTE
}

func MegaBytesFloat(x float64) float64 {
	return x / MEBIBYTE
}

func TeraBytes(x float64) float64 {
	return x / TEBIBYTE
}

func FriendlyBytes(x int64) string {
	if x == 0 {
		return "-"
	}
	if x < 900*KIBIBYTE {
		return fmt.Sprintf("%.1f KB", KiloBytes(x))
	}
	if x < 900*MEBIBYTE {
		return fmt.Sprintf("%.1f MB", MegaBytes(x))
	}
	if x < 900*GIBIBYTE {
		return fmt.Sprintf("%.1f GB", GigaBytes(x))
	}
	return fmt.Sprintf("%.1f TB", TeraBytes(float64(x)))
}

func MapDiskOpsToColor(ops float64) string {
	switch {
	case ops < 500*KIBIBYTE:
		return "DarkGray"
	case ops < 5*MEBIBYTE:
		return "Gray"
	case ops < 25*MEBIBYTE:
		return "White"
	case ops > 100*MEBIBYTE:
		return "Red"
	default:
		return "Cyan"
	}
}

func MapQueueSizeToColor(value float64) string {
	switch {
	case value < 10*MEBIBYTE:
		return "DarkGray"
	case value < 100*MEBIBYTE:
		return "Gray"
	case value < 1*GIBIBYTE:
		return "White"
	case value < 5*GIBIBYTE:
		return "Cyan"
	case value < 10*GIBIBYTE:
		return "DarkYellow"
	default:
		return "DarkRed"
	}
}

func MapMegabitsToColor(megaBits float64) string {
	megaBits /= 8
	switch {
	case megaBits < 0.1:
		return "DarkGray"
	case megaBits >= 1000:
		return "Red"
	case megaBits >= 100:
		return "DarkRed"
	case megaBits >= 50:
		return "DarkYellow"
	case megaBits >= 10:
		return "Cyan"
	case megaBits >= 1:
		return "White"
	default:
		return "Gray"
	}
}

func MapConnectionsToColor(connections int64) string {
	switch {
	case connections < 10:
		return "DarkGray"
	case connections < 50:
		return "Gray"
	case connections < 100:
		return "White"
	case connections < 250:
		return "Cyan"
	case connections < 500:
		return "DarkYellow"
	default:
		return "DarkRed"
	}
}

func MapMemoryToColor(value int64) string {
	switch {
	case value < GIBIBYTE:
		return "DarkGray"
	case value <= 3*GIBIBYTE:
		return "Gray"
	case value <= 5*GIBIBYTE:
		return "White"
	case value <= 7*GIBIBYTE:
		return "DarkYellow"
	default:
		return "DarkRed"
	}
}

func Nice(value float64, zero string, epsilon float64, small string) string {
	if value == 0 {
		return zero
	}
	if value < epsilon {
		if small != "" {
			return small
		}
		return "."
	}
	return fmt.Sprintf("%.2f", value)
}

func TimeSpanInSecondsWithRounding(seconds float64) time.Duration {
	roundedSeconds := math.Round(seconds)
	return time.Duration(roundedSeconds) * time.Second
}
