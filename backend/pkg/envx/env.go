package envx

type Mode string

const (
	Test  Mode = "test"
	Local Mode = "local"
	Dev   Mode = "dev"
	Prod  Mode = "prod"
)
