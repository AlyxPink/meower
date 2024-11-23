package web

func main() {
	app := NewServer()
	app.Listen("0.0.0.0:3000")
}
