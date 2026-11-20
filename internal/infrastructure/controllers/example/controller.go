package example

import (
	"net/http"

	"github.com/edgarpezoa/go-template/internal/application/use-cases/example"
	"github.com/edgarpezoa/go-template/internal/config"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type ExampleController struct {
	useCases  exampleUseCases
	endpoints exampleEndpoints
	validator *validation.Validator
}

type exampleUseCases struct {
	getCache        *example.GetCacheUseCase
	setCache        *example.SetCacheUseCase
	getCacheHash    *example.GetCacheHashUseCase
	setCacheHash    *example.SetCacheHashUseCase
	sendQueue       *example.SendQueueUseCase
	sendTopic       *example.SendTopicUseCase
	insertDatabase  *example.InsertDatabaseUseCase
	findDatabase     *example.FindDatabaseUseCase
	probeValidation *example.ProbeValidationUseCase
}

type exampleEndpoints struct {
	getCache        *GetCacheEndpoint
	setCache        *SetCacheEndpoint
	getCacheHash    *GetCacheHashEndpoint
	setCacheHash    *SetCacheHashEndpoint
	sendQueue       *SendQueueEndpoint
	sendTopic       *SendTopicEndpoint
	insertDatabase  *InsertDatabaseEndpoint
	findDatabase     *FindDatabaseEndpoint
	probeValidation *ProbeValidationEndpoint
}

func NewExampleController(redis ports.RedisPort, serviceBus ports.ServiceBusPort, database ports.DatabasePort, configuration config.ServiceBusConfig) (*ExampleController, error) {
	controller := &ExampleController{validator: validation.NewValidator()}
	if err := controller.wireUseCases(redis, serviceBus, database, configuration); err != nil {
		return nil, err
	}
	controller.wireEndpoints()

	return controller, nil
}

func (controller *ExampleController) wireUseCases(redis ports.RedisPort, serviceBus ports.ServiceBusPort, database ports.DatabasePort, configuration config.ServiceBusConfig) error {
	controller.useCases.getCache = example.NewGetCacheUseCase(redis)
	controller.useCases.setCache = example.NewSetCacheUseCase(redis)
	controller.useCases.getCacheHash = example.NewGetCacheHashUseCase(redis)
	controller.useCases.setCacheHash = example.NewSetCacheHashUseCase(redis)
	sendQueue, err := example.NewSendQueueUseCase(serviceBus, configuration.Queue)
	if err != nil {
		return err
	}
	controller.useCases.sendQueue = sendQueue
	sendTopic, err := example.NewSendTopicUseCase(serviceBus, configuration.Topic)
	if err != nil {
		return err
	}
	controller.useCases.sendTopic = sendTopic
	controller.useCases.insertDatabase = example.NewInsertDatabaseUseCase(database)
	controller.useCases.findDatabase = example.NewFindDatabaseUseCase(database)
	controller.useCases.probeValidation = example.NewProbeValidationUseCase()
	return nil
}

func (controller *ExampleController) wireEndpoints() {
	controller.endpoints.getCache = &GetCacheEndpoint{execute: controller.useCases.getCache.Execute, validator: controller.validator}
	controller.endpoints.setCache = &SetCacheEndpoint{execute: controller.useCases.setCache.Execute, validator: controller.validator}
	controller.endpoints.getCacheHash = &GetCacheHashEndpoint{execute: controller.useCases.getCacheHash.Execute, validator: controller.validator}
	controller.endpoints.setCacheHash = &SetCacheHashEndpoint{execute: controller.useCases.setCacheHash.Execute, validator: controller.validator}
	controller.endpoints.sendQueue = &SendQueueEndpoint{execute: controller.useCases.sendQueue.Execute, validator: controller.validator}
	controller.endpoints.sendTopic = &SendTopicEndpoint{execute: controller.useCases.sendTopic.Execute, validator: controller.validator}
	controller.endpoints.insertDatabase = &InsertDatabaseEndpoint{execute: controller.useCases.insertDatabase.Execute, validator: controller.validator}
	controller.endpoints.findDatabase = &FindDatabaseEndpoint{execute: controller.useCases.findDatabase.Execute, validator: controller.validator}
	controller.endpoints.probeValidation = &ProbeValidationEndpoint{execute: controller.useCases.probeValidation.Execute, validator: controller.validator}
}

func (controller *ExampleController) Routes(mux *http.ServeMux) {
	mux.Handle("GET /cache/{key}", httptransport.Handler(controller.endpoints.getCache.Handle))
	mux.Handle("PUT /cache/{key}", httptransport.Handler(controller.endpoints.setCache.Handle))
	mux.Handle("GET /cache/hash/{key}", httptransport.Handler(controller.endpoints.getCacheHash.Handle))
	mux.Handle("PUT /cache/hash/{key}", httptransport.Handler(controller.endpoints.setCacheHash.Handle))
	mux.Handle("POST /service-bus/queue/messages", httptransport.Handler(controller.endpoints.sendQueue.Handle))
	mux.Handle("POST /service-bus/topic/messages", httptransport.Handler(controller.endpoints.sendTopic.Handle))
	mux.Handle("POST /validation/probe", httptransport.Handler(controller.endpoints.probeValidation.Handle))
	mux.Handle("POST /database/examples", httptransport.Handler(controller.endpoints.insertDatabase.Handle))
	mux.Handle("GET /database/examples/{id}", httptransport.Handler(controller.endpoints.findDatabase.Handle))
}
