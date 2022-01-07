package init

import (
	"github.com/openconfig/ondatra/fakebind"
	"github.com/openconfig/ondatra/internal/binding"
)

func Init() (binding.Binding, error) {
	return fakebind.New()
}
