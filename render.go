package main

import (
	"fmt"
	"sort"
	"strings"

	"cli-web-monitor/stats"
)

func renderTable(statsMap map[string]*stats.Stats) {
	// prepare data
	var urls []string
	longest := 0

	for url := range statsMap {
		urls = append(urls, url)

		if len(url) > longest {
			longest = len(url)
		}
	}

	sort.Strings(urls)

	fmt.Print("\033[2J\033[H") // clear console
	// header
	halfOfUrlPart := strings.Repeat(" ", (longest+1)/2)
	fmt.Printf("┌%s┐\n", strings.Repeat("-", longest+60))
	fmt.Printf("| %sURL%s Duration (ms)            Size (≈KiB)             OK   |\n", halfOfUrlPart, halfOfUrlPart)
	fmt.Printf("├%s┤\n", strings.Repeat("-", longest+60))

	// line of each url - [Min / Avg / Max]
	format := fmt.Sprintf("| %%-%ds  %%4d / %%4d / %%4d      %%4d / %%4d / %%4d       %%3d/%%3d |\n", longest)

	for _, url := range urls {
		stat := statsMap[url]

		report := stat.Get()

		if stat == nil || report == nil {
			continue
		}

		fmt.Printf(format,
			url,
			report.MinDuration.Milliseconds(), report.AvgDuration.Milliseconds(), report.MaxDuration.Milliseconds(),
			report.MinSize, report.AvgSize, report.MaxSize,
			report.Success, report.Total,
		)
	}

	// footer
	fmt.Printf("└%s┘\n", strings.Repeat("-", longest+60))
}
