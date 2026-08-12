package git

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GitUploadPack() {
}

func (s *Service) GitReceivePack() {
}

func (s *Service) GitInfoRefs() {
}
