package helper

import (
	"context"
	"io/ioutil"
	"os"
	"shareway/schemas"

	"github.com/chromedp/chromedp"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func RenderChartToPNG(htmlFilePath, pngFilePath string) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Đọc nội dung file HTML
	htmlContent, err := ioutil.ReadFile(htmlFilePath)
	if err != nil {
		return err
	}

	var buf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate("data:text/html,"+string(htmlContent)),
		chromedp.WaitVisible("body"),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return err
	}

	// Ghi dữ liệu hình ảnh vào file PNG
	return ioutil.WriteFile(pngFilePath, buf, 0644)
}

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

	// Render biểu đồ ra file HTML tạm thời
	htmlFilePath := "temp_user_growth_chart.html"
	f, err := os.Create(htmlFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(htmlFilePath) // Xóa file HTML sau khi sử dụng

	err = line.Render(f)
	if err != nil {
		return err
	}

	// Chuyển đổi file HTML thành PNG
	return RenderChartToPNG(htmlFilePath, filePath)
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

	// Render biểu đồ ra file HTML tạm thời
	htmlFilePath := "temp_transaction_chart.html"
	f, err := os.Create(htmlFilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(htmlFilePath) // Xóa file HTML sau khi sử dụng

	err = bar.Render(f)
	if err != nil {
		return err
	}

	// Chuyển đổi file HTML thành PNG
	return RenderChartToPNG(htmlFilePath, filePath)
}
