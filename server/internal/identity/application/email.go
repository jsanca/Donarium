package application

type EmailNormalizer interface {
	Normalize(email string) (string, error)
}
