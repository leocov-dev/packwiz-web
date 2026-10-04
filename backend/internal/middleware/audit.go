package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"packwiz-web/internal/log"
	"packwiz-web/internal/middleware/meta"
	"packwiz-web/internal/params"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/utils"
	"strings"
)

func SkipAudit(c *gin.Context) {
	c.Set("skipAudit", true)
	c.Next()
}

func ApiAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

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
				actionParams["body"] = utils.RedactSensitiveValues(bodyJson)
			}
		}

		if err := c.Request.ParseForm(); err == nil {
			if len(c.Request.Form) != 0 {
				formCopy := utils.DeepCopyMapStringSlice(c.Request.Form)
				if _, ok := formCopy["password"]; ok {
					formCopy["password"] = []string{"********"}
				}
				actionParams["form"] = formCopy
			}
		}

		c.Next()

		if c.GetBool("skipAudit") {
			return
		}

		// api auth middleware is bound after this one so this needs to be after
		// the call to c.Next()
		if tag, ok := c.Get("meta.category"); ok {
			auditRecord.Action = string(tag.(meta.TagCategory))
		} else if action := c.GetString("auditAction"); action != "" {
			auditRecord.Action = action
		} else {
			auditRecord.Action = c.Request.Method + " " + c.FullPath()
		}

		// api auth middleware is bound after this one so this needs to be after
		// the call to c.Next()
		if user, ok := c.Get("user"); ok {
			auditRecord.UserId = user.(tables.User).ID
			actionParams["user"] = user.(tables.User).Username
		}

		// TODO: do we want to detect all requests? intrusion detection? or just
		//  valid requests that might actually do something?
		if auditRecord.UserId == 0 {
			return
		}

		actionParams["code"] = c.Writer.Status()

		if jsonData, err := json.Marshal(actionParams); err == nil {
			auditRecord.ActionParams = string(jsonData)
		} else {
			log.Error(fmt.Sprintf("Failed to marshal action params: %s", err))
		}

		if recordAsJson, err := json.Marshal(auditRecord); err == nil {
			log.Debug("API Audit:", string(recordAsJson))
		}

		if err := db.Create(auditRecord).Error; err != nil {
			log.Error(fmt.Sprintf("Failed to create audit record: %s", err))
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

// PackAccessRecorder stores consumer access records. Implemented by
// pack_access_svc.PackAccessService.
type PackAccessRecorder interface {
	Record(access tables.PackAccess)
	RecordInstanceDownload(download tables.InstanceDownload)
}

const packTomlSuffix = "/pack.toml"

// instanceZipSuffix is the route pattern suffix of the instance zip.
var instanceZipSuffix = fmt.Sprintf("/%s/:%s", params.InstanceZipDir, params.InstanceFile)

// IsInstanceZipRoute reports whether the request matched the instance zip route.
func IsInstanceZipRoute(c *gin.Context) bool {
	return strings.HasSuffix(c.FullPath(), instanceZipSuffix)
}

// PackwizAudit records every request for a pack's pack.toml (as a pack access)
// and for its instance zip (as an instance download), including ones
// ConsumerAuthentication rejects, so it must be registered before it.
// ConsumerAuthentication publishes the pack and user it resolved in the
// context; both can be absent on failure.
func PackwizAudit(recorder PackAccessRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		isToml := strings.HasSuffix(c.FullPath(), packTomlSuffix)
		isZip := IsInstanceZipRoute(c)
		if !isToml && !isZip {
			return
		}

		status := c.Writer.Status()
		access := tables.PackAccess{
			PackSlug:   c.Param(string(params.PackSlug)),
			IpAddress:  c.ClientIP(),
			UserAgent:  truncate(c.Request.UserAgent(), 512),
			StatusCode: status,
			Success:    status >= 200 && status < 300,
		}
		if pack, ok := c.Get(ConsumerPackKey); ok {
			id := pack.(tables.Pack).ID
			access.PackId = &id
		}
		if user, ok := c.Get(ConsumerUserKey); ok {
			id := user.(tables.User).ID
			access.UserId = &id
		}

		if isToml {
			recorder.Record(access)
			return
		}
		recorder.RecordInstanceDownload(tables.InstanceDownload(access))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
