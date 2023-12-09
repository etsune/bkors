package main

import "github.com/labstack/echo/v4"

func router(e *echo.Echo, h *AppHandler) {
	e.GET("/", h.IndexPageHandler) // IndexPage

	e.GET("/search", h.SearchPageHandler)               // Search
	e.GET("/download", h.DownloadPageHandler)           // DownloadPage
	e.GET("/page/:dict/:vol/:page", h.SheetPageHandler) // DictionaryPage
	e.GET("/page/:dict/:num", h.SheetNextHandler)       // DictionaryPage
	e.GET("/page-export/:vol/:page", h.ExportPageToTxt)
	// POST - get search component

	e.GET("/login", h.LoginPageHandler)
	e.POST("/login", h.LoginRequestHandler)
	e.GET("/logout", h.LogoutRequestHandler)

	e.POST("/entries/:id/edits", h.CreateEdit)
	e.GET("/edits", h.EditsPageHandler) // EditsPage
	e.GET("/edits-pending", h.EditsPendingPageHandler)

	e.POST("/edits/:id/approve", h.ApproveEdit)
	e.POST("/edits/:id/decline", h.DeclineEdit)

	// Login Discord
	// Login Github

	// GET editor component

	// Download Export etc
	// e.GET("/export", h.ExportRequestHandler)
	// e.POST("/import", h.ImportListHandler)
}
