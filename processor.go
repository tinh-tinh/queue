package queue

import "github.com/tinh-tinh/tinhtinh/v2/core"

type Processor struct {
	core.DynamicProvider
	module core.Module
	name   string
}

func NewProcessor(name string, module core.Module) *Processor {
	return &Processor{
		module: module,
		name:   name,
	}
}

func (p *Processor) Process(jobFnc JobFnc) {
	pQueue := Inject(p.module, p.name)
	if pQueue != nil {
		pQueue.Process(jobFnc)
	}
}
