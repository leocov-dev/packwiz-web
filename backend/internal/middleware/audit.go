package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/logger"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
)

func ApiAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		actionParams := make(map[string]interface{})
		auditRecord := &tables.Audit{
			IpAddress: c.ClientIP(),
		}

		if c.Params != nil {
			actionParams["params"] = paramsToMap(c.Params)
		}
		if c.Request.URL.RawQuery != "" {
			actionParams["query"] = c.Request.URL.Query()
		}

		if c.Request.Header.Get("Content-Type") == "application/json" {
			var bodyJson map[string]interface{}
			if err := c.ShouldBindBodyWithJSON(&bodyJson); err == nil {
				actionParams["body"] = bodyJson
			}
		}

		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) != 0 {
				formCopy := utils.DeepCopyMapStringSlice(c.Request.Form)
				if _, ok := formCopy["password"]; ok {
					formCopy["password"] = []string{"********"}
				}
			}
		}

		c.Next()

		// api auth middleware is bound after this one so this needs to be after
		// the call to c.Next()
		if action := c.GetString("auditAction"); action != "" {
			auditRecord.Action = action
		} else {
			auditRecord.Action = c.Request.Method + " " + c.FullPath()
		}

		// api auth middleware is bound after this one so this needs to be after
		// the call to c.Next()
		if user, ok := c.Get("user"); ok {
			auditRecord.UserId = user.(tables.User).Id
			actionParams["user"] = user.(tables.User).Username
		}

		// TODO: do we want to detect all requests? intrusion detection? or just
		// valid requests that might actually do something?
		if auditRecord.UserId == 0 {
			return
		}

		actionParams["code"] = c.Writer.Status()

		if jsonData, err := json.Marshal(actionParams); err == nil {
			auditRecord.ActionParams = string(jsonData)
		} else {
			logger.Error(fmt.Sprintf("Failed to marshal action params: %s", err))
		}

		if recordAsJson, err := json.Marshal(auditRecord); err == nil {
			logger.Debug("API Audit:", string(recordAsJson))
		}

		if err := db.Create(auditRecord).Error; err != nil {
			logger.Error(fmt.Sprintf("Failed to create audit record: %s", err))
		}
	}
}

func paramsToMap(params gin.Params) map[string]string {
	paramsMap := make(map[string]string)
	for _, param := range params {
		paramsMap[param.Key] = param.Value
	}
	return paramsMap
}

func PackwizAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
