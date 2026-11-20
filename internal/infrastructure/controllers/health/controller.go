package health

import (
	"net/http"

	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type HealthController struct {
	readiness func() bool
	validator  *validation.Validator
}

func NewHealthController(readiness ...func() bool) *HealthController {
	ready := func() bool { return true }
	if len(readiness) > 0 && readiness[0] != nil {
		ready = readiness[0]
	}
	return &HealthController{readiness: ready, validator: validation.NewValidator()}
}

func (controller *HealthController) Routes(mux *http.ServeMux) {
	mux.Handle("GET /health", httptransport.Handler((&healthEndpoint{validator: controller.validator}).Handle))
	mux.Handle("GET /ready", httptransport.Handler((&readyEndpoint{readiness: controller.readiness, validator: controller.validator}).Handle))
}
