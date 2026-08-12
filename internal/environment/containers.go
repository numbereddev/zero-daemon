package environment

type Environment struct{}

func (e *Environment) IsRunning() bool {
	return false
}

func (e *Environment) Create() error {
	return nil
}

func (e *Environment) Delete() error {
	return nil
}

func (e *Environment) Start() error {
	return nil
}

func (e *Environment) Stop() error {
	return nil
}

func (e *Environment) Restart() error {
	return nil
}

func (e *Environment) Attach() error {
	return nil
}

// TODO: callback type and stuff
func (e *Environment) LogCallback() {}
