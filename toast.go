package main

import (
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Toast provides methods for showing toast notifications in the frontend
type Toast struct {
	app *App
}

// NewToast creates a new Toast helper
func NewToast(app *App) *Toast {
	return &Toast{app: app}
}

// Success shows a success toast
func (t *Toast) Success(message string) {
	runtime.EventsEmit(t.app.ctx, "toast:success", message)
}

// Error shows an error toast
func (t *Toast) Error(message string) {
	runtime.EventsEmit(t.app.ctx, "toast:error", message)
}

// Warning shows a warning toast
func (t *Toast) Warning(message string) {
	runtime.EventsEmit(t.app.ctx, "toast:warning", message)
}

// Info shows an info toast
func (t *Toast) Info(message string) {
	runtime.EventsEmit(t.app.ctx, "toast:info", message)
}

// Show shows a toast with a custom type
func (t *Toast) Show(message string, toastType string) {
	runtime.EventsEmit(t.app.ctx, "toast:"+toastType, message)
}
