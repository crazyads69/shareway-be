package controller

import (
	"shareway/helper"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type MapController struct {
	MapsService service.IMapService
	validate    *validator.Validate
}

func NewMapController(mapsService service.IMapService, validate *validator.Validate) *MapController {
	return &MapController{
		MapsService: mapsService,
		validate:    validate,
	}
}

// GetAutoComplete returns a list of places that match the query string
// GetAutoComplete godoc
// @Summary Get autocomplete suggestions for places
// @Description Returns a list of places that match the query string
// @Tags map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer <access_token>"
// @Param input query string true "Input string to search for"
// @Param limit query int false "Limit the number of results"
// @Param location query string false "Location coordinates (lat,lng)"
// @Param radius query int false "Search radius in meters"
// @Param more_compound query bool false "Include more compound results"
// @Success 200 {object} helper.Response "Successfully retrieved autocomplete data"
// @Failure 400 {object} helper.Response "Invalid request query"
// @Failure 500 {object} helper.Response "Failed to get autocomplete data"
// @Router /maps/autocomplete [get]
func (ctrl *MapController) GetAutoComplete(ctx *gin.Context) {

	var req schemas.AutoCompleteRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Invalid request query",
			"Câu truy vấn không hợp lệ",
		)
		helper.GinResponse(ctx, 400, response)
		return
	}

	places, err := ctrl.MapsService.GetAutoComplete(ctx.Request.Context(), req.Input, req.Limit, req.Location, req.Radius, req.MoreCompound)
	if err != nil {
		response := helper.ErrorResponseWithMessage(
			err,
			"Failed to get autocomplete data",
			"Không thể lấy dữ liệu gợi ý",
		)
		helper.GinResponse(ctx, 500, response)
		return
	}

	response := helper.SuccessResponse(
		places,
		"Successfully retrieved autocomplete data",
		"Lấy dữ liệu gợi ý thành công",
	)
	helper.GinResponse(ctx, 200, response)
}
