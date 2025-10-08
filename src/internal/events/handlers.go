package events

func (s *Service) addAllHandlers() {
	// This is just for demonstration purposes.
	// In a real application, you would have different handlers for different topics.
	s.Router.AddHandler(
		"demoHandler",       // handler name, must be unique
		s.Conf.ConsumeTopic, // topic from which messages should be consumed
		s.Subscriber,
		s.Conf.PublishTopic, // topic to which messages should be published
		s.Publisher,
		s.demoHandlerFunc(),
	)

	// Add the signup handler
	s.Router.AddHandler(
		"signupHandler",
		s.Conf.ConsumeTopicSignup,
		s.Subscriber,
		s.Conf.PublishTopicSignup,
		s.Publisher,
		s.signupHandlerFunc(),
	)

	// Add the signup step 2 handler
	s.Router.AddHandler(
		"signupStepTwoHandler",
		s.Conf.PublishTopicSignup,
		s.Subscriber,
		"signup_step_2_processed", // This is just for demo
		s.Publisher,
		s.signupStepTwoHandlerFunc(),
	)
}
