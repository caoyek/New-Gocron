package loginlog

import (
	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
	"github.com/caoyek/New-Gocron/internal/modules/utils"
	"github.com/caoyek/New-Gocron/internal/routers/base"
	macaron "gopkg.in/macaron.v1"
)

func Index(ctx *macaron.Context) string {
	loginLogModel := new(models.LoginLog)
	params := models.CommonMap{}
	base.ParsePageAndPageSize(ctx, params)
	total, err := loginLogModel.Total()
	if err != nil {
		logger.Error(err)
	}
	loginLogs, err := loginLogModel.List(params)
	if err != nil {
		logger.Error(err)
	}

	jsonResp := utils.JsonResponse{}

	return jsonResp.Success(utils.SuccessContent, map[string]interface{}{
		"total": total,
		"data":  loginLogs,
	})
}
