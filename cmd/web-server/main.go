package main

import (
	"github.com/AlyxPink/meower/internal/web"
)

func main() {
	app := web.NewServer()
	app.Listen("0.0.0.0:3000")
}
