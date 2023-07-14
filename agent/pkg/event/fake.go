package event

type Fake struct {
}

func NewFake() Interface {
	return &Fake{}
}

func (f *Fake) CreateDeploymentEvent(namespace, deployment, event, message string) error {
	return nil
}
