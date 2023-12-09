package main

import (
	"context"
	"fmt"
	"github.com/etsune/bkors/server/templates/components"
	"github.com/etsune/bkors/server/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	editService  *services.EditService
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
	edits, err := h.editService.GetAll(false)
	if err != nil {
		return err
	}

	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.EditsPage(pageOptions, edits).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) EditsPendingPageHandler(c echo.Context) error {
	edits, err := h.editService.GetAll(true)
	if err != nil {
		return err
	}

	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	return pages.EditsPage(pageOptions, edits).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) SheetNextHandler(c echo.Context) error {
	dict, numPar := c.Param("dict"), c.Param("num")

	num, _ := strconv.Atoi(numPar)

	fmt.Println(num)
	sheet, err := h.sheetService.GetByNum(dict, num)
	if err != nil {
		return err
	}

	redirectUrl := fmt.Sprintf("/page/%s/%d/%d", dict, sheet.Volume, sheet.Page)

	return c.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *AppHandler) ApproveEdit(c echo.Context) error {
	user := getCtxUserdata(c)
	if user == nil || !user.IsAdmin {
		return c.String(http.StatusBadRequest, "user has no access")
	}
	editId := c.Param("id")

	err := h.editService.Approve(editId)
	if err != nil {
		return err
	}

	return components.Ok().Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) DeclineEdit(c echo.Context) error {
	user := getCtxUserdata(c)
	if user == nil || !user.IsAdmin {
		return c.String(http.StatusBadRequest, "user has no access")
	}
	editId := c.Param("id")

	err := h.editService.SetEditStatus(editId, models.StatusDeclined)
	if err != nil {
		return err
	}

	return components.Ok().Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) SheetPageHandler(c echo.Context) error {
	pageOptions := pages.NewPageOptions(getCtxUserdata(c))
	dict, volPar, pagePar := c.Param("dict"), c.Param("vol"), c.Param("page")

	vol, _ := strconv.Atoi(volPar)
	page, _ := strconv.Atoi(pagePar)

	sheet, err := h.sheetService.Get(dict, vol, page)
	if err != nil {
		return err
	}

	entries, err := h.entryService.GetEntriesForPage(sheet.Volume, sheet.Page)
	if err != nil {
		return err
	}

	return pages.SheetPage(pageOptions, sheet, entries).Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) CreateEdit(c echo.Context) error {
	edit := new(models.EditEntry)
	if err := c.Bind(edit); err != nil {
		return err
	}
	entryId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return err
	}

	var username string
	var user *models.DBUser
	if user = getCtxUserdata(c); user != nil {
		username = user.Username
	} else {
		username = utils.GetAnonymName(c.RealIP())
	}

	editId, err := h.editService.CreateEdit(edit, entryId, username)
	if err != nil {
		return c.HTML(http.StatusOK, err.Error())
	}

	if user != nil && user.HasAutoApprove {
		if err = h.editService.Approve(editId.Hex()); err != nil {
			return err
		}
	}

	return components.OkWithMsg("Правка отправлена").Render(context.Background(), c.Response().Writer)
}

func (h *AppHandler) ExportPageToTxt(c echo.Context) error {
	volPar, pagePar := c.Param("vol"), c.Param("page")

	vol, _ := strconv.Atoi(volPar)
	page, _ := strconv.Atoi(pagePar)

	res, err := h.entryService.ExportPageToTxt(vol, page)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment;filename=\"bkors-%d-%d.txt\"", vol, page))
	return c.Blob(http.StatusOK, "text/plain", res)
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
