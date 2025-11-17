// Package events provides application-agnostic helpers for wiring
// Watermill publishers, subscribers, and middlewares.  Create a
// Service, register your handlers, and finally start the router.
//
// Basic usage:
//
//	ctx := context.Background()
//	deps := events.ServiceDependencies{
//		Validator: yourValidator,
//		Outbox:    yourOutbox,
//	}
//	svc := events.NewService(cfg, logger, ctx, deps)
//	_ = svc.RegisterHandler(events.HandlerRegistration{
//		Name:         "process_signup",
//		ConsumeQueue: cfg.ConsumeQueueSignup,
//		PublishQueue: cfg.PublishQueueSignup,
//		Handler:      yourHandler,
//	})
//
// You can tweak the middleware chain by appending to
// ServiceDependencies.Middlewares or by replacing it entirely with
// DisableDefaultMiddlewares.
//
// When processing protobuf events, you can rely on RegisterProtoHandler to
// automatically unmarshal payloads into typed messages and emit proto-based
// responses without hand-written boilerplate:
//
//	_ = events.RegisterProtoHandler(svc, events.ProtoHandlerRegistration[*domain.Signup]{
//		Name:         "signup_to_billing",
//		ConsumeQueue: cfg.ConsumeQueueSignup,
//		PublishQueue: cfg.PublishQueueSignup,
//		MessagePrototype: &domain.Signup{},
//		Handler: func(_ context.Context, evt events.ProtoMessageContext[*domain.Signup]) ([]events.ProtoMessageOutput, error) {
//			md := evt.CloneMetadata()
//			md["handler"] = "signup_to_billing"
//			return []events.ProtoMessageOutput{{
//				Message:  &domain.BillingAddress{},
//				Metadata: md,
//			}}, nil
//		},
//	})
//
// The library does not ship demo handlers anymore, but you can find
// equivalent logic inside the application package for reference.
package events
