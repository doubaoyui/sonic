package api

import (
	"github.com/gin-gonic/gin"

	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type CommentHandler struct {
	OptionService      service.OptionService
	BaseCommentService service.BaseCommentService
}

func NewCommentHandler(optionService service.OptionService, baseCommentService service.BaseCommentService) *CommentHandler {
	return &CommentHandler{
		OptionService:      optionService,
		BaseCommentService: baseCommentService,
	}
}

func (c *CommentHandler) Like(ctx *gin.Context) (interface{}, error) {
	enabled, _ := c.OptionService.GetOrByDefault(ctx, property.CommentAPIEnabled).(bool)
	if !enabled {
		return nil, xerr.WithStatus(xerr.NoType.New("comment api disabled"), xerr.StatusNotFound).WithMsg("Not Found")
	}
	commentID, err := util.ParamInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	return nil, c.BaseCommentService.IncreaseLike(ctx, commentID)
}
