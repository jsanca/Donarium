package runtime

import "github.com/go-chi/chi/v5"

type ModuleRuntime interface {
	RegisterRoutes(r chi.Router)
}
