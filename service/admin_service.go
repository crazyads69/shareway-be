package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"shareway/helper"
	"shareway/infra/bucket"
	"shareway/infra/db/migration"
	"shareway/infra/ws"
	"shareway/repository"
	"shareway/schemas"
	"shareway/util"
	"shareway/util/sanctum"
	"time"

	"github.com/google/uuid"
	"github.com/phpdave11/gofpdf"
	"github.com/xuri/excelize/v2"
)

type AdminService struct {
	repo         repository.IAdminRepository
	hub          *ws.Hub
	cfg          util.Config
	cloudinary   *bucket.CloudinaryService
	sanctumToken *sanctum.SanctumToken
}

func NewAdminService(repo repository.IAdminRepository, hub *ws.Hub, cfg util.Config, cloudinary *bucket.CloudinaryService, sanctumToken *sanctum.SanctumToken) IAdminService {
	return &AdminService{
		repo:         repo,
		hub:          hub,
		cfg:          cfg,
		cloudinary:   cloudinary,
		sanctumToken: sanctumToken,
	}
}

type IAdminService interface {
	CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error)
	VerifyPassword(password, hashedPassword string) bool
	CreateToken(admin migration.Admin) (string, error)
	GetAdminProfile(adminID uuid.UUID) (migration.Admin, error)
	GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error)
	GetUserDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.UserDashboardDataResponse, error)
	GetRideDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.RideDashboardDataResponse, error)
	GetTransactionDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.TransactionDashboardDataResponse, error)
	GetVehicleDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.VehicleDashboardDataResponse, error)
	GetUserList(req schemas.UserListRequest) ([]migration.User, int64, int64, error)
	GetRideList(req schemas.RideListRequest) ([]migration.Ride, int64, int64, error)
	GetVehicleList(req schemas.VehicleListRequest) ([]migration.Vehicle, int64, int64, error)
	GetTransactionList(req schemas.TransactionListRequest) ([]migration.Transaction, int64, int64, error)
	GetDashboardData(req schemas.DashboardReportRequest) (schemas.ReportData, error)
	AnalyzeDashboardData(data schemas.ReportData) (string, error)
	CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error)
	CreatePDFReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error)
}

// CheckAdminExists checks if an admin exists with the given email and password
func (s *AdminService) CheckAdminExists(req schemas.LoginAdminRequest) (migration.Admin, error) {
	return s.repo.CheckAdminExists(req)
}

// VerifyPassword verifies if the given password matches the hashed password
func (s *AdminService) VerifyPassword(password, hashedPassword string) bool {
	return s.sanctumToken.Cryto.VerifyPassword(hashedPassword, password)
}

// CreateToken creates a new token for the admin
func (s *AdminService) CreateToken(admin migration.Admin) (string, error) {
	return s.sanctumToken.CreateSanctumToken(admin.ID, time.Duration(s.cfg.RefreshTokenExpiredDuration)*time.Second)
}

// GetAdminProfile gets the profile of the admin
func (s *AdminService) GetAdminProfile(adminID uuid.UUID) (migration.Admin, error) {
	return s.repo.GetAdminProfile(adminID)
}

// GetDashboardGeneralData gets the general data for the dashboard
func (s *AdminService) GetDashboardGeneralData() (schemas.DashboardGeneralDataResponse, error) {
	return s.repo.GetDashboardGeneralData()
}

// GetUserDashboardData gets the data for the user dashboard
func (s *AdminService) GetUserDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.UserDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.UserDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetUserDashboardData(startDate, endDate)
}

// GetRideDashboardData gets the data for the ride dashboard
func (s *AdminService) GetRideDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.RideDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.RideDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetRideDashboardData(startDate, endDate)
}

// GetTransactionDashboardData gets the data for the transaction dashboard
func (s *AdminService) GetTransactionDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.TransactionDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.TransactionDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetTransactionDashboardData(startDate, endDate)
}

// GetVehicleDashboardData gets the data for the vehicle dashboard
func (s *AdminService) GetVehicleDashboardData(filter string, customStartDate time.Time, customEndDate time.Time) (schemas.VehicleDashboardDataResponse, error) {
	var startDate, endDate time.Time
	now := time.Now()

	switch filter {
	case "all_time":
		// You might want to set a reasonable start date here, or use the earliest record in your database
		startDate = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = now
	case "last_week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "last_month":
		startDate = now.AddDate(0, -1, 0)
		endDate = now
	case "last_year":
		startDate = now.AddDate(-1, 0, 0)
		endDate = now
	case "custom":
		startDate = customStartDate
		endDate = customEndDate
	default:
		return schemas.VehicleDashboardDataResponse{}, fmt.Errorf("invalid filter")
	}
	return s.repo.GetVehicleDashboardData(startDate, endDate)
}

// GetUserList gets the list of users
func (s *AdminService) GetUserList(req schemas.UserListRequest) ([]migration.User, int64, int64, error) {
	return s.repo.GetUserList(req)
}

// GetRideList gets the list of rides
func (s *AdminService) GetRideList(req schemas.RideListRequest) ([]migration.Ride, int64, int64, error) {
	return s.repo.GetRideList(req)
}

// GetVehicleList gets the list of vehicles
func (s *AdminService) GetVehicleList(req schemas.VehicleListRequest) ([]migration.Vehicle, int64, int64, error) {
	return s.repo.GetVehicleList(req)
}

// GetTransactionList gets the list of transactions
func (s *AdminService) GetTransactionList(req schemas.TransactionListRequest) ([]migration.Transaction, int64, int64, error) {
	return s.repo.GetTransactionList(req)
}

// GetDashboardData gets the data for the dashboard report
func (s *AdminService) GetDashboardData(req schemas.DashboardReportRequest) (schemas.ReportData, error) {
	return s.repo.GetDashboardData(req)
}

// AnalyzeDashboardData analyzes the data for the dashboard report using OpenRouter AI with LLM
func (s *AdminService) AnalyzeDashboardData(data schemas.ReportData) (string, error) {
	prompt := fmt.Sprintf(`Phân tích dữ liệu bảng điều khiển sau đây và cung cấp thông tin chi tiết:

Tổng số người dùng: %d
Người dùng hoạt động: %d
Tổng số chuyến đi: %d
Chuyến đi hoàn thành: %d
Chuyến đi bị hủy: %d
Tổng giá trị giao dịch: %d VND
Đánh giá trung bình: %.2f

Tuyến đường phổ biến:
%v

Tăng trưởng người dùng:
%v

Doanh thu theo ngày:
%v

Phân bố loại xe:
%v

Hãy cung cấp phân tích chi tiết về dữ liệu, bao gồm xu hướng, các lĩnh vực tiềm năng cần cải thiện và đề xuất cho sự tăng trưởng kinh doanh. Trả lời bằng tiếng Việt và định dạng phù hợp để dễ dàng đưa vào file Excel. Phân tích nên được chia thành các phần sau, mỗi phần cách nhau bằng dấu hai chấm và xuống dòng:

1. Tổng quan: [Phân tích tổng quan]
2. Xu hướng chính: [Liệt kê và phân tích các xu hướng chính]
3. Điểm mạnh: [Liệt kê và phân tích điểm mạnh]
4. Điểm yếu: [Liệt kê và phân tích điểm yếu]
5. Cơ hội: [Liệt kê và phân tích cơ hội]
6. Thách thức: [Liệt kê và phân tích thách thức]
7. Đề xuất cải thiện: [Liệt kê và giải thích các đề xuất cải thiện]

Mỗi phần nên ngắn gọn, súc tích và dễ hiểu.`,
		data.TotalUsers, data.ActiveUsers, data.TotalRides, data.CompletedRides, data.CancelledRides,
		data.TotalTransactions, data.AverageRating, data.PopularRoutes, data.UserGrowth, data.TransactionByDay,
		data.VehicleTypeDistribution)

	// Call OpenRouter API directly
	requestBody := schemas.OpenRouterRequestBody{
		Model: "mistralai/mistral-nemo",
		Messages: []schemas.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal the request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Create request
	req, err := http.NewRequest("POST", s.cfg.OpenRouterAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+s.cfg.OpenRouterAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Unmarshal response
	var openRouterResponse schemas.OpenRouterResponse
	err = json.Unmarshal(body, &openRouterResponse)
	if err != nil {
		return "", err
	}

	// Extract the response
	if len(openRouterResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenRouter")
	}

	// Extract the response
	response := openRouterResponse.Choices[0].Message.Content

	// Check if the response is a string
	if str, ok := response.(string); ok {
		return str, nil
	} else {
		return "", fmt.Errorf("invalid response from OpenRouter")
	}
}

// CreateExcelReport creates an Excel report from the data and analysis
// func (s *AdminService) CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
// 	f := excelize.NewFile()

// 	// Tạo sheet tổng quan
// 	f.NewSheet("Tổng quan")
// 	f.SetCellValue("Tổng quan", "A1", "Chỉ số")
// 	f.SetCellValue("Tổng quan", "B1", "Giá trị")
// 	f.SetCellValue("Tổng quan", "A2", "Tổng số người dùng")
// 	f.SetCellValue("Tổng quan", "B2", data.TotalUsers)
// 	f.SetCellValue("Tổng quan", "A3", "Người dùng hoạt động")
// 	f.SetCellValue("Tổng quan", "B3", data.ActiveUsers)
// 	f.SetCellValue("Tổng quan", "A4", "Tông số chuyến đi")
// 	f.SetCellValue("Tổng quan", "B4", data.TotalRides)
// 	f.SetCellValue("Tổng quan", "A5", "Số chuyến đi hoàn thành")
// 	f.SetCellValue("Tổng quan", "B5", data.CompletedRides)
// 	f.SetCellValue("Tổng quan", "A6", "Số chuyến đi bị hủy")
// 	f.SetCellValue("Tổng quan", "B6", data.CancelledRides)
// 	f.SetCellValue("Tổng quan", "A7", "Tổng giá trị giao dịch")
// 	f.SetCellValue("Tổng quan", "B7", data.TotalTransactions)
// 	f.SetCellValue("Tổng quan", "A8", "Trung bình đánh giá")
// 	f.SetCellValue("Tổng quan", "B8", data.AverageRating)

// 	// Tạo sheet cho các tuyến đường phổ biến
// 	f.NewSheet("Tuyến đường phổ biến")
// 	f.SetCellValue("", "A1", "Địa chỉ bắt đầu")
// 	f.SetCellValue("Tuyến đường phổ biến", "B1", "Địa chỉ kết thúc")
// 	f.SetCellValue("Tuyến đường phổ biến", "C1", "Tổng số lần")
// 	for i, route := range data.PopularRoutes {
// 		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("A%d", i+2), route.StartAddress)
// 		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("B%d", i+2), route.EndAddress)
// 		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("C%d", i+2), route.Count)
// 	}

// 	// Tạo sheet cho tăng trưởng người dùng
// 	f.NewSheet("Chỉ số tăng trưởng người dùng")
// 	f.SetCellValue("Chỉ số tăng trưởng người dùng", "A1", "Ngày")
// 	f.SetCellValue("Chỉ số tăng trưởng người dùng", "B1", "Số lượng người dùng mới")
// 	for i, growth := range data.UserGrowth {
// 		f.SetCellValue("Chỉ số tăng trưởng người dùng", fmt.Sprintf("A%d", i+2), growth.Date.Format("02/01/2006 15:04"))
// 		f.SetCellValue("Chỉ số tăng trưởng người dùng", fmt.Sprintf("B%d", i+2), growth.Count)
// 	}

// 	// Tạo sheet cho giao dịch theo ngày
// 	f.NewSheet("Giao dịch theo ngày")
// 	f.SetCellValue("Giao dịch theo ngày", "A1", "Ngày")
// 	f.SetCellValue("Giao dịch theo ngày", "B1", "Tổng giá trị")
// 	for i, revenue := range data.TransactionByDay {
// 		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("A%d", i+2), revenue.Date.Format("02/01/2006 15:04"))
// 		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("B%d", i+2), revenue.Transaction)
// 	}

// 	// Tạo sheet cho phân tích
// 	f.NewSheet("Phân tích")
// 	f.SetCellValue("Phân tích", "A1", "Phân tích")
// 	f.SetCellValue("Phân tích", "A2", analysis)

// 	buffer, err := f.WriteToBuffer()
// 	if err != nil {
// 		return nil, err
// 	}

//		return buffer, nil
//	}
func (s *AdminService) CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	// Define common styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16, Color: "#1F497D", Family: "Arial"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "bottom", Color: "#1F497D", Style: 2}},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF", Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "#000000", Style: 1}, {Type: "top", Color: "#000000", Style: 1}, {Type: "bottom", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1}},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Arial"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "#000000", Style: 1}, {Type: "top", Color: "#000000", Style: 1}, {Type: "bottom", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1}},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 11, Family: "Arial"},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:       []excelize.Border{{Type: "left", Color: "#000000", Style: 1}, {Type: "top", Color: "#000000", Style: 1}, {Type: "bottom", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1}},
		CustomNumFmt: &[]string{"#,##0"}[0],
	})

	// Common chart style
	chartStyle := &excelize.Chart{
		Legend: excelize.ChartLegend{
			Position: "bottom",
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName: true,
			ShowPercent: true,
			ShowSerName: true,
			ShowVal:     true,
		},
	}

	// Create Overview sheet
	overviewSheet := "Tổng quan"
	f.NewSheet(overviewSheet)
	f.SetColWidth(overviewSheet, "A", "B", 25)
	f.SetRowHeight(overviewSheet, 1, 30)

	// Set title
	f.MergeCell(overviewSheet, "A1", "B1")
	f.SetCellValue(overviewSheet, "A1", "Báo cáo Tổng quan")
	f.SetCellStyle(overviewSheet, "A1", "B1", titleStyle)

	// Overview table
	headers := []string{"Chỉ số", "Giá trị"}
	overviewData := [][]interface{}{
		{"Tổng số người dùng", data.TotalUsers},
		{"Người dùng hoạt động", data.ActiveUsers},
		{"Tổng số chuyến đi", data.TotalRides},
		{"Chuyến đi hoàn thành", data.CompletedRides},
		{"Chuyến đi bị hủy", data.CancelledRides},
		{"Tổng giá trị giao dịch", data.TotalTransactions},
		{"Trung bình đánh giá", data.AverageRating},
	}

	// Apply styles and set data for overview
	setTableData(f, overviewSheet, headers, overviewData, headerStyle, dataStyle, numberStyle, 3)

	// Vehicle Distribution Sheet
	if len(data.VehicleTypeDistribution) > 0 {
		pieChart := "Phân bố loại xe"
		createChartSheet(f, pieChart, data.VehicleTypeDistribution, headerStyle, dataStyle, chartStyle)
	}

	// User Growth Sheet
	if len(data.UserGrowth) > 0 {
		userGrowthSheet := "Tăng trưởng người dùng"
		createTimeSeriesSheet(f, userGrowthSheet, data.UserGrowth, headerStyle, dataStyle, numberStyle, chartStyle)
	}

	// Analysis Sheet
	analysisSheet := "Phân tích"
	createAnalysisSheet(f, analysisSheet, analysis, titleStyle)

	// Popular Routes Sheet
	if len(data.PopularRoutes) > 0 {
		routesSheet := "Tuyến đường phổ biến"
		createRoutesSheet(f, routesSheet, data.PopularRoutes, headerStyle, dataStyle, numberStyle)
	}

	// Transaction Sheet
	if len(data.TransactionByDay) > 0 {
		transactionSheet := "Giao dịch theo ngày"
		createTransactionSheet(f, transactionSheet, data.TransactionByDay, headerStyle, dataStyle, numberStyle, chartStyle)
	}

	// Set default sheet
	sheetIndex, _ := f.GetSheetIndex(overviewSheet)
	f.SetActiveSheet(sheetIndex)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// Helper functions

func setTableData(f *excelize.File, sheet string, headers []string, data [][]interface{}, headerStyle, dataStyle, numberStyle int, startRow int) {
	// Set headers
	for col, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+col)), startRow)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Set data
	for row, record := range data {
		for col, value := range record {
			cell := fmt.Sprintf("%s%d", string(rune('A'+col)), row+startRow+1)
			f.SetCellValue(sheet, cell, value)
			if col == 0 {
				f.SetCellStyle(sheet, cell, cell, dataStyle)
			} else {
				f.SetCellStyle(sheet, cell, cell, numberStyle)
			}
		}
	}
}

func createChartSheet(f *excelize.File, sheet string, data []schemas.VehicleTypeData, headerStyle, dataStyle int, chartStyle *excelize.Chart) {
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "B", 20)

	// Set title
	f.MergeCell(sheet, "A1", "B1")
	f.SetCellValue(sheet, "A1", sheet)

	headers := []string{"Loại xe", "Số lượng"}
	var chartData [][]interface{}
	for _, vt := range data {
		chartData = append(chartData, []interface{}{vt.Type, vt.Count})
	}

	setTableData(f, sheet, headers, chartData, headerStyle, dataStyle, dataStyle, 2)

	pieChartStyle := *chartStyle
	pieChartStyle.Type = excelize.Pie
	pieChartStyle.Series = []excelize.ChartSeries{
		{
			Name:       sheet,
			Categories: fmt.Sprintf("%s!$A$3:$A$%d", sheet, len(data)+2),
			Values:     fmt.Sprintf("%s!$B$3:$B$%d", sheet, len(data)+2),
		},
	}
	pieChartStyle.Title = []excelize.RichTextRun{{Text: sheet}}

	f.AddChart(sheet, "D2", &pieChartStyle)
}

func createTimeSeriesSheet(f *excelize.File, sheet string, data []schemas.UserGrowthData, headerStyle, dataStyle, numberStyle int, chartStyle *excelize.Chart) {
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "B", 20)

	// Set title
	f.MergeCell(sheet, "A1", "B1")
	f.SetCellValue(sheet, "A1", sheet)

	headers := []string{"Ngày", "Số lượng"}
	var chartData [][]interface{}
	for _, ug := range data {
		chartData = append(chartData, []interface{}{ug.Date.Format("02/01/2006"), ug.Count})
	}

	setTableData(f, sheet, headers, chartData, headerStyle, dataStyle, numberStyle, 2)

	lineChartStyle := *chartStyle
	lineChartStyle.Type = excelize.Line
	lineChartStyle.Series = []excelize.ChartSeries{
		{
			Name:       sheet,
			Categories: fmt.Sprintf("%s!$A$3:$A$%d", sheet, len(data)+2),
			Values:     fmt.Sprintf("%s!$B$3:$B$%d", sheet, len(data)+2),
		},
	}
	lineChartStyle.Title = []excelize.RichTextRun{{Text: sheet}}
	lineChartStyle.XAxis = excelize.ChartAxis{MajorUnit: 1}
	lineChartStyle.YAxis = excelize.ChartAxis{MajorUnit: 10}

	f.AddChart(sheet, "D2", &lineChartStyle)
}

func createAnalysisSheet(f *excelize.File, sheet string, analysis string, titleStyle int) {
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "A", 100)

	// Set title
	f.SetCellValue(sheet, "A1", "Phân tích chi tiết")
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetRowHeight(sheet, 1, 30)

	// Set analysis content
	f.SetCellValue(sheet, "A2", analysis)
}

func createRoutesSheet(f *excelize.File, sheet string, data []schemas.PopularRoute, headerStyle, dataStyle, numberStyle int) {
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "C", 30)

	// Set title
	f.MergeCell(sheet, "A1", "C1")
	f.SetCellValue(sheet, "A1", sheet)

	headers := []string{"Địa chỉ bắt đầu", "Địa chỉ kết thúc", "Tổng số lần"}
	var routeData [][]interface{}
	for _, route := range data {
		routeData = append(routeData, []interface{}{route.StartAddress, route.EndAddress, route.Count})
	}

	setTableData(f, sheet, headers, routeData, headerStyle, dataStyle, numberStyle, 2)
}

func createTransactionSheet(f *excelize.File, sheet string, data []schemas.TransactionDayData, headerStyle, dataStyle, numberStyle int, chartStyle *excelize.Chart) {
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "B", 20)

	// Set title
	f.MergeCell(sheet, "A1", "B1")
	f.SetCellValue(sheet, "A1", sheet)

	headers := []string{"Ngày", "Tổng giá trị"}
	var transactionData [][]interface{}
	for _, td := range data {
		transactionData = append(transactionData, []interface{}{td.Date.Format("02/01/2006"), td.Transaction})
	}

	setTableData(f, sheet, headers, transactionData, headerStyle, dataStyle, numberStyle, 2)

	colChartStyle := *chartStyle
	colChartStyle.Type = excelize.Col
	colChartStyle.Series = []excelize.ChartSeries{
		{
			Name:       sheet,
			Categories: fmt.Sprintf("%s!$A$3:$A$%d", sheet, len(data)+2),
			Values:     fmt.Sprintf("%s!$B$3:$B$%d", sheet, len(data)+2),
		},
	}
	colChartStyle.Title = []excelize.RichTextRun{{Text: sheet}}
	colChartStyle.XAxis = excelize.ChartAxis{MajorUnit: 1}
	colChartStyle.YAxis = excelize.ChartAxis{MajorUnit: 1000000}

	f.AddChart(sheet, "D2", &colChartStyle)
}

// CreatePDFReport creates a PDF report from the data and analysis
func (s *AdminService) CreatePDFReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	// **Section 1: Report Title**
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Báo cáo Bảng Điều Khiển")
	pdf.Ln(15)

	// **Section 2: Summary Data**
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "1. Tổng Quan")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(70, 10, "Tổng số người dùng:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.TotalUsers), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Người dùng hoạt động:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.ActiveUsers), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Tổng số chuyến đi:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.TotalRides), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Chuyến đi hoàn thành:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.CompletedRides), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Chuyến đi bị hủy:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.CancelledRides), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Tổng giá trị giao dịch (VND):", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%d", data.TotalTransactions), "", 1, "", false, 0, "")
	pdf.CellFormat(70, 10, "Đánh giá trung bình:", "", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, fmt.Sprintf("%.2f", data.AverageRating), "", 1, "", false, 0, "")
	pdf.Ln(10)

	// **Section 3: Analysis**
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "2. Phân Tích")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, analysis, "", "", false)
	pdf.Ln(10)

	// **Section 4: Popular Routes**
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "3. Tuyến Đường Phổ Biến")
	pdf.Ln(12)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(60, 10, "Địa chỉ bắt đầu", "1", 0, "C", true, 0, "")
	pdf.CellFormat(60, 10, "Địa chỉ kết thúc", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 10, "Số lượt", "1", 1, "C", true, 0, "")

	// Table data
	pdf.SetFont("Arial", "", 12)
	for _, route := range data.PopularRoutes {
		pdf.CellFormat(60, 10, route.StartAddress, "1", 0, "", false, 0, "")
		pdf.CellFormat(60, 10, route.EndAddress, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d", route.Count), "1", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// **Section 5: Vehicle Type Distribution**
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "4. Phân Bố Loại Xe")
	pdf.Ln(12)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(70, 10, "Loại xe", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Số lượng", "1", 1, "C", true, 0, "")

	// Table data
	pdf.SetFont("Arial", "", 12)
	for _, vehicle := range data.VehicleTypeDistribution {
		pdf.CellFormat(70, 10, vehicle.Type, "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, fmt.Sprintf("%d", vehicle.Count), "1", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// **Section 6: Charts**
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "5. Biểu Đồ")
	pdf.Ln(12)

	// Generate and add User Growth chart
	userGrowthChartPath := "user_growth_chart.png"
	err := helper.GenerateUserGrowthChart(data.UserGrowth, userGrowthChartPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(userGrowthChartPath)

	pdf.ImageOptions(userGrowthChartPath, 15, pdf.GetY(), 180, 90, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.Ln(95)

	// Generate and add Transactions by Day chart
	transactionChartPath := "transaction_chart.png"
	err = helper.GenerateTransactionChart(data.TransactionByDay, transactionChartPath)
	if err != nil {
		return nil, err
	}
	defer os.Remove(transactionChartPath)

	pdf.ImageOptions(transactionChartPath, 15, pdf.GetY(), 180, 90, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	pdf.Ln(95)

	// Output to buffer
	buf := new(bytes.Buffer)
	err = pdf.Output(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Ensure that the AdminService implements the IAdminService interface
var _ IAdminService = (*AdminService)(nil)
