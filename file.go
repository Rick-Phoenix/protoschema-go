package schemabuilder

import "path"

type FileSchema struct {
	Package    *ProtoPackage
	Name       string
	Imports    Set
	Extensions Extensions
	Options    []ProtoOption
	Enums      []EnumGroup
	Messages   []*MessageSchema
	Services   []*ServiceSchema
}

type FileData struct {
	Package    *ProtoPackage
	Name       string
	Imports    Set
	Extensions Extensions
	Options    []ProtoOption
	Enums      []EnumGroup
	Messages   []MessageData
	Services   []ServiceData
}

func (f *FileSchema) NewMessage(s *MessageSchema) *MessageSchema {
	s.Package = f.Package
	s.File = f
	s.ImportPath = path.Join(f.Package.protoPackagePath, f.Name)
	f.Messages = append(f.Messages, s)
	return s
}

func (f *FileSchema) NewService(s *ServiceSchema) *ServiceSchema {
	s.Package = f.Package
	s.File = f
	f.Services = append(f.Services, s)
	return s
}
