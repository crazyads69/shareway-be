package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	"github.com/russross/blackfriday/v2"
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
		Temperature: &[]float64{0.5}[0],
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

	// Định nghĩa kiểu chung
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "#1F497D", Family: "Arial"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#DCE6F1"}, Pattern: 1},
		Border:    []excelize.Border{{Type: "bottom", Color: "#1F497D", Style: 2}},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 12, Color: "#FFFFFF", Family: "Arial"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "#000000", Style: 1}, {Type: "top", Color: "#000000", Style: 1}, {Type: "bottom", Color: "#000000", Style: 1}, {Type: "right", Color: "#000000", Style: 1}},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Family: "Arial"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
		Border:    []excelize.Border{{Type: "left", Color: "#D9D9D9", Style: 1}, {Type: "top", Color: "#D9D9D9", Style: 1}, {Type: "bottom", Color: "#D9D9D9", Style: 1}, {Type: "right", Color: "#D9D9D9", Style: 1}},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
	})

	numberStyle, _ := f.NewStyle(&excelize.Style{
		Font:         &excelize.Font{Size: 11, Family: "Arial"},
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border:       []excelize.Border{{Type: "left", Color: "#D9D9D9", Style: 1}, {Type: "top", Color: "#D9D9D9", Style: 1}, {Type: "bottom", Color: "#D9D9D9", Style: 1}, {Type: "right", Color: "#D9D9D9", Style: 1}},
		Fill:         excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
		CustomNumFmt: &[]string{"#,##0"}[0],
	})

	// Hàm helper để đặt dữ liệu bảng và áp dụng kiểu
	setTableData := func(sheet string, headers []string, data [][]interface{}, startRow int) {
		// Đặt tiêu đề
		for col, header := range headers {
			cell := fmt.Sprintf("%s%d", string(rune('A'+col)), startRow)
			f.SetCellValue(sheet, cell, header)
			f.SetCellStyle(sheet, cell, cell, headerStyle)
		}

		// Đặt dữ liệu
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

		// Thêm định dạng có điều kiện
		lastRow := startRow + len(data)
		lastCol := string(rune('A' + len(headers) - 1))
		styleID, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
		})
		if err != nil {
			return
		}
		f.SetConditionalFormat(sheet, fmt.Sprintf("A%d:%s%d", startRow+1, lastCol, lastRow), []excelize.ConditionalFormatOptions{
			{
				Type:     "expression",
				Criteria: "=MOD(ROW(),2)=0",
				Format:   &styleID,
			},
		})
	}

	// Tạo trang Tổng quan
	overviewSheet := "Tổng quan"
	f.NewSheet(overviewSheet)
	f.SetColWidth(overviewSheet, "A", "B", 30)
	f.SetRowHeight(overviewSheet, 1, 40)

	f.MergeCell(overviewSheet, "A1", "B1")
	f.SetCellValue(overviewSheet, "A1", "Báo cáo Tổng quan")
	f.SetCellStyle(overviewSheet, "A1", "B1", titleStyle)

	overviewHeaders := []string{"Chỉ số", "Giá trị"}
	overviewData := [][]interface{}{
		{"Tổng số người dùng", data.TotalUsers},
		{"Người dùng hoạt động", data.ActiveUsers},
		{"Tổng số chuyến đi", data.TotalRides},
		{"Chuyến đi hoàn thành", data.CompletedRides},
		{"Chuyến đi bị hủy", data.CancelledRides},
		{"Tổng giá trị giao dịch", data.TotalTransactions},
		{"Trung bình đánh giá", data.AverageRating},
	}
	setTableData(overviewSheet, overviewHeaders, overviewData, 3)

	// Tạo trang Tóm tắt
	summarySheet := "Tóm tắt"
	f.NewSheet(summarySheet)
	f.SetColWidth(summarySheet, "A", "C", 25)

	f.MergeCell(summarySheet, "A1", "C1")
	f.SetCellValue(summarySheet, "A1", "Tóm tắt báo cáo")
	f.SetCellStyle(summarySheet, "A1", "C1", titleStyle)

	summaryHeaders := []string{"Chỉ số", "Giá trị", "Tỷ lệ"}
	summaryData := [][]interface{}{
		{"Tổng số người dùng", data.TotalUsers, fmt.Sprintf("%.2f%%", float64(data.ActiveUsers)/float64(data.TotalUsers)*100)},
		{"Tổng số chuyến đi", data.TotalRides, fmt.Sprintf("%.2f%%", float64(data.CompletedRides)/float64(data.TotalRides)*100)},
		{"Tổng giá trị giao dịch", data.TotalTransactions, ""},
		{"Trung bình đánh giá", data.AverageRating, ""},
	}
	setTableData(summarySheet, summaryHeaders, summaryData, 3)

	// Tạo trang Tuyến đường phổ biến (nếu có dữ liệu)
	if len(data.PopularRoutes) > 0 {
		routesSheet := "Tuyến đường phổ biến"
		f.NewSheet(routesSheet)
		f.SetColWidth(routesSheet, "A", "C", 25)

		f.MergeCell(routesSheet, "A1", "C1")
		f.SetCellValue(routesSheet, "A1", routesSheet)
		f.SetCellStyle(routesSheet, "A1", "C1", titleStyle)

		routesHeaders := []string{"Địa chỉ bắt đầu", "Địa chỉ kết thúc", "Số lượt"}
		var routesData [][]interface{}
		for _, route := range data.PopularRoutes {
			routesData = append(routesData, []interface{}{route.StartAddress, route.EndAddress, route.Count})
		}
		setTableData(routesSheet, routesHeaders, routesData, 2)
	}

	// Tạo trang Giao dịch theo ngày (nếu có dữ liệu)
	if len(data.TransactionByDay) > 0 {
		transactionSheet := "Giao dịch theo ngày"
		f.NewSheet(transactionSheet)
		f.SetColWidth(transactionSheet, "A", "B", 20)

		f.MergeCell(transactionSheet, "A1", "B1")
		f.SetCellValue(transactionSheet, "A1", transactionSheet)
		f.SetCellStyle(transactionSheet, "A1", "B1", titleStyle)

		transactionHeaders := []string{"Ngày", "Tổng giá trị"}
		var transactionData [][]interface{}
		for _, td := range data.TransactionByDay {
			transactionData = append(transactionData, []interface{}{td.Date.Format("02/01/2006"), td.Transaction})
		}
		setTableData(transactionSheet, transactionHeaders, transactionData, 2)
	}

	// Đặt trang mặc định
	sheetIndex, err := f.GetSheetIndex(overviewSheet)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(sheetIndex)

	// Ghi vào buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// CreatePDFReport creates a PDF report from the data and analysis
func (s *AdminService) CreatePDFReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	ht := pdf.PointConvert(12)

	// Thêm đầu trang
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 10, "Báo cáo Bảng Điều Khiển")
		pdf.Ln(5)
	})

	// Thêm chân trang với số trang
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Trang %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("{nb}") // Để sử dụng tổng số trang trong chân trang

	// Tạo mục lục
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Mục lục")
	pdf.Ln(15)

	sections := []string{"Tổng Quan", "Phân Tích", "Tuyến Đường Phổ Biến", "Phân Bố Loại Xe"}
	for i, section := range sections {
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 10, fmt.Sprintf("%d. %s", i+1, section))
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(0, 10, fmt.Sprintf("%d", i+2), "", 1, "R", false, 0, "")
	}

	// Nội dung báo cáo
	pdf.AddPage()

	// **Section 1: Tổng Quan**
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "1. Tổng Quan")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	summaryData := [][]string{
		{"Tổng số người dùng:", fmt.Sprintf("%d", data.TotalUsers)},
		{"Người dùng hoạt động:", fmt.Sprintf("%d", data.ActiveUsers)},
		{"Tổng số chuyến đi:", fmt.Sprintf("%d", data.TotalRides)},
		{"Chuyến đi hoàn thành:", fmt.Sprintf("%d", data.CompletedRides)},
		{"Chuyến đi bị hủy:", fmt.Sprintf("%d", data.CancelledRides)},
		{"Tổng giá trị giao dịch (VND):", fmt.Sprintf("%d", data.TotalTransactions)},
		{"Đánh giá trung bình:", fmt.Sprintf("%.2f", data.AverageRating)},
	}

	for _, row := range summaryData {
		pdf.CellFormat(70, 10, row[0], "", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, row[1], "", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// **Section 2: Phân Tích**
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "2. Phân Tích")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)

	// Chuyển đổi markdown thành HTML
	html := blackfriday.Run([]byte(analysis))

	// Phân tích HTML và thêm vào PDF
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	htmlFile := pdf.HTMLBasicNew()
	htmlFile.Write(ht, tr(string(html)))
	pdf.Ln(10)

	// **Section 3: Tuyến Đường Phổ Biến**
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
		// Kiểm tra nếu cần thêm trang mới
		if pdf.GetY() > 250 {
			pdf.AddPage()
			// In lại header của bảng
			pdf.SetFillColor(200, 200, 200)
			pdf.SetFont("Arial", "B", 12)
			pdf.CellFormat(60, 10, "Địa chỉ bắt đầu", "1", 0, "C", true, 0, "")
			pdf.CellFormat(60, 10, "Địa chỉ kết thúc", "1", 0, "C", true, 0, "")
			pdf.CellFormat(30, 10, "Số lượt", "1", 1, "C", true, 0, "")
			pdf.SetFont("Arial", "", 12)
		}
		pdf.CellFormat(60, 10, route.StartAddress, "1", 0, "", false, 0, "")
		pdf.CellFormat(60, 10, route.EndAddress, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("%d", route.Count), "1", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// **Section 4: Phân Bố Loại Xe**
	pdf.AddPage()
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

	// Output to buffer
	buf := new(bytes.Buffer)
	if err := pdf.Output(buf); err != nil {
		return nil, fmt.Errorf("error outputting PDF: %w", err)
	}

	return buf, nil
}

// Ensure that the AdminService implements the IAdminService interface
var _ IAdminService = (*AdminService)(nil)
