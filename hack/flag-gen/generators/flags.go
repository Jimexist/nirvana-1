/*
Copyright 2017 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generators

import (
	"io"

	flagNamer "github.com/caicloud/nirvana/hack/flag-gen/namer"
	"github.com/golang/glog"
	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public":  flagNamer.NewPublicNamer(0),
		"private": flagNamer.NewPrivateNamer(0),
		"raw":     namer.NewRawNamer("", nil),
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

// Packages makes the sets package definition.
func Packages(_ *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		glog.Fatalf("Failed loading boilerplate: %v", err)
	}
	return generator.Packages{&generator.DefaultPackage{
		PackageName: "cli",
		PackagePath: arguments.OutputPackagePath,
		HeaderText: append(boilerplate, []byte(
			`
// This file was autogenerated by set-gen. Do not edit it manually!

`)...),
		GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
			glog.Info("generator func")
			generators = []generator.Generator{}

			for _, t := range c.Order {
				generators = append(generators,
					&flagsGenerator{
						DefaultGen: generator.DefaultGen{
							OptionalName: "flags_generated",
						},
						outputPackage: arguments.OutputPackagePath,
						typeToMatch:   t,
						imports:       generator.NewImportTracker(),
					},
					&flagsTestGenerator{
						DefaultGen: generator.DefaultGen{
							OptionalName: "flags_generated_test",
						},
						outputPackage: arguments.OutputPackagePath,
						typeToMatch:   t,
						imports:       generator.NewImportTracker(),
					},
					&viperGenerator{
						DefaultGen: generator.DefaultGen{
							OptionalName: "config_generated",
						},
						outputPackage: arguments.OutputPackagePath,
						typeToMatch:   t,
						imports:       generator.NewImportTracker(),
					},
					&viperTestGenerator{
						DefaultGen: generator.DefaultGen{
							OptionalName: "config_generated_test",
						},
						outputPackage: arguments.OutputPackagePath,
						typeToMatch:   t,
						imports:       generator.NewImportTracker(),
					},
				)
			}
			return generators
		},
		FilterFunc: func(c *generator.Context, t *types.Type) bool {
			switch t.Kind {
			case types.Builtin, types.Alias:
				if t.Name == types.Byte.Name {
					return false
				}
				return true
			case types.Slice:
				if t.Elem.Name == types.Byte.Name {
					return false
				}
				return true
			case types.Struct:
				if t.Name.String() == "net.IPNet" {
					return true
				}
				return false
			}
			return false
		},
	}}
}

var _ generator.Generator = &flagsGenerator{}

type flagsGenerator struct {
	generator.DefaultGen
	outputPackage string
	imports       namer.ImportTracker
	typeToMatch   *types.Type
}

func (g *flagsGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToMatch
}

func (g *flagsGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *flagsGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "github.com/spf13/pflag")
	imports = append(imports, "github.com/spf13/cast")
	return
}

func (g *flagsGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	glog.V(5).Infof("processing type %v", t)
	m := map[string]interface{}{
		"type": t,
	}
	sw.Do(flagCode, m)

	return sw.Error()
}

var flagCode = `

var _ Flag = $.type|public$Flag{}

// $.type|public$Flag is a flag of type $.type|raw$
type $.type|public$Flag struct {
	// Name as it appears on command line
	Name        string
	// one-letter abbreviated flag
	Shorthand   string
	// help message
	Usage       string
	// specify whether the flag is persistent
	Persistent  bool
	// used by cobra.Command bash autocomple code
	Annotations map[string][]string
	// If this flag is deprecated, this string is the new or now thing to use
	Deprecated          string
	// If the shorthand of this flag is deprecated, this string is the new or now thing to use
	ShorthandDeprecated string
	// used by cobra.Command to allow flags to be hidden from help/usage text
	Hidden              bool
	// bind the flag to env key, you can use AutomaticEnv to bind all flags to env automatically
	// if EnvKey is set, it will override the automatic generated env key
	EnvKey string
	// the default value
	DefValue   $.type|raw$
	// points to a variable in which to store the value of the flag
	Destination *$.type|raw$
}	

// IsPersistent specify whether the flag is persistent
func (f $.type|public$Flag) IsPersistent() bool {
	return f.Persistent
}

// GetName returns the flag's name
func (f $.type|public$Flag) GetName() string {
	return f.Name
}

// ApplyTo adds the flag to given FlagSet
func (f $.type|public$Flag) ApplyTo(fs *pflag.FlagSet) error {

	if f.Destination == nil {
		f.Destination = new($.type|raw$)
	}

	realEnv, value := getEnv(f.Name, f.EnvKey, f.DefValue)
	defValue := cast.To$.type|public$(value)

	// append env key to usage
	usage := appendEnvToUsage(f.Usage, realEnv)

	fs.$.type|public$VarP(f.Destination, f.Name, f.Shorthand, defValue, usage)

	var err error

	if f.Deprecated != "" {
		err = fs.MarkDeprecated(f.Name, f.Deprecated)
		if err != nil {
			return err
		}
	}
	if f.ShorthandDeprecated != "" {
		err = fs.MarkShorthandDeprecated(f.Name, f.ShorthandDeprecated)
		if err != nil {
			return err
		}
	}
	if f.Hidden {
		err = fs.MarkHidden(f.Name)
		if err != nil {
			return err
		}
	}
	for key, values := range f.Annotations {
		err = fs.SetAnnotation(f.Name, key, values)
		if err != nil {
			return err
		}
	}

	return v.BindPFlag(f.Name, fs.Lookup(f.Name))
}
`

type flagsTestGenerator struct {
	generator.DefaultGen
	outputPackage string
	imports       namer.ImportTracker
	typeToMatch   *types.Type
}

func (g *flagsTestGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToMatch
}

func (g *flagsTestGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *flagsTestGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "github.com/spf13/cast")
	imports = append(imports, "github.com/spf13/pflag")
	imports = append(imports, "github.com/spf13/viper")
	imports = append(imports, "testing")
	imports = append(imports, "reflect")
	return
}

func (g *flagsTestGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	glog.V(5).Infof("processing type %v", t)
	m := map[string]interface{}{
		"type": t,
	}
	sw.Do(flagTestCode, m)

	return sw.Error()
}

var flagTestCode = `
func Test$.type|public$Flag(t *testing.T) {
	// reset viper
	Reset()	

	testcase := getTestCase("$.type|public$")
	dest := new($.type|raw$)
	f := $.type|public$Flag{
		Name:                "test",
		Shorthand:           "t",
		Usage:               "help",
		Persistent:          true,
		Annotations:         map[string][]string{"key": []string{"value"}},
		Deprecated:          "for test",
		ShorthandDeprecated: "for test",
		Hidden:              true,
		EnvKey:              "TEST",
		DefValue:            *dest,
		Destination:         dest,
	}

	f.IsPersistent()
	f.GetName()

	fs := pflag.NewFlagSet("dev", pflag.ContinueOnError)
	f.ApplyTo(fs)


	// test flag
	fs.Parse([]string{"-t=" + testcase.flag})
	want, err := cast.To$.type|public$E(testcase.want)
	assert.Nil(t, err)
	assert.Equal(t, want, *dest)

	// test viper
	v := v.Get(f.Name)
	got, err := cast.To$.type|public$E(v)
	assert.Nil(t, err)
	assert.Equal(t, want, got)

}
`
