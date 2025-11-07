package ui

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"
	"time"

	"github.com/0xADE/ade-ctld/client/exe"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// UI represents the main UI state
type UI struct {
	client       *exe.Client
	theme        *material.Theme
	window       *app.Window
	filterEditor widget.Editor
	list         widget.List
	query        string
	applications []exe.Application
	selectedIdx  int
	status       string
	statusMx     sync.RWMutex
	initialized  bool

	// Filter debouncing
	queryInput   chan string
	queryResults chan []exe.Application
	stopFilter   chan struct{}

	// Key repeat handling
	keyRepeatActive bool
	keyRepeatName   key.Name
	keyRepeatStart  time.Time
}

// New creates a new UI instance
func New(c *exe.Client) *UI {
	ui := &UI{
		client:       c,
		list:         widget.List{List: layout.List{Axis: layout.Vertical}},
		queryInput:   make(chan string, 64),
		queryResults: make(chan []exe.Application, 1),
		stopFilter:   make(chan struct{}),
		applications: []exe.Application{},
	}

	ui.filterEditor.SingleLine = true
	ui.filterEditor.Submit = true

	// Load initial list
	go ui.updateApplications("")

	ui.startFilterWorker()
	return ui
}

// startFilterWorker starts a goroutine that debounces filter queries
func (ui *UI) startFilterWorker() {
	go func() {
		var timer *time.Timer
		var latestQuery string
		debounceDelay := 300 * time.Millisecond

		for {
			select {
			case query := <-ui.queryInput:
				latestQuery = query
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(debounceDelay, func() {
					ui.updateApplications(latestQuery)
				})
			case <-ui.stopFilter:
				if timer != nil {
					timer.Stop()
				}
				return
			}
		}
	}()
}

// updateApplications updates the application list based on filter query
func (ui *UI) updateApplications(query string) {
	var apps []exe.Application
	var err error

	if query == "" {
		err = ui.client.ResetFilters()
	} else {
		err = ui.client.SetFilterName(query)
	}
	if err != nil {
		ui.setStatus(fmt.Sprintf("Filter error: %v", err))
		return
	}

	apps, err = ui.client.List()
	if err != nil {
		ui.setStatus(fmt.Sprintf("List error: %v", err))
		return
	}

	select {
	case ui.queryResults <- apps:
		if ui.window != nil {
			ui.window.Invalidate()
		}
	default:
		// Skip if channel full
	}
}

func (ui *UI) setStatus(msg string) {
	fmt.Println(msg) // NOTE on debug
	ui.statusMx.Lock()
	ui.status = msg
	ui.statusMx.Unlock()
	if ui.window != nil {
		ui.window.Invalidate()
	}
}

func (ui *UI) moveSelectionUp() {
	if ui.selectedIdx > 0 {
		ui.selectedIdx--
		if ui.list.Position.First > ui.selectedIdx {
			ui.list.Position.First = ui.selectedIdx
		}
		if ui.window != nil {
			ui.window.Invalidate()
		}
	}
}

func (ui *UI) moveSelectionDown() {
	if ui.selectedIdx < len(ui.applications)-1 {
		ui.selectedIdx++
		if ui.list.Position.Count > 0 && ui.list.Position.First+ui.list.Position.Count <= ui.selectedIdx {
			ui.list.Position.First = ui.selectedIdx - ui.list.Position.Count + 1
		}
		if ui.window != nil {
			ui.window.Invalidate()
		}
	}
}

func (ui *UI) runSelected() {
	if ui.selectedIdx >= len(ui.applications) {
		ui.setStatus("No application selected")
		return
	}

	app := ui.applications[ui.selectedIdx]
	ui.setStatus(fmt.Sprintf("Running: %s", app.Name))

	// Run client.Run() asynchronously to avoid blocking UI event loop
	go func() {
		err := ui.client.Run(app.ID)
		if err != nil {
			ui.setStatus(fmt.Sprintf("Run error: %v", err))
		} else {
			ui.setStatus(fmt.Sprintf("Launched: %s", app.Name))
		}
	}()
}

// Run starts the UI application loop
func (ui *UI) Run() error {
	ui.window = new(app.Window)
	ui.window.Option(app.Title("xopen"))
	ui.window.Option(app.Size(unit.Dp(800), unit.Dp(600)))

	go func() {
		if err := ui.loop(); err != nil {
			panic(err)
		}
	}()

	app.Main()
	return nil
}

func (ui *UI) loop() error {
	th := material.NewTheme()
	ui.theme = th

	var ops op.Ops
	for {
		switch e := ui.window.Event().(type) {
		case app.DestroyEvent:
			close(ui.stopFilter)
			return e.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			// Process filter results
			select {
			case apps := <-ui.queryResults:
				ui.applications = apps
				if ui.selectedIdx >= len(ui.applications) {
					ui.selectedIdx = 0
				}
			default:
				// No results ready
			}

			if !ui.initialized {
				gtx.Execute(key.FocusCmd{Tag: &ui.filterEditor})
				ui.initialized = true
			}

			// Focus filter editor
			gtx.Execute(key.FocusCmd{Tag: &ui.filterEditor})

			// Register global key listener
			area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			event.Op(gtx.Ops, ui.window)

			// Handle keyboard shortcuts
			var filters []event.Filter
			filters = append(filters,
				key.Filter{Name: key.NameEscape},
				key.Filter{Name: key.NameUpArrow},
				key.Filter{Name: key.NameDownArrow},
				key.Filter{Name: key.NameReturn},
			)

			for {
				ev, ok := gtx.Event(filters...)
				if !ok {
					break
				}
				if kev, ok := ev.(key.Event); ok {
					switch kev.State {
					case key.Press:
						switch kev.Name {
						case key.NameEscape:
							os.Exit(0)
						case key.NameUpArrow:
							ui.moveSelectionUp()
							ui.keyRepeatActive = true
							ui.keyRepeatName = key.NameUpArrow
							ui.keyRepeatStart = time.Now()
						case key.NameDownArrow:
							ui.moveSelectionDown()
							ui.keyRepeatActive = true
							ui.keyRepeatName = key.NameDownArrow
							ui.keyRepeatStart = time.Now()
						case key.NameReturn:
							ui.runSelected()
						}
					case key.Release:
						// Stop key repeat when key is released
						if ui.keyRepeatActive && kev.Name == ui.keyRepeatName {
							ui.keyRepeatActive = false
						}
					}
				}
			}

			// Handle key repeat for arrow keys
			if ui.keyRepeatActive {
				elapsed := time.Since(ui.keyRepeatStart)
				initialDelay := 200 * time.Millisecond
				repeatInterval := 30 * time.Millisecond

				if elapsed > initialDelay {
					repeatCount := int((elapsed - initialDelay) / repeatInterval)
					lastRepeatTime := initialDelay + time.Duration(repeatCount)*repeatInterval
					nextRepeatTime := lastRepeatTime + repeatInterval

					if elapsed >= nextRepeatTime {
						switch ui.keyRepeatName {
						case key.NameUpArrow:
							ui.moveSelectionUp()
						case key.NameDownArrow:
							ui.moveSelectionDown()
						}
						ui.keyRepeatStart = ui.keyRepeatStart.Add(nextRepeatTime)
					}

					// Schedule next frame
					gtx.Execute(op.InvalidateCmd{})
				} else {
					// Wait for initial delay
					gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(initialDelay - elapsed)})
				}
			}

			// Handle filter editor events
			for {
				ev, ok := ui.filterEditor.Update(gtx)
				if !ok {
					break
				}
				switch ev.(type) {
				case widget.ChangeEvent:
					ui.query = ui.filterEditor.Text()
					select {
					case ui.queryInput <- ui.query:
					default:
						// Channel full, skip this update
					}
				case widget.SubmitEvent:
					ui.runSelected()
				}
			}

			ui.layout(gtx)
			area.Pop()
			e.Frame(gtx.Ops)
		}
	}
}

func (ui *UI) layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				editor := material.Editor(ui.theme, &ui.filterEditor, "Filter by name...")
				editor.TextSize = unit.Sp(20)
				return editor.Layout(gtx)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, ui.layoutApplicationList)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ui.statusMx.RLock()
				status := ui.status
				ui.statusMx.RUnlock()
				label := material.Body2(ui.theme, status)
				label.Color = color.NRGBA{R: 170, G: 170, B: 170, A: 255}
				return label.Layout(gtx)
			})
		}),
	)
}

func (ui *UI) layoutApplicationList(gtx layout.Context) layout.Dimensions {
	return material.List(ui.theme, &ui.list).Layout(gtx, len(ui.applications), func(gtx layout.Context, index int) layout.Dimensions {
		isSelected := index == ui.selectedIdx

		// Create clickable for this item
		clickable := widget.Clickable{}

		// First render the content to get its height
		macro := op.Record(gtx.Ops)
		dims := layout.UniformInset(unit.Dp(8)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(ui.theme, ui.applications[index].Name)
			label.TextSize = unit.Sp(18)
			return label.Layout(gtx)
		})
		call := macro.Stop()

		// Draw background if selected, using full width
		if isSelected {
			selectionColor := color.NRGBA{R: 100, G: 150, B: 200, A: 100}
			bgRect := image.Pt(gtx.Constraints.Max.X, dims.Size.Y)
			defer clip.Rect{Max: bgRect}.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: selectionColor}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
		}

		// Draw the content on top
		call.Add(gtx.Ops)

		// Handle pointer clicks
		clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return dims
		})
		for clickable.Clicked(gtx) {
			ui.selectedIdx = index
			ui.runSelected()
		}

		return dims
	})
}
