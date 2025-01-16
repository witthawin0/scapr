
package handler

import (
	"github.com/gofiber/fiber/v2"
	"{{ .ProjectName }}/internal/service"
)

type {{ .StructName }}Handler struct {
	*Handler
	{{ .StructNameLowerFirst }}Service service.{{ .StructName }}Service
}

func New{{ .StructName }}Handler(
    handler *Handler,
    {{ .StructNameLowerFirst }}Service service.{{ .StructName }}Service,
) *{{ .StructName }}Handler {
	return &{{ .StructName }}Handler{
		Handler:      handler,
		{{ .StructNameLowerFirst }}Service: {{ .StructNameLowerFirst }}Service,
	}
}

func (h *{{ .StructName }}Handler) Get{{ .StructName }}(c *fiber.Ctx) error {
	panic("Not implemented")
}