package services

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/tungstenfabric-preview/intent-service/pkg/auth"
)

// RESTGetObjPerms handles GET operation of obj-perms request.
func (service *ContrailService) RESTGetObjPerms(c echo.Context) error {
	return c.JSON(http.StatusOK, auth.GetContext(c).GetObjPerms())
}
