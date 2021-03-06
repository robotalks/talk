package engine

import (
	"fmt"
	"reflect"

	"github.com/easeway/langx.go/errors"
	"github.com/robotalks/talk/contract/v0"
)

// ComponentFactoryFunc is func form of ComponentFactory
type ComponentFactoryFunc func(v0.ComponentRef) (v0.Component, error)

// CreateComponent implements ComponentFactory
func (f ComponentFactoryFunc) CreateComponent(ref v0.ComponentRef) (v0.Component, error) {
	return f(ref)
}

// ConfigComponent is a helper to map configuration into component
func ConfigComponent(comp v0.Component, ref v0.ComponentRef) error {
	conf := &MapConfig{Map: ref.ComponentConfig()}
	return conf.As(comp)
}

// SetupComponent is a helper to initialize a component using reflect
func SetupComponent(comp v0.Component, ref v0.ComponentRef) error {
	v := reflect.Indirect(reflect.ValueOf(comp))
	if v.Kind() != reflect.Struct {
		panic("not a struct")
	}
	errs := errors.AggregatedError{}
	errs.Add(ConfigComponent(comp, ref))
	t := v.Type()
	injectMap := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" || f.Anonymous { // unexported or anonymous field
			continue
		}
		if f.Type.Kind() != reflect.Interface {
			continue
		}

		if key := f.Tag.Get("inject"); key != "" {
			injectMap[key] = i
		}
	}

	for name, injection := range ref.Injections() {
		index, ok := injectMap[name]
		if !ok {
			continue
		}
		delete(injectMap, name)
		f := t.Field(index)
		fv := v.Field(index)
		if !fv.CanSet() {
			panic("field " + f.Name + " must be settable")
		}
		iv := reflect.ValueOf(injection)
		if ref, ok := injection.(v0.ComponentRef); ok {
			iv = reflect.ValueOf(ref.Component())
		}
		it := iv.Type()
		if it.AssignableTo(f.Type) {
			fv.Set(iv)
		} else if it.ConvertibleTo(f.Type) {
			fv.Set(iv.Convert(f.Type))
		} else {
			errs.Add(fmt.Errorf("%s injection %s type mismatch",
				comp.Ref().MessagePath(), name))
		}
	}

	for name := range injectMap {
		errs.Add(fmt.Errorf("%s injection %s unresolved",
			comp.Ref().MessagePath(), name))
	}

	return errs.Aggregate()
}

// RegisterComponentTypes registers named component types
func RegisterComponentTypes(types ...v0.ComponentType) {
	for _, t := range types {
		v0.DefaultComponentTypeRegistry.RegisterComponentType(t)
	}
}

// CustomComponentType is used as template to create instance types
type CustomComponentType struct {
	TypeName         string
	TypeDesc         string
	ComponentFactory v0.ComponentFactory
}

// Name implements v0.ComponentType
func (t *CustomComponentType) Name() string {
	return t.TypeName
}

// Description implements v0.ComponentType
func (t *CustomComponentType) Description() string {
	return t.TypeDesc
}

// Factory implements v0.ComponentType
func (t *CustomComponentType) Factory() v0.ComponentFactory {
	return t.ComponentFactory
}

// Describe provides type description
func (t *CustomComponentType) Describe(desc string) *CustomComponentType {
	t.TypeDesc = desc
	return t
}

// Register wraps RegisterInstanceType
func (t *CustomComponentType) Register() *CustomComponentType {
	RegisterComponentTypes(t)
	return t
}

// DefineComponentType defines a custom instance type
func DefineComponentType(name string, factory v0.ComponentFactory) *CustomComponentType {
	return &CustomComponentType{TypeName: name, ComponentFactory: factory}
}
