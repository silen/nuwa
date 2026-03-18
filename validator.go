package nuwa

import (
	"github.com/gin-gonic/gin/binding"

	nuwabinding "github.com/silen/nuwa/binding"
)

// Validator returns Nuwa's default Gin binding validator.
func Validator() binding.StructValidator {
	return nuwabinding.Validator()
}
