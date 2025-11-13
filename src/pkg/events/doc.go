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
// The library does not ship demo handlers anymore, but you can find
// equivalent logic inside the application package for reference.
package events
