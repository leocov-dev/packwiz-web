package controllers

import (
	"context"

	"github.com/gin-gonic/gin"

	"packwiz-web/internal/params"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

// packWebhookService is what PackWebhookController needs from the webhook service.
type packWebhookService interface {
	Get(ctx context.Context, packID uint) (dto.PackWebhookResponse, response.ServerError)
	Save(ctx context.Context, packID uint, form dto.PackWebhookForm) (dto.PackWebhookResponse, response.ServerError)
	Test(ctx context.Context, packID uint, form dto.PackWebhookForm) (dto.PackWebhookTestResponse, response.ServerError)
}

// PackWebhookController serves the per-pack webhook configuration.
type PackWebhookController struct {
	svc packWebhookService
}

func NewPackWebhookController(svc packWebhookService) *PackWebhookController {
	return &PackWebhookController{svc: svc}
}

func (wc *PackWebhookController) GetWebhook(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	hook, err := wc.svc.Get(c.Request.Context(), packId)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, hook)
}

func (wc *PackWebhookController) SaveWebhook(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	var form dto.PackWebhookForm
	if err := mustBindJson(c, &form); err != nil {
		err.JSON(c)
		return
	}

	hook, err := wc.svc.Save(c.Request.Context(), packId, form)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, hook)
}

func (wc *PackWebhookController) TestWebhook(c *gin.Context) {
	packId, err := mustBindIdParam(c, params.PackId)
	if err != nil {
		err.JSON(c)
		return
	}

	var form dto.PackWebhookForm
	if err := mustBindJson(c, &form); err != nil {
		err.JSON(c)
		return
	}

	result, err := wc.svc.Test(c.Request.Context(), packId, form)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, result)
}
