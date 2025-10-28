package cmd

import (
	"fmt"
	"time"

	"github.com/raucheacho/rosia-cli/internal/telemetry"
	"github.com/raucheacho/rosia-cli/pkg/logger"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Display telemetry statistics",
	Long:  `Display statistics about scan and clean operations including total scans, total space cleaned, and average sizes by profile type.`,
	RunE:  runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	// Get the stats file path
	statsPath, err := telemetry.GetDefaultStatsPath()
	if err != nil {
		logger.Error("Failed to get stats path: %v", err)
		return fmt.Errorf("failed to get stats path: %w", err)
	}

	// Create telemetry store
	store, err := telemetry.NewFileStore(statsPath)
	if err != nil {
		logger.Error("Failed to initialize telemetry store: %v", err)
		return fmt.Errorf("failed to initialize telemetry store: %w", err)
	}

	// Get statistics
	stats, err := store.GetStats()
	if err != nil {
		logger.Error("Failed to get statistics: %v", err)
		return fmt.Errorf("failed to get statistics: %w", err)
	}

	// Display statistics
	displayStats(stats)

	return nil
}

func displayStats(stats *telemetry.Stats) {
	fmt.Println("📊 Rosia Statistics")
	fmt.Println("==================")
	fmt.Println()

	// Total scans
	fmt.Printf("Total Scans:        %d\n", stats.TotalScans)

	// Total cleaned with human-readable format
	fmt.Printf("Total Cleaned:      %s\n", formatSize(stats.TotalCleaned))

	// Last scan timestamp
	if !stats.LastScan.IsZero() {
		fmt.Printf("Last Scan:          %s\n", formatTimestamp(stats.LastScan))
	} else {
		fmt.Printf("Last Scan:          Never\n")
	}

	// Average sizes by type
	if len(stats.AverageSizeByType) > 0 {
		fmt.Println()
		fmt.Println("Average Size by Profile:")
		for profileName, avgSize := range stats.AverageSizeByType {
			fmt.Printf("  %-20s %s\n", profileName+":", formatSize(avgSize))
		}
	}

	fmt.Println()
}

// formatTimestamp formats a timestamp in a human-readable way
func formatTimestamp(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}
