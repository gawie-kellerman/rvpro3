package utils

type IRunnableService interface {
	// SetupDefaults sets both all defaults for the service type
	SetupDefaults(*Settings)

	// SetupAndStart follows a consistent recipe where:
	// 1. Read the config and determine the number of services to be instantiated if any
	// 2. Create the number of services required
	// 3. Initialize each service from the config
	// 4. Register the service with the state by a unique name
	// 5. Initiate, and Start the service
	SetupAndStart(state *State, config *Settings)

	GetServiceName() string
	GetServiceNames() []string
}
