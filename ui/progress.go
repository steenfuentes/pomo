// Package ui provides terminal display components for the pomodoro timer.
package ui

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/steenfuentes/pomo/engine"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

var (
	workColor      = color.New(color.FgRed)
	shortColor     = color.New(color.FgCyan)
	longColor      = color.New(color.FgGreen)
	overallColor   = color.New(color.FgWhite)
	dimColor       = color.New(color.Faint)
)

// Progress displays pomodoro progress using mpb.
type Progress struct {
	container   *mpb.Progress
	phaseBar    *mpb.Bar
	overallBar  *mpb.Bar
	showOverall bool
	totalPhases int
	phaseTotal  int64
	lastPhase   engine.Phase
}

// NewProgress creates a progress display.
// If totalPhases > 0, an overall progress bar is shown.
func NewProgress(totalPhases int, output io.Writer) *Progress {
	opts := []mpb.ContainerOption{
		mpb.WithWidth(50),
		mpb.WithRefreshRate(50 * time.Millisecond),
	}
	if output != nil {
		opts = append(opts, mpb.WithOutput(output))
	}

	p := &Progress{
		container:   mpb.New(opts...),
		showOverall: totalPhases > 0,
		totalPhases: totalPhases,
		lastPhase:   engine.Phase(-1),
	}

	if p.showOverall {
		p.overallBar = p.container.New(int64(totalPhases),
			mpb.BarStyle().Lbound("[").Filler("=").Tip(">").Padding("-").Rbound("]"),
			mpb.PrependDecorators(
				decor.Name(overallColor.Sprint("  Total "), decor.WCSyncSpaceR),
			),
			mpb.AppendDecorators(
				decor.CountersNoUnit(dimColor.Sprint(" %d/%d"), decor.WCSyncSpace),
			),
			mpb.BarFillerClearOnComplete(),
		)
	}

	return p
}

// Update processes a timer event and updates the display.
func (p *Progress) Update(e engine.TimerEvent) {
	if p.phaseBar == nil || e.Phase != p.lastPhase {
		if p.phaseBar != nil {
			p.phaseBar.SetCurrent(p.phaseTotal)
			p.phaseBar.EnableTriggerComplete()
		}

		p.lastPhase = e.Phase
		p.phaseTotal = int64(e.Total / time.Millisecond)

		p.phaseBar = p.container.New(p.phaseTotal,
			barStyleForPhase(e.Phase),
			mpb.PrependDecorators(
				decor.Name(formatPhaseName(e), decor.WCSyncSpaceR),
			),
			mpb.AppendDecorators(
				decor.Any(func(s decor.Statistics) string {
					elapsed := time.Duration(s.Current) * time.Millisecond
					total := time.Duration(s.Total) * time.Millisecond
					return dimColor.Sprintf(" %s/%s", formatDuration(elapsed), formatDuration(total))
				}, decor.WCSyncSpace),
			),
			mpb.BarFillerClearOnComplete(),
		)
	}

	elapsed := int64(e.Elapsed / time.Millisecond)
	p.phaseBar.SetCurrent(elapsed)

	if e.PhaseComplete && p.showOverall && p.overallBar != nil {
		p.overallBar.Increment()
	}
}

// Wait blocks until all bars complete.
func (p *Progress) Wait() {
	if p.phaseBar != nil {
		p.phaseBar.SetCurrent(p.phaseTotal)
		p.phaseBar.EnableTriggerComplete()
	}
	p.container.Wait()
}

func barStyleForPhase(phase engine.Phase) mpb.BarFillerBuilder {
	style := mpb.BarStyle().Lbound("[").Tip(">").Padding("-").Rbound("]")

	switch phase {
	case engine.PhaseWork:
		return style.Filler(workColor.Sprint("="))
	case engine.PhaseShortBreak:
		return style.Filler(shortColor.Sprint("="))
	case engine.PhaseLongBreak:
		return style.Filler(longColor.Sprint("="))
	default:
		return style.Filler("=")
	}
}

func formatPhaseName(e engine.TimerEvent) string {
	var c *color.Color
	name := e.Phase.String()

	switch e.Phase {
	case engine.PhaseWork:
		c = workColor
	case engine.PhaseShortBreak:
		c = shortColor
	case engine.PhaseLongBreak:
		c = longColor
	default:
		c = overallColor
	}

	if e.TotalCycles > 0 {
		cycleNum := e.CycleNum
		if e.Phase != engine.PhaseWork {
			cycleNum = e.CycleNum - 1
		}
		return c.Sprintf("%s (%d/%d)", name, cycleNum, e.TotalCycles)
	}

	return c.Sprint(name)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	return fmt.Sprintf("%02d:%02d", m, s)
}
