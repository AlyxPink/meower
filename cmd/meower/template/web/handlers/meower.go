package handlers

import (
	"bytes"
	"context"

	meowV1 "TEMPLATE_MODULE_PATH/api/proto/meow/v1"
	"TEMPLATE_MODULE_PATH/pkg/urls"
	"TEMPLATE_MODULE_PATH/web/sse"
	"TEMPLATE_MODULE_PATH/web/views"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
)

type Meower struct{ *App }

// meowsChannel is the SSE channel the timeline subscribes to. Publishing a meow
// event here live-updates every connected browser viewing /meows.
const meowsChannel = "meows"

// currentUserID returns the logged-in user's id from the request locals (set by
// AuthMiddleware), or "" if not available.
func currentUserID(c *fiber.Ctx) string {
	if v, ok := c.Locals("user_id").(string); ok {
		return v
	}
	return ""
}

// renderFragment renders a templ component to an HTML string for embedding in an
// SSE event. It uses the request's user context (which carries the active trace
// span under otelfiber) so cancellation and trace context propagate.
func renderFragment(c *fiber.Ctx, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(c.UserContext(), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// publish broadcasts an SSE event built from a templ fragment to the meows
// channel. Broadcasting is best-effort: a render failure is returned, but a
// missing hub (e.g. in tests) is silently ignored.
func (h *Meower) publish(c *fiber.Ctx, event, id string, fragment templ.Component) error {
	if h.Hub == nil {
		return nil
	}
	data, err := renderFragment(c, fragment)
	if err != nil {
		return err
	}
	h.Hub.Broadcast(meowsChannel, sse.SSEEvent{Event: event, ID: id, Data: data})
	return nil
}

// New renders the standalone composer page.
func (h *Meower) New(c *fiber.Ctx) error {
	return renderTempl(c, views.NewMeow(c))
}

// Create posts a new meow authored by the logged-in user, broadcasts it to the
// timeline, and redirects back to the feed (where the live card is already in
// place for the author too, via the SSE connection).
func (h *Meower) Create(c *fiber.Ctx) error {
	content := c.FormValue("content")
	if content == "" {
		return c.Redirect(urls.MeowIndex{}.URL())
	}

	resp, err := h.API.MeowService.CreateMeow(c.UserContext(), &meowV1.CreateMeowRequest{
		Content:  content,
		AuthorId: currentUserID(c),
	})
	if err != nil {
		return err
	}

	if err := h.publish(c, "meow-created", resp.Meow.Id, views.MeowCreatedFragment(resp.Meow)); err != nil {
		return err
	}

	return c.Redirect(urls.MeowIndex{}.URL())
}

// Index renders the home timeline.
func (h *Meower) Index(c *fiber.Ctx) error {
	resp, err := h.API.MeowService.IndexMeow(c.UserContext(), &meowV1.IndexMeowRequest{})
	if err != nil {
		return err
	}

	return renderTempl(c, views.IndexMeows(c, resp, currentUserID(c)))
}

// Show renders the permalink page for a single meow.
func (h *Meower) Show(c *fiber.Ctx) error {
	resp, err := h.getMeow(c.UserContext(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "meow not found")
	}

	canManage := resp.Meow.AuthorId == currentUserID(c) && currentUserID(c) != ""
	return renderTempl(c, views.ShowMeow(c, resp.Meow, canManage))
}

// Edit renders the edit form. Only the author may edit.
func (h *Meower) Edit(c *fiber.Ctx) error {
	resp, err := h.getMeow(c.UserContext(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "meow not found")
	}

	if resp.Meow.AuthorId != currentUserID(c) || currentUserID(c) == "" {
		return fiber.NewError(fiber.StatusForbidden, "you can only edit your own meows")
	}

	return renderTempl(c, views.EditMeow(c, resp.Meow))
}

// Update saves an edit, broadcasts the change to the timeline, and redirects to
// the meow's permalink. Only the author may update.
func (h *Meower) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	existing, err := h.getMeow(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "meow not found")
	}
	if existing.Meow.AuthorId != currentUserID(c) || currentUserID(c) == "" {
		return fiber.NewError(fiber.StatusForbidden, "you can only edit your own meows")
	}

	content := c.FormValue("content")
	if content == "" {
		return c.Redirect(urls.MeowEdit{ID: id}.URL())
	}

	resp, err := h.API.MeowService.UpdateMeow(c.UserContext(), &meowV1.UpdateMeowRequest{
		Id:      id,
		Content: content,
	})
	if err != nil {
		return err
	}

	if err := h.publish(c, "meow-updated", resp.Meow.Id, views.MeowUpdatedFragment(resp.Meow)); err != nil {
		return err
	}

	return c.Redirect(urls.Meow{ID: id}.URL())
}

// Delete removes a meow, broadcasts its removal to the timeline, and redirects
// to the feed. Only the author may delete.
func (h *Meower) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	existing, err := h.getMeow(c.UserContext(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "meow not found")
	}
	if existing.Meow.AuthorId != currentUserID(c) || currentUserID(c) == "" {
		return fiber.NewError(fiber.StatusForbidden, "you can only delete your own meows")
	}

	if _, err := h.API.MeowService.DeleteMeow(c.UserContext(), &meowV1.DeleteMeowRequest{Id: id}); err != nil {
		return err
	}

	if err := h.publish(c, "meow-deleted", id, views.MeowDeletedFragment(id)); err != nil {
		return err
	}

	return c.Redirect(urls.MeowIndex{}.URL())
}

// getMeow fetches a single meow by id via the gRPC API.
func (h *Meower) getMeow(ctx context.Context, id string) (*meowV1.GetMeowResponse, error) {
	return h.API.MeowService.GetMeow(ctx, &meowV1.GetMeowRequest{Id: id})
}
