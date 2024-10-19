package controller

import (
	"log"
	"shareway/helper"
	"shareway/schemas"
	"shareway/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type MapsController struct {
	MapsService service.IMapsService
	validate    *validator.Validate
}

func NewMapsController(mapsService service.IMapsService, validate *validator.Validate) *MapsController {
	return &MapsController{
		MapsService: mapsService,
		validate:    validate,
	}
}

// GetAutoComplete returns a list of places that match the query string
func (ctrl *MapsController) GetAutoComplete(ctx *gin.Context) {

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

	log.Printf("Found %d places", len(places))
}
