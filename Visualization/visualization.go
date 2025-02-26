package visualization

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// GenerateVisualization creates an HTML page with three charts:
// 1. Performance Chart (Duration vs Memory)
// 2. Cost Chart (Cost vs Memory)
// 3. Combined Chart with dual y-axes (Duration and Cost vs Memory)
func GenerateVisualization(memory []int, durations []float64, costs []float64) error {
	if len(memory) == 0 || len(memory) != len(durations) || len(memory) != len(costs) {
		return fmt.Errorf("invalid input slices")
	}

	// Prepare x-axis values (memory as strings)
	var xAxis []string
	var durationData []opts.LineData
	var costData []opts.LineData
	for i, mem := range memory {
		xAxis = append(xAxis, strconv.Itoa(mem))
		durationData = append(durationData, opts.LineData{Value: durations[i]})
		costData = append(costData, opts.LineData{Value: costs[i]})
	}

	// Performance Chart: Duration vs Memory
	performanceChart := charts.NewLine()
	performanceChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Lambda Performance (Duration vs Memory)"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Memory (MB)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Duration (ms)"}),
	)
	performanceChart.SetXAxis(xAxis).
		AddSeries("Duration", durationData)

	// Cost Chart: Cost vs Memory
	costChart := charts.NewLine()
	costChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Lambda Cost (Cost vs Memory)"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Memory (MB)"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Cost ($)"}),
	)
	costChart.SetXAxis(xAxis).
		AddSeries("Cost", costData)

	// Combined Chart: Both Duration and Cost on a dual-axis chart.
	combinedChart := charts.NewLine()
	combinedChart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Balanced Sweet Spot (Cost & Performance)"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Memory (MB)"}),
		// Primary Y-axis for Duration:
		charts.WithYAxisOpts(opts.YAxis{
			Name: "Duration (ms)",
		}),
	)
	// Add a secondary y-axis for Cost.
	combinedChart.ExtendYAxis(opts.YAxis{
		Name:     "Cost ($)",
		Position: "right",
	})

	// Add Duration series on primary y-axis (index 0)
	combinedChart.SetXAxis(xAxis).
		AddSeries("Duration", durationData, charts.WithLineChartOpts(opts.LineChart{YAxisIndex: 0}))
	// Add Cost series on secondary y-axis (index 1)
	combinedChart.AddSeries("Cost", costData, charts.WithLineChartOpts(opts.LineChart{YAxisIndex: 1}))

	// Combine all charts into a single page.
	page := components.NewPage()
	page.AddCharts(performanceChart, costChart, combinedChart)

	// Create and save the HTML file.
	f, err := os.Create("visualization.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return page.Render(f)
}
