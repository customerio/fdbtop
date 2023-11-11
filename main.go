package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"log"
	"sync"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

type DisplayMode int

const (
	Metrics DisplayMode = iota
	Latency
	Processes
	Roles
	Transactions
)

type StatusEvent struct {
	when   time.Time
	status FdbStatus
}

func (s *StatusEvent) When() time.Time {
	return s.when
}

var screen tcell.Screen

func main() {
	// Different API versions may expose different runtime behaviors.
	if err := fdb.APIVersion(710); err != nil {
		log.Fatal("fdb init failed", err)
	}

	db, err := fdb.OpenDefault()
	if err != nil {
		log.Fatal("fdb init failed", err)
	}

	// Initialize screen
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	screen.EnableMouse()
	screen.EnablePaste()
	screen.Clear()

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	var (
		lap        = time.Now()
		mode       = Metrics
		repaint    = true
		status     FdbStatus
		fast       = true
		speed      = time.Second
		speedMutex sync.Mutex
	)

	setSpeed := func(s time.Duration) {
		speedMutex.Lock()
		defer speedMutex.Unlock()
		speed = s
	}
	getSpeed := func() time.Duration {
		speedMutex.Lock()
		defer speedMutex.Unlock()
		return speed
	}

	go func() {
		for {
			select {
			case <-time.After(getSpeed()):
				status, err := getMetrics(db)
				if err != nil {
					log.Printf("get metrics: %v\n", err)
				}

				err = screen.PostEvent(&StatusEvent{when: time.Now(), status: status})
				if err != nil {
					log.Printf("post event: %v\n", err)
				}
			}
		}
	}()

	// Event loop
	for {
		if repaint {
			screen.Clear()
			repaint = false
		}

		RepaintTopBar()
		if len(History) > 0 {
			UpdateTopBar(status, History[len(History)-1])
		}

		RepaintBottomBar(mode)

		switch mode {
		case Metrics:
			ShowMetricsScreen()
		case Transactions:
			ShowTransactionsScreen()
		case Latency:
			ShowLatencyScreen()
		case Processes:
			ShowProcessesScreen(status)
		case Roles:
			ShowRolesScreen(status)
		}

		// Update screen
		screen.Show()

		// Poll event
		ev := screen.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *StatusEvent:
			status = ev.status
			metric := HistoryMetric{
				Available:             status.ReadVersion > 0,
				LocalTime:             time.Now().Sub(lap),
				Timestamp:             status.Cluster.ClusterControllerTimestamp,
				ReadVersion:           status.ReadVersion,
				ReadsPerSecond:        status.Cluster.Workload.Operations.Reads.Hz,
				WritesPerSecond:       status.Cluster.Workload.Operations.Writes.Hz,
				WrittenBytesPerSecond: status.Cluster.Workload.Bytes.Written.Hz,
				TransStarted:          status.Cluster.Workload.Transactions.Started.Hz,
				TransCommitted:        status.Cluster.Workload.Transactions.Committed.Hz,
				TransConflicted:       status.Cluster.Workload.Transactions.Conflicted.Hz,
				LatencyCommit:         status.Cluster.LatencyProbe.CommitSeconds,
				LatencyRead:           status.Cluster.LatencyProbe.ReadSeconds,
				LatencyStart:          status.Cluster.LatencyProbe.TransactionStartSeconds,
			}
			History = append(History, metric)
			if len(History) > maxHistory {
				History = History[1:]
			}

		case *tcell.EventResize:
			screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' {
				return
			} else if ev.Rune() == 'c' {
				repaint = true
				lap = time.Now()
				History = nil
			} else if ev.Rune() == 'f' {
				fast = !fast
				if fast {
					setSpeed(time.Second)
				} else {
					setSpeed(time.Second * 2)
				}
			} else if ev.Rune() == 'p' {
				if mode != Processes {
					mode = Processes
					repaint = true
				}
			} else if ev.Rune() == 'm' {
				if mode != Metrics {
					mode = Metrics
					repaint = true
				}
			} else if ev.Rune() == 'l' {
				if mode != Latency {
					mode = Latency
					repaint = true
				}
			} else if ev.Rune() == 'r' {
				if mode != Roles {
					mode = Roles
					repaint = true
				}
			} else if ev.Rune() == 't' {
				if mode != Transactions {
					mode = Transactions
					repaint = true
				}
			}
		}
	}
}

var nameColors = map[string]tcell.Color{
	"DarkBlack":   tcell.ColorBlack,
	"DarkRed":     tcell.ColorDarkRed,
	"DarkGreen":   tcell.ColorDarkGreen,
	"DarkYellow":  tcell.ColorDarkGoldenrod,
	"DarkBlue":    tcell.ColorDarkBlue,
	"DarkMagenta": tcell.ColorDarkMagenta,
	"DarkCyan":    tcell.ColorDarkCyan,
	"DarkGray":    tcell.ColorDarkGray,

	"Black":   tcell.ColorBlack,
	"Red":     tcell.ColorRed,
	"Green":   tcell.ColorGreen,
	"Yellow":  tcell.ColorYellow,
	"Blue":    tcell.ColorBlue,
	"Magenta": tcell.ColorMaroon,
	"Cyan":    tcell.ColorLightCyan,
	"Gray":    tcell.ColorSilver,
	"White":   tcell.ColorWhite,
}

func MapColor(name string) tcell.Color {
	c, ok := nameColors[name]
	if !ok {
		panic(fmt.Sprint("missing color: ", name))
	}
	return c
}

var currentFg tcell.Color
var currentBg tcell.Color = tcell.ColorDefault

func SetBackground(name string) {
	currentBg = MapColor(name)
}

func SetColor(name string) {
	currentFg = MapColor(name)
}

func SetColorIf(v bool, c1 string, c2 string) {
	if v {
		SetColor(c1)
	} else {
		SetColor(c2)
	}
}

func WriteAt(x int, y int, format string, args ...interface{}) {
	currentStyle := tcell.StyleDefault.Foreground(currentFg).Background(currentBg)
	text := fmt.Sprintf(format, args...)
	for _, r := range []rune(text) {
		screen.SetContent(x, y, r, nil, currentStyle)
		x++
	}
}

func WriteAtS(x int, y int, text string) {
	currentStyle := tcell.StyleDefault.Foreground(currentFg).Background(currentBg)
	for _, r := range []rune(text) {
		screen.SetContent(x, y, r, nil, currentStyle)
		x++
	}
}
