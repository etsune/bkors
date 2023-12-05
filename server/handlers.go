package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/etsune/bkors/server/models"
	"github.com/etsune/bkors/server/services"
	"github.com/etsune/bkors/server/templates/pages"
	"github.com/labstack/echo/v4"
)

type AppHandler struct {
	entryService *services.EntryService
	authService  *services.AuthService
	userService  *services.UserService
	sheetService *services.SheetService
}

func (h *AppHandler) IndexPageHandler(c echo.Context) error {

	// component := templates.Layout()
	// return c.String(http.StatusOK, "Hello, World!")
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.Index(pageOptions).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) LogoutRequestHandler(c echo.Context) error {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Expires: time.Now()})
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AppHandler) LoginRequestHandler(c echo.Context) error {
	cookie, err := h.userService.RegisterUser(c.FormValue("username"), c.FormValue("password"))
	if err != nil {
		return err
	}
	c.SetCookie(cookie)
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AppHandler) LoginPageHandler(c echo.Context) error {
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.Login(pageOptions).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) DownloadPageHandler(c echo.Context) error {
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.DownloadPage(pageOptions).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) EditsPageHandler(c echo.Context) error {
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.EditsPage(pageOptions).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) SheetPageHandler(c echo.Context) error {
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	dict, numPar := c.Param("dict"), c.Param("num")

	num, _ := strconv.Atoi(numPar)

	sheet, err := h.sheetService.Get(dict, num)
	if err != nil {
		return err
	}

	entries, err := h.entryService.GetEntriesForPage(sheet.Volume, sheet.Page)
	if err != nil {
		return err
	}

	return pages.SheetPage(pageOptions, sheet, entries).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) ImportListHandler(c echo.Context) error {
	user := getCtxUserdata(c)
	if user == nil || !user.IsAdmin {
		return c.String(http.StatusBadRequest, "user has no access")
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	src, err := file.Open()
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	defer src.Close()
	err = h.entryService.ImportFile(src)
	if err != nil {
		c.Logger().Error(err)
		return err
	}

	return nil
}

func (h *AppHandler) SearchPageHandler(c echo.Context) error {

	// &[]models.DBEntry{}
	// component := templates.Layout()
	// return c.String(http.StatusOK, "Hello, World!")
	term := c.FormValue("s")
	list, err := h.entryService.SearchEntries(term)
	if err != nil {
		return err
	}
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.Search(&list, term, pageOptions).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) ExportRequestHandler(c echo.Context) error {
	// user := getCtxUserdata(c)
	// if user == nil || !user.IsAdmin {
	// 	return c.String(http.StatusBadRequest, "user has no access")
	// }
	data := []byte(h.entryService.ExportEntries())
	return c.Blob(http.StatusOK, "text/csv", data)
}

func getCtxUserdata(c echo.Context) *models.DBUser {
	data, ok := c.Get("userdata").(models.DBUser)
	if !ok {
		return nil
	}
	return &data
}
