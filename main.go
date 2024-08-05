package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/thrcorrea/dora/internal/dora"
	"github.com/thrcorrea/dora/pkg/github"
	"github.com/thrcorrea/dora/pkg/shortcut"
)

func startOfWeek(t time.Time, numberWeeks int) time.Time {
	weekday := t.Weekday()
	offset := int(time.Monday - weekday)
	if offset > 0 {
		offset = -6
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, offset*numberWeeks)
}

func endOfWeek(t time.Time, numberWeeks int) time.Time {
	start := startOfWeek(t, numberWeeks)
	return start.AddDate(0, 0, 6*numberWeeks).Add(time.Hour*23 + time.Minute*59 + time.Second*59)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	shortcutToken := os.Getenv("SHORTCUT_TOKEN")
	githubClient := github.NewGithubClient(githubToken)
	shortcutClient := shortcut.NewShortcutClient(shortcutToken)
	doraService := dora.NewDoraService(githubClient, shortcutClient)

	today := time.Now().Add(-time.Hour * 24 * 3)
	startPeriod := startOfWeek(today, 1)
	endPeriod := endOfWeek(today, 1)

	// startPeriod := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	// endPeriod := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	period := []time.Time{startPeriod, endPeriod}
	cyclePeriod := []time.Time{startPeriod.Add(-time.Hour * 24 * 23), endPeriod}
	fmt.Printf("Period: %v\n", period)
	fmt.Printf("CyclePeriod: %v\n", cyclePeriod)
	fmt.Println("")

	prCount, err := doraService.GetDeploymentMetric(period)
	if err != nil {
		panic(err)
	}

	cfrMetric, err := doraService.GetCFRMetric(period, prCount)
	if err != nil {
		panic(err)
	}

	cycleTime, err := doraService.GetCycleTime(cyclePeriod)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Cycle Time: %s\n", cycleTime)

	mttrTime, err := doraService.GetMTTRMetric(cyclePeriod)
	if err != nil {
		panic(err)
	}
	fmt.Printf("MTTR: %s\n", mttrTime)

	fmt.Printf("Total PRs: %d\n", prCount)
	if prCount > 0 {
		cfr := (float32(cfrMetric) / float32(prCount)) * 100
		fmt.Printf("CFR: %f\n", cfr)
	} else {
		fmt.Printf("CFR: %d\n", cfrMetric*100)
	}
}
