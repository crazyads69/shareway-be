package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
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
	LoadFonts(pdf *gofpdf.Fpdf) error
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
		Model: "nousresearch/hermes-3-llama-3.1-70b",
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

func (s *AdminService) LoadFonts(pdf *gofpdf.Fpdf) error {
	fontPaths := []struct {
		family string
		style  string
		file   string
	}{
		{"DejaVu", "", "fonts/DejaVuSansCondensed.ttf"},
		{"DejaVu", "B", "fonts/DejaVuSansCondensed-Bold.ttf"},
		{"DejaVu", "I", "fonts/DejaVuSansCondensed-Oblique.ttf"},
	}

	for _, font := range fontPaths {
		// Check if font file exists
		if _, err := os.Stat(font.file); os.IsNotExist(err) {
			return fmt.Errorf("font file not found: %s", font.file)
		}

		pdf.AddUTF8Font(font.family, font.style, font.file)
	}
	return nil
}

// CreateExcelReport creates an Excel report from the data and analysis
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

	// Tạo trang Phân bố loại xe (nếu có dữ liệu)
	if len(data.VehicleTypeDistribution) > 0 {
		vehicleSheet := "Phân bố loại xe"
		f.NewSheet(vehicleSheet)
		f.SetColWidth(vehicleSheet, "A", "B", 20)

		f.MergeCell(vehicleSheet, "A1", "B1")
		f.SetCellValue(vehicleSheet, "A1", vehicleSheet)
		f.SetCellStyle(vehicleSheet, "A1", "B1", titleStyle)

		vehicleHeaders := []string{"Loại xe", "Số lượng"}
		var vehicleData [][]interface{}
		for _, vd := range data.VehicleTypeDistribution {
			vehicleData = append(vehicleData, []interface{}{vd.Type, vd.Count})
		}
		setTableData(vehicleSheet, vehicleHeaders, vehicleData, 2)
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

	// Load fonts
	if err := s.LoadFonts(pdf); err != nil {
		return nil, fmt.Errorf("error loading fonts: %w", err)
	}

	pdf.SetFont("DejaVu", "", 12)

	// Đặt lề
	topMargin := 30.0
	bottomMargin := 20.0
	leftMargin := 10.0
	rightMargin := 10.0
	pdf.SetMargins(leftMargin, topMargin, rightMargin)

	// Bật tự động thêm trang mới
	pdf.SetAutoPageBreak(true, bottomMargin)

	// Thêm đầu trang
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("DejaVu", "B", 12)
		pdf.SetY(10)
		pdf.Cell(0, 10, "ShareWay - Báo Cáo Dữ Liệu")
		pdf.Ln(5)
	})

	// Thêm chân trang với số trang
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("DejaVu", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Trang %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("{nb}")

	// Tạo mục lục
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 16)
	pdf.Cell(0, 10, "Mục lục")
	pdf.Ln(15)

	sections := []string{"Tổng Quan", "Phân Tích", "Tuyến Đường Phổ Biến", "Phân Bố Loại Xe"}
	for i, section := range sections {
		pdf.SetFont("DejaVu", "", 12)
		pdf.Cell(0, 10, fmt.Sprintf("%d. %s", i+1, section))
		pdf.SetFont("DejaVu", "", 12)
		pdf.CellFormat(0, 10, fmt.Sprintf("%d", i+2), "", 1, "R", false, 0, "")
	}

	// Nội dung báo cáo
	pdf.AddPage()

	checkAndAddPage := func() {
		if pdf.GetY() > 270 { // Điều chỉnh giá trị này nếu cần
			pdf.AddPage()
			pdf.SetY(topMargin) // Đặt lại vị trí Y sau khi thêm trang mới
		}
	}

	// **Section 1: Tổng Quan**
	checkAndAddPage()
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 10, "1. Tổng Quan")
	pdf.Ln(12)

	pdf.SetFont("DejaVu", "", 12)
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
		checkAndAddPage()
		pdf.CellFormat(70, 10, row[0], "", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, row[1], "", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// **Section 2: Phân Tích**
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 10, "2. Phân Tích")
	pdf.Ln(12)

	pdf.SetFont("DejaVu", "", 12)

	// Xử lý và thêm nội dung phân tích vào PDF
	lines := strings.Split(analysis, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			pdf.Ln(5)
		} else if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			// Tiêu đề
			checkAndAddPage()
			pdf.SetFont("DejaVu", "B", 12)
			pdf.MultiCell(0, 6, line[2:len(line)-2], "", "", false)
			pdf.Ln(3)
			pdf.SetFont("DejaVu", "", 12)
		} else if strings.HasPrefix(line, "-") {
			// Danh sách không có thứ tự
			checkAndAddPage()
			pdf.SetX(pdf.GetX() + 5)
			pdf.MultiCell(0, 6, "• "+line[1:], "", "", false)
			pdf.Ln(3)
		} else {
			// Văn bản thông thường
			checkAndAddPage()
			pdf.MultiCell(0, 6, line, "", "", false)
			pdf.Ln(3)
		}
		checkAndAddPage() // Kiểm tra lại sau khi thêm nội dung
	}
	pdf.Ln(10)

	// // **Section 3: Tuyến Đường Phổ Biến**
	// pdf.AddPage()
	// pdf.SetFont("DejaVu", "B", 14)
	// pdf.Cell(0, 10, "3. Tuyến Đường Phổ Biến")
	// pdf.Ln(12)

	// // Table header
	// pdf.SetFillColor(200, 200, 200)
	// pdf.SetFont("DejaVu", "B", 12)
	// pdf.CellFormat(60, 10, "Địa chỉ bắt đầu", "1", 0, "C", true, 0, "")
	// pdf.CellFormat(60, 10, "Địa chỉ kết thúc", "1", 0, "C", true, 0, "")
	// pdf.CellFormat(30, 10, "Số lượt", "1", 1, "C", true, 0, "")

	// // Table data
	// pdf.SetFont("DejaVu", "", 12)
	// for _, route := range data.PopularRoutes {
	// 	checkAndAddPage()
	// 	pdf.CellFormat(60, 10, route.StartAddress, "1", 0, "", false, 0, "")
	// 	pdf.CellFormat(60, 10, route.EndAddress, "1", 0, "", false, 0, "")
	// 	pdf.CellFormat(30, 10, fmt.Sprintf("%d", route.Count), "1", 1, "", false, 0, "")
	// }
	// pdf.Ln(10)

	// **Section 3: Tuyến Đường Phổ Biến**
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 10, "3. Tuyến Đường Phổ Biến")
	pdf.Ln(12)

	// Định nghĩa chiều rộng cột
	colWidth := []float64{70, 70, 30}

	// Hàm để cắt ngắn và thêm dấu "..." nếu văn bản quá dài
	truncateText := func(text string, width float64, fontSize float64) string {
		pdf.SetFont("DejaVu", "", fontSize)
		if pdf.GetStringWidth(text) > width {
			for len(text) > 0 {
				text = text[:len(text)-1]
				if pdf.GetStringWidth(text+"...") <= width {
					return text + "..."
				}
			}
		}
		return text
	}

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("DejaVu", "B", 10)
	headers := []string{"Địa chỉ bắt đầu", "Địa chỉ kết thúc", "Số lượt"}
	for i, header := range headers {
		pdf.CellFormat(colWidth[i], 10, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("DejaVu", "", 9)
	for _, route := range data.PopularRoutes {
		startX := pdf.GetX()
		startY := pdf.GetY()

		// Kiểm tra nếu cần thêm trang mới
		if startY > 250 {
			pdf.AddPage()
			pdf.SetFont("DejaVu", "B", 10)
			for i, header := range headers {
				pdf.CellFormat(colWidth[i], 10, header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("DejaVu", "", 9)
			startY = pdf.GetY()
			startX = pdf.GetX()
		}

		// Cắt ngắn nội dung nếu cần
		startAddress := truncateText(route.StartAddress, colWidth[0]-2, 9)
		endAddress := truncateText(route.EndAddress, colWidth[1]-2, 9)

		// Tính chiều cao cần thiết cho mỗi ô
		pdf.SetFont("DejaVu", "", 9)
		startHeight := pdf.SplitLines([]byte(startAddress), colWidth[0]-2)
		endHeight := pdf.SplitLines([]byte(endAddress), colWidth[1]-2)
		maxHeight := math.Max(float64(len(startHeight)), float64(len(endHeight))) * 5 // 5 là chiều cao của mỗi dòng

		// Vẽ ô và điền nội dung
		pdf.Rect(startX, startY, colWidth[0], maxHeight, "D")
		pdf.MultiCell(colWidth[0], 5, startAddress, "", "", false)
		pdf.SetXY(startX+colWidth[0], startY)

		pdf.Rect(startX+colWidth[0], startY, colWidth[1], maxHeight, "D")
		pdf.MultiCell(colWidth[1], 5, endAddress, "", "", false)

		pdf.SetXY(startX+colWidth[0]+colWidth[1], startY)
		pdf.Rect(startX+colWidth[0]+colWidth[1], startY, colWidth[2], maxHeight, "D")
		pdf.CellFormat(colWidth[2], maxHeight, fmt.Sprintf("%d", route.Count), "", 0, "C", false, 0, "")

		// Đặt lại vị trí cho hàng tiếp theo
		pdf.SetXY(startX, startY+maxHeight)
	}
	pdf.Ln(10)

	// **Section 4: Phân Bố Loại Xe**
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(0, 10, "4. Phân Bố Loại Xe")
	pdf.Ln(12)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("DejaVu", "B", 12)
	pdf.CellFormat(70, 10, "Loại xe", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, "Số lượng", "1", 1, "C", true, 0, "")

	// Table data
	pdf.SetFont("DejaVu", "", 12)
	for _, vehicle := range data.VehicleTypeDistribution {
		checkAndAddPage()
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
