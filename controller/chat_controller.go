package controller

import (
	"net/http"
	"shareway/service"

	"github.com/go-playground/validator/v10"
)

type ChatController struct {
	validate    *validator.Validate
	ChatService service.IChatService
}

func NewChatController(validate *validator.Validate, chatService service.IChatService) *ChatController {
	return &ChatController{
		validate:    validate,
		ChatService: chatService,
	}
}

func (cc *ChatController) HandleWS(w http.ResponseWriter, r *http.Request) {
	cc.ChatService.HandleWSService(w, r)
}
