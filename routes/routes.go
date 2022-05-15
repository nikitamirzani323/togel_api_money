package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/nikitamirzani323/togel_api_money/controller"
)

const idleTimeout = 5 * time.Second

func Init() *fiber.App {
	app := fiber.New(fiber.Config{
		IdleTimeout: idleTimeout,
	})

	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/"
		},
		Format: "${time} | ${status} | ${latency} | ${ips[0]} | ${method} | ${path} - ${queryParams} ${body}\n",
	}))
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(requestid.New())
	app.Use(limiter.New(limiter.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		Max:        20,
		Expiration: 20 * time.Second,
	}))
	app.Use(etag.New())
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("refresh") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))
	app.Use(func(c *fiber.Ctx) error {

		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Download-Options", "noopen")
		c.Set("Strict-Transport-Security", "max-age=5184000")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-DNS-Prefetch-Control", "off")

		return c.Next()
	})

	app.Post("/api/servicetoken", controller.Fetch_token)
	return app
}
