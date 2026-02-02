package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeoutContext(t *testing.T) {
	app := fiber.New()
	app.Get("/timeout", func(c *fiber.Ctx) error {
		ctx, cancel := TimeoutContext(c, 2*time.Second)
		defer cancel()
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.True(t, time.Until(deadline) <= 2*time.Second)
		return c.SendString("ok")
	})

	res, err := app.Test(httptest.NewRequest("GET", "/timeout", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)
}

func TestTimeoutContext_CancelStopsContext(t *testing.T) {
	app := fiber.New()
	done := make(chan struct{})
	app.Get("/cancel", func(c *fiber.Ctx) error {
		ctx, cancel := TimeoutContext(c, time.Hour)
		cancel()
		select {
		case <-ctx.Done():
			close(done)
			return c.SendString("cancelled")
		case <-time.After(100 * time.Millisecond):
			t.Error("context should be done after cancel")
			return c.Status(500).SendString("timeout")
		}
	})

	res, err := app.Test(httptest.NewRequest("GET", "/cancel", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)
	<-done
}

type limitOffsetBody struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func TestLimitOffset_Defaults(t *testing.T) {
	app := fiber.New()
	app.Get("/list", func(c *fiber.Ctx) error {
		limit, offset := LimitOffset(c, 20, 100, 0)
		return c.JSON(fiber.Map{"limit": limit, "offset": offset})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/list", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var body limitOffsetBody
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, 20, body.Limit)
	assert.Equal(t, 0, body.Offset)
}

func TestLimitOffset_QueryParams(t *testing.T) {
	app := fiber.New()
	app.Get("/list", func(c *fiber.Ctx) error {
		limit, offset := LimitOffset(c, 20, 100, 0)
		return c.JSON(fiber.Map{"limit": limit, "offset": offset})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/list?limit=50&offset=10", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var body limitOffsetBody
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, 50, body.Limit)
	assert.Equal(t, 10, body.Offset)
}

func TestLimitOffset_LimitCappedByMax(t *testing.T) {
	app := fiber.New()
	app.Get("/list", func(c *fiber.Ctx) error {
		limit, offset := LimitOffset(c, 20, 100, 0)
		return c.JSON(fiber.Map{"limit": limit, "offset": offset})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/list?limit=500&offset=0", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var body limitOffsetBody
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, 100, body.Limit)
}

func TestLimitOffset_NegativeLimitAndOffsetUseDefaults(t *testing.T) {
	app := fiber.New()
	app.Get("/list", func(c *fiber.Ctx) error {
		limit, offset := LimitOffset(c, 20, 100, 0)
		return c.JSON(fiber.Map{"limit": limit, "offset": offset})
	})

	res, err := app.Test(httptest.NewRequest("GET", "/list?limit=-1&offset=-5", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, res.StatusCode)

	var body limitOffsetBody
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	assert.Equal(t, 20, body.Limit)
	assert.Equal(t, 0, body.Offset)
}
