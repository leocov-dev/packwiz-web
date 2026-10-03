package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"packwiz-web/internal/params"
	"packwiz-web/internal/services/audit_svc"
	"packwiz-web/internal/services/authz_svc"
	"packwiz-web/internal/services/user_svc"
	"packwiz-web/internal/tables"
	"packwiz-web/internal/types/dto"
	"packwiz-web/internal/types/response"
)

type AdminController struct {
	db       *gorm.DB
	svc      *user_svc.UserService
	authzSvc *authz_svc.Service
	auditSvc *audit_svc.AuditService
}

func NewAdminController(db *gorm.DB) *AdminController {
	return &AdminController{
		db:       db,
		svc:      user_svc.NewUserService(db),
		authzSvc: authz_svc.NewService(db),
		auditSvc: audit_svc.NewAuditService(db),
	}
}

func (uc *AdminController) GetUsersPaginated(c *gin.Context) {
	var query dto.ListUsersQuery
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	users, total, err := uc.svc.ListUsers(query)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, response.NewPaginated(
		users,
		query.Page,
		query.PageSize,
		total,
	))
}

func (uc *AdminController) GetAuditsPaginated(c *gin.Context) {
	var query dto.ListAuditsQuery
	if err := mustBindQuery(c, &query); err != nil {
		err.JSON(c)
		return
	}

	audits, total, err := uc.auditSvc.ListAudits(query)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, response.NewPaginated(
		audits,
		query.Page,
		query.PageSize,
		total,
	))
}

func (uc *AdminController) GetUserById(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	user, dbErr := uc.svc.FindById(userId)
	if dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			response.New(http.StatusNotFound, fmt.Sprintf("user %d not found", userId)).JSON(c)
			return
		}
		response.New(http.StatusInternalServerError, "failed to fetch user").JSON(c)
		return
	}

	withRoles, rolesErr := uc.svc.WithRoles([]tables.User{user})
	if rolesErr != nil {
		response.Wrap(rolesErr).JSON(c)
		return
	}

	dataOK(c, withRoles[0])
}

func (uc *AdminController) CreateUser(c *gin.Context) {
	var request dto.CreateUserRequest
	if err := mustBindJson(c, &request); err != nil {
		err.JSON(c)
		return
	}

	user, err := uc.svc.CreateUser(request)
	if err != nil {
		err.JSON(c)
		return
	}

	dataOK(c, user)
}

func (uc *AdminController) UpdateUser(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	var request dto.EditUserRequest
	if err := mustBindJson(c, &request); err != nil {
		err.JSON(c)
		return
	}

	if err := uc.svc.UpdateUser(userId, request); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

func (uc *AdminController) DeactivateUser(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	actingUser := c.MustGet("user").(tables.User)

	if err := uc.svc.DeactivateUser(actingUser, userId); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

func (uc *AdminController) ReactivateUser(c *gin.Context) {
	userId, err := mustBindIdParam(c, params.UserID)
	if err != nil {
		err.JSON(c)
		return
	}

	if err := uc.svc.ReactivateUser(userId); err != nil {
		err.JSON(c)
		return
	}

	isOK(c)
}

func (uc *AdminController) ListRoles(c *gin.Context) {
	roles, err := uc.authzSvc.ListRoles(c.Query("scope"))
	if err != nil {
		response.Wrap(err).JSON(c)
		return
	}

	dataOK(c, gin.H{"roles": roles})
}

func (uc *AdminController) ListPermissions(c *gin.Context) {
	perms, err := uc.authzSvc.ListPermissions()
	if err != nil {
		response.Wrap(err).JSON(c)
		return
	}

	dataOK(c, gin.H{"permissions": perms})
}

func (uc *AdminController) SetUserRoles(c *gin.Context) {
	userId, bindErr := mustBindIdParam(c, params.UserID)
	if bindErr != nil {
		bindErr.JSON(c)
		return
	}

	var request dto.SetUserRolesRequest
	if bindErr := mustBindJson(c, &request); bindErr != nil {
		bindErr.JSON(c)
		return
	}

	switch err := uc.authzSvc.SetUserRoles(userId, request.RoleIDs); {
	case err == nil:
		isOK(c)
	case errors.Is(err, authz_svc.ErrUserNotFound):
		response.New(http.StatusNotFound, fmt.Sprintf("user %d not found", userId)).JSON(c)
	case errors.Is(err, authz_svc.ErrSuperuserProtected):
		response.New(http.StatusForbidden, err.Error()).JSON(c)
	case errors.Is(err, authz_svc.ErrInvalidRole):
		response.New(http.StatusBadRequest, "one or more roles are invalid or not assignable").JSON(c)
	default:
		response.Wrap(err).JSON(c)
	}
}
