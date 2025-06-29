package schemabuilder

import "path"

type ProtoPackageConfig struct {
	Name          string
	ProtoRoot     string
	GoPackage     string
	GoPackageName string
	GoModule      string
}

type ProtoPackage struct {
	name          string
	protoRoot     string
	goPackagePath string
	goPackageName string
	goModule      string
}

func NewProtoPackage(conf ProtoPackageConfig) *ProtoPackage {
	p := &ProtoPackage{
		name:          conf.Name,
		protoRoot:     conf.ProtoRoot,
		goPackageName: conf.GoPackageName,
		goPackagePath: conf.GoPackage,
		goModule:      conf.GoModule,
	}

	if p.goPackageName == "" {
		p.goPackageName = path.Base(p.goPackagePath)
	}

	return p
}

func (p *ProtoPackage) NewMessage(s MessageSchema) MessageSchema {
	s.GoPackageName = p.goPackageName
	s.GoPackagePath = p.goPackagePath
	s.ProtoPackage = p.name
	return s
}
