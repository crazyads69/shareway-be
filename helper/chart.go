package helper

import (
	"os"
	"shareway/schemas"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// GenerateUserGrowthChart generates a line chart for user growth
func GenerateUserGrowthChart(userGrowth []schemas.UserGrowthData, filePath string) error {
	dates := make([]string, len(userGrowth))
	counts := make([]opts.LineData, len(userGrowth))

	for i, growth := range userGrowth {
		dates[i] = growth.Date.Format("02/01/2006")
		counts[i] = opts.LineData{Value: growth.Count}
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Tăng Trưởng Người Dùng"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Ngày"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Số lượng người dùng mới"}),
	)
	line.SetXAxis(dates).AddSeries("Người dùng mới", counts)

	// Render chart to file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return line.Render(f)
}

// GenerateTransactionChart generates a bar chart for transactions by day
func GenerateTransactionChart(transactions []schemas.TransactionDayData, filePath string) error {
	dates := make([]string, len(transactions))
	amounts := make([]opts.BarData, len(transactions))

	for i, transaction := range transactions {
		dates[i] = transaction.Date.Format("02/01/2006")
		amounts[i] = opts.BarData{Value: transaction.Transaction}
	}

	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Giao dịch Theo Ngày"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Ngày"}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Tổng giá trị (VND)"}),
	)
	bar.SetXAxis(dates).AddSeries("Giao dịch", amounts)

	// Render chart to file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return bar.Render(f)
}
