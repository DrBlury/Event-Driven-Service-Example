// Package events provides application-agnostic helpers for wiring
// Watermill publishers, subscribers, and middlewares.  Create a
// Service, register typed handlers via the helpers in this package,
// and finally start the router.  The middleware chain can be tweaked by
// appending to ServiceDependencies.Middlewares or by replacing it entirely
// with DisableDefaultMiddlewares.
//
// When processing protobuf events, you can rely on RegisterProtoHandler to
// automatically unmarshal payloads into typed messages and emit proto-based
// responses without hand-written boilerplate:
//
//	_ = events.RegisterProtoHandler(svc, events.ProtoHandlerRegistration[*domain.Signup]{
//		Name:         "signup_to_billing",
//		ConsumeQueue: cfg.ConsumeQueueSignup,
//		PublishQueue: cfg.PublishQueueSignup,
//		ConsumeMessageType: &domain.Signup{},
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
//
// The PublishMessageType field is optional—omit it if your handler builds
// outgoing messages manually, or pass WithPublishMessageTypes when you need to
// register additional schemas for validator/outbox bookkeeping.
//
// When dealing with JSON payloads you can reach for RegisterJSONHandler
// which mirrors the protobuf helper but uses sonic for marshaling:
//
//	_ = events.RegisterJSONHandler(svc, events.JSONHandlerRegistration[*signupEvent, *processedSignupEvent]{
//		Name:             "audit_signup",
//		ConsumeQueue:     cfg.ConsumeQueueSignup,
//		PublishQueue:     cfg.PublishQueueSignup,
//		ConsumeMessageType: &signupEvent{},
//		Handler: func(_ context.Context, evt events.JSONMessageContext[*signupEvent]) ([]events.JSONMessageOutput[*processedSignupEvent], error) {
//			md := evt.CloneMetadata()
//			md["handler"] = "audit_signup"
//			return []events.JSONMessageOutput[*processedSignupEvent]{
//				{Message: &processedSignupEvent{ID: evt.Payload.ID}, Metadata: md},
//			}, nil
//		},
//	})
package events
