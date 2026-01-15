package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/steenfuentes/pomo/engine"
	"github.com/steenfuentes/pomo/ui"
)

var (
	workMinutes       int
	shortBreakMinutes int
	longBreakMinutes  int
	longBreakEvery    int
	cycles            int
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a pomodoro session",
	Long: `Start a pomodoro session with configurable work and break durations.

Examples:
  pomo start                           # Default: 50min work, 10min short, 30min long every 4
  pomo start -p 25 -s 5 -l 15          # Classic pomodoro: 25min work, 5min short, 15min long
  pomo start -e 0                      # Disable long breaks
  pomo start -c 4                      # Run exactly 4 work cycles`,
	Run: runStart,
}

func init() {
	startCmd.Flags().IntVarP(&workMinutes, "pomodoro", "p", 50, "Work duration in minutes")
	startCmd.Flags().IntVarP(&shortBreakMinutes, "short", "s", 10, "Short break duration in minutes")
	startCmd.Flags().IntVarP(&longBreakMinutes, "long", "l", 30, "Long break duration in minutes")
	startCmd.Flags().IntVarP(&longBreakEvery, "long-every", "e", 4, "Long break every N work cycles (0 = no long breaks)")
	startCmd.Flags().IntVarP(&cycles, "cycles", "c", 0, "Total work cycles (0 = infinite)")

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) {
	cfg := engine.Config{
		WorkDuration:       time.Duration(workMinutes) * time.Minute,
		ShortBreakDuration: time.Duration(shortBreakMinutes) * time.Minute,
		LongBreakDuration:  time.Duration(longBreakMinutes) * time.Minute,
		LongBreakEvery:     longBreakEvery,
		TotalCycles:        cycles,
	}

	fmt.Printf("Starting pomodoro: %dm work, %dm short break", workMinutes, shortBreakMinutes)
	if longBreakEvery > 0 {
		fmt.Printf(", %dm long break every %d cycles", longBreakMinutes, longBreakEvery)
	}
	if cycles > 0 {
		fmt.Printf(" (%d cycles)", cycles)
	}
	fmt.Println()
	fmt.Println()

	timer := engine.NewTimer(cfg)
	events := make(chan engine.TimerEvent)

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, stopping...")
		cancel()
	}()

	progress := ui.NewProgress(timer.Session().TotalPhases(), nil)

	errChan := make(chan error, 1)
	go func() {
		errChan <- timer.Run(ctx, events)
	}()

	for event := range events {
		progress.Update(event)
	}

	progress.Wait()

	if err := <-errChan; err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Session complete!")
}
