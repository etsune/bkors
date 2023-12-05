package main

import "github.com/labstack/echo/v4"

func router(e *echo.Echo, h *AppHandler) {
	e.GET("/", h.IndexPageHandler) // IndexPage

	e.GET("/search", h.SearchPageHandler)         // Search
	e.GET("/download", h.DownloadPageHandler)     // DownloadPage
	e.GET("/edits", h.EditsPageHandler)           // EditsPage
	e.GET("/page/:dict/:num", h.SheetPageHandler) // DictionaryPage
	// POST - get search component

	e.GET("/login", h.LoginPageHandler)
	e.POST("/login", h.LoginRequestHandler)
	e.GET("/logout", h.LogoutRequestHandler)

	// Login Discord
	// Login Github

	// GET editor component

	// Edit Get
	// Edit List
	// Edit Create
	// Edit Accept
	// Edit Decline

	// Download Export etc
	// e.GET("/export", h.ExportRequestHandler)
	// e.POST("/import", h.ImportListHandler)
}
