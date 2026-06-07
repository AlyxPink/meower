package handlers

import (
	"fmt"
	"html"

	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"

	"TEMPLATE_MODULE_PATH/web/sse"
	"TEMPLATE_MODULE_PATH/web/views"

	"github.com/gofiber/fiber/v2"
)

type Meower struct{ *App }

func (h *Meower) New(c *fiber.Ctx) error {
	return renderTempl(c, views.NewMeow(c))
}

func (h *Meower) Create(c *fiber.Ctx) error {
	content := c.FormValue("content")
	req := &meowV1.CreateMeowRequest{Content: content}

	resp, err := h.API.MeowService.CreateMeow(c.Context(), req)
	if err != nil {
		return err
	}

	// Demo of the SSE hub: publish the new meow to the "meows" channel so every
	// browser subscribed at /events/stream?channels=meows live-appends it. The
	// event Data is the HTML fragment the browser inserts into the list.
	if h.Hub != nil {
		fragment := fmt.Sprintf(
			`<li class="py-2 rounded bg-pink-200 p-2 my-4"><p class="font-bold">%s</p></li>`,
			html.EscapeString(resp.Meow.Content),
		)
		h.Hub.Broadcast("meows", sse.SSEEvent{
			Event: "meow-created",
			ID:    resp.Meow.Id,
			Data:  fragment,
		})
	}

	return renderTempl(c, views.CreateMeow(c, resp))
}

func (h *Meower) Index(c *fiber.Ctx) error {
	req := &meowV1.IndexMeowRequest{}

	resp, err := h.API.MeowService.IndexMeow(c.Context(), req)
	if err != nil {
		return err
	}

	return renderTempl(c, views.IndexMeows(c, resp))
}
