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
	"strings"
	"time"

	"github.com/google/uuid"
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

// // CreateExcelReport creates an Excel report from the data and analysis
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

// 	return buffer, nil
// }

func (s *AdminService) CreateExcelReport(data schemas.ReportData, analysis string) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	// Định nghĩa các style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#E0EBF5"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 43,
	})

	// Tạo sheet tổng quan
	f.NewSheet("Tổng quan")
	f.SetCellValue("Tổng quan", "A1", "Báo cáo Tổng quan")
	f.MergeCell("Tổng quan", "A1", "B1")
	f.SetCellStyle("Tổng quan", "A1", "B1", titleStyle)

	headers := []string{"Chỉ số", "Giá trị"}
	reportData := [][]interface{}{
		{"Tổng số người dùng", data.TotalUsers},
		{"Người dùng hoạt động", data.ActiveUsers},
		{"Tổng số chuyến đi", data.TotalRides},
		{"Số chuyến đi hoàn thành", data.CompletedRides},
		{"Số chuyến đi bị hủy", data.CancelledRides},
		{"Tổng giá trị giao dịch", data.TotalTransactions},
		{"Trung bình đánh giá", data.AverageRating},
	}

	for col, header := range headers {
		cell := fmt.Sprintf("%c2", 'A'+col)
		f.SetCellValue("Tổng quan", cell, header)
		f.SetCellStyle("Tổng quan", cell, cell, headerStyle)
	}

	for row, record := range reportData {
		for col, value := range record {
			cell := fmt.Sprintf("%c%d", 'A'+col, row+3)
			f.SetCellValue("Tổng quan", cell, value)
			if col == 1 {
				f.SetCellStyle("Tổng quan", cell, cell, numberStyle)
			}
		}
	}

	f.SetColWidth("Tổng quan", "A", "B", 25)

	// Tạo biểu đồ tròn cho tỷ lệ chuyến đi hoàn thành/hủy
	if err := createPieChart(f, "Tổng quan", "D2", "Tỷ lệ chuyến đi", data.CompletedRides, data.CancelledRides); err != nil {
		return nil, err
	}

	// Tạo sheet cho các tuyến đường phổ biến
	f.NewSheet("Tuyến đường phổ biến")
	f.SetCellValue("Tuyến đường phổ biến", "A1", "Các Tuyến Đường Phổ Biến")
	f.MergeCell("Tuyến đường phổ biến", "A1", "C1")
	f.SetCellStyle("Tuyến đường phổ biến", "A1", "C1", titleStyle)

	headers = []string{"Địa chỉ bắt đầu", "Địa chỉ kết thúc", "Tổng số lần"}
	for col, header := range headers {
		cell := fmt.Sprintf("%c2", 'A'+col)
		f.SetCellValue("Tuyến đường phổ biến", cell, header)
		f.SetCellStyle("Tuyến đường phổ biến", cell, cell, headerStyle)
	}

	for i, route := range data.PopularRoutes {
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("A%d", i+3), route.StartAddress)
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("B%d", i+3), route.EndAddress)
		f.SetCellValue("Tuyến đường phổ biến", fmt.Sprintf("C%d", i+3), route.Count)
	}

	f.SetColWidth("Tuyến đường phổ biến", "A", "B", 30)
	f.SetColWidth("Tuyến đường phổ biến", "C", "C", 15)

	// Tạo biểu đồ cột cho các tuyến đường phổ biến
	if err := createBarChart(f, "Tuyến đường phổ biến", "E2", "Top 5 Tuyến Đường Phổ Biến", data.PopularRoutes); err != nil {
		return nil, err
	}

	// Tạo sheet cho tăng trưởng người dùng
	f.NewSheet("Tăng trưởng người dùng")
	f.SetCellValue("Tăng trưởng người dùng", "A1", "Chỉ số Tăng Trưởng Người Dùng")
	f.MergeCell("Tăng trưởng người dùng", "A1", "B1")
	f.SetCellStyle("Tăng trưởng người dùng", "A1", "B1", titleStyle)

	headers = []string{"Ngày", "Số lượng người dùng mới"}
	for col, header := range headers {
		cell := fmt.Sprintf("%c2", 'A'+col)
		f.SetCellValue("Tăng trưởng người dùng", cell, header)
		f.SetCellStyle("Tăng trưởng người dùng", cell, cell, headerStyle)
	}

	for i, growth := range data.UserGrowth {
		f.SetCellValue("Tăng trưởng người dùng", fmt.Sprintf("A%d", i+3), growth.Date.Format("02/01/2006"))
		f.SetCellValue("Tăng trưởng người dùng", fmt.Sprintf("B%d", i+3), growth.Count)
	}

	f.SetColWidth("Tăng trưởng người dùng", "A", "B", 25)

	// Tạo biểu đồ đường cho tăng trưởng người dùng
	if err := createLineChart(f, "Tăng trưởng người dùng", "D2", "Tăng Trưởng Người Dùng Theo Thời Gian", data.UserGrowth); err != nil {
		return nil, err
	}

	// Tạo sheet cho giao dịch theo ngày
	f.NewSheet("Giao dịch theo ngày")
	f.SetCellValue("Giao dịch theo ngày", "A1", "Giao Dịch Theo Ngày")
	f.MergeCell("Giao dịch theo ngày", "A1", "B1")
	f.SetCellStyle("Giao dịch theo ngày", "A1", "B1", titleStyle)

	headers = []string{"Ngày", "Tổng giá trị"}
	for col, header := range headers {
		cell := fmt.Sprintf("%c2", 'A'+col)
		f.SetCellValue("Giao dịch theo ngày", cell, header)
		f.SetCellStyle("Giao dịch theo ngày", cell, cell, headerStyle)
	}

	for i, transaction := range data.TransactionByDay {
		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("A%d", i+3), transaction.Date.Format("02/01/2006"))
		f.SetCellValue("Giao dịch theo ngày", fmt.Sprintf("B%d", i+3), transaction.Transaction)
		f.SetCellStyle("Giao dịch theo ngày", fmt.Sprintf("B%d", i+3), fmt.Sprintf("B%d", i+3), numberStyle)
	}

	f.SetColWidth("Giao dịch theo ngày", "A", "B", 25)

	// Tạo biểu đồ cột cho giao dịch theo ngày
	if err := createColumnChart(f, "Giao dịch theo ngày", "D2", "Giao Dịch Theo Ngày", data.TransactionByDay); err != nil {
		return nil, err
	}

	// Tạo sheet cho phân tích
	f.NewSheet("Phân tích")
	f.SetCellValue("Phân tích", "A1", "Phân Tích Chi Tiết")
	f.SetCellStyle("Phân tích", "A1", "A1", titleStyle)

	// Phân tích văn bản thành các phần
	sections := strings.Split(analysis, "\n")
	row := 2
	for _, section := range sections {
		if strings.Contains(section, ":") {
			parts := strings.SplitN(section, ":", 2)
			if len(parts) == 2 {
				f.SetCellValue("Phân tích", fmt.Sprintf("A%d", row), strings.TrimSpace(parts[0]))
				f.SetCellValue("Phân tích", fmt.Sprintf("B%d", row), strings.TrimSpace(parts[1]))
				f.SetCellStyle("Phân tích", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), headerStyle)
				row++
			}
		}
	}

	f.SetColWidth("Phân tích", "A", "A", 20)
	f.SetColWidth("Phân tích", "B", "B", 100)

	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// Hàm tạo biểu đồ tròn
func createPieChart(f *excelize.File, sheet, cell, title string, completed, cancelled int64) error {
	categories := fmt.Sprintf("='{%s}'!$A$4:$A$5", sheet)
	values := fmt.Sprintf("='{%s}'!$B$4:$B$5", sheet)

	if err := f.AddChart(sheet, cell, &excelize.Chart{
		Type: excelize.Pie,
		Series: []excelize.ChartSeries{
			{
				Name:       "Tỷ lệ chuyến đi",
				Categories: categories,
				Values:     values,
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: title,
			},
		},
	}); err != nil {
		return err
	}
	return nil
}

// Hàm tạo biểu đồ cột
func createBarChart(f *excelize.File, sheet, cell, title string, routes []schemas.PopularRoute) error {
	categories := fmt.Sprintf("='{%s}'!$A$3:$A$7", sheet)
	values := fmt.Sprintf("='{%s}'!$C$3:$C$7", sheet)

	if err := f.AddChart(sheet, cell, &excelize.Chart{
		Type: excelize.Bar,
		Series: []excelize.ChartSeries{
			{
				Name:       "Số lượt",
				Categories: categories,
				Values:     values,
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: title,
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     false,
			ShowSerName:     false,
			ShowVal:         true,
		},
	}); err != nil {
		return err
	}
	return nil
}

// Hàm tạo biểu đồ đường
func createLineChart(f *excelize.File, sheet, cell, title string, growth []schemas.UserGrowthData) error {
	categories := fmt.Sprintf("='{%s}'!$A$3:$A$%d", sheet, len(growth)+2)
	values := fmt.Sprintf("='{%s}'!$B$3:$B$%d", sheet, len(growth)+2)

	if err := f.AddChart(sheet, cell, &excelize.Chart{
		Type: excelize.Line,
		Series: []excelize.ChartSeries{
			{
				Name:       "Số người dùng mới",
				Categories: categories,
				Values:     values,
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: title,
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     false,
			ShowSerName:     false,
			ShowVal:         true,
		},
	}); err != nil {
		return err
	}
	return nil
}

// Hàm tạo biểu đồ cột
func createColumnChart(f *excelize.File, sheet, cell, title string, transactions []schemas.TransactionDayData) error {
	categories := fmt.Sprintf("='{%s}'!$A$3:$A$%d", sheet, len(transactions)+2)
	values := fmt.Sprintf("='{%s}'!$B$3:$B$%d", sheet, len(transactions)+2)

	if err := f.AddChart(sheet, cell, &excelize.Chart{
		Type: excelize.Col,
		Series: []excelize.ChartSeries{
			{
				Name:       "Tổng giá trị giao dịch",
				Categories: categories,
				Values:     values,
			},
		},
		Title: []excelize.RichTextRun{
			{
				Text: title,
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     false,
			ShowSerName:     false,
			ShowVal:         true,
		},
	}); err != nil {
		return err
	}
	return nil
}

// Ensure that the AdminService implements the IAdminService interface
var _ IAdminService = (*AdminService)(nil)
