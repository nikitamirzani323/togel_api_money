package controller

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/nikitamirzani323/togel_api_money/entities"
)

var ctx = context.Background()

func Fetch_token(c *fiber.Ctx) error {
	client := new(entities.Controller_clientToken)

	if err := c.BodyParser(client); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"status":  fiber.StatusBadRequest,
			"message": err.Error(),
			"record":  nil,
		})
	}

	member_username := ""
	member_company := ""
	member_saldo := 0
	switch client.Token {
	case "qC5YmBvXzabGp34jJlKvnC6wCrr3pLCwBzsLoSzl4k=":
		member_username = "developer"
		member_company = "MMD"
		member_saldo = 5000000
	case "qwertyuiop":
		member_username = "Jhon Wick"
		member_company = "MMD"
		member_saldo = 1000000
	case "asdfghjkl":
		member_username = "Edo Febrian"
		member_company = "MMD"
		member_saldo = 200000
	case "1234567890":
		member_username = "antonnukue"
		member_company = "NUK"
		member_saldo = 1000000
	case "0987654321":
		member_username = "developernuke"
		member_company = "NUK"
		member_saldo = 2000000
	}
	return c.JSON(fiber.Map{
		"status":          fiber.StatusOK,
		"token":           client.Token,
		"member_username": member_username,
		"member_company":  member_company,
		"member_credit":   member_saldo,
	})
}
