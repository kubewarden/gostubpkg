package gen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	inputDir = "testdata/testmod"
	module   = "github.com/gostubpkg/testmod"
)

type GenTestSuite struct {
	suite.Suite
	outputDir string
	inputDir  string
}

func (suite *GenTestSuite) SetupTest() {
	suite.outputDir = suite.T().TempDir()
}

func (suite *GenTestSuite) TestGenerateAllPackages() {
	err := GenerateStubs(inputDir, []string{"./..."}, suite.outputDir, false, nil, nil)
	suite.NoError(err)

	suite.True(suite.fileExists("main.go"))
	suite.True(suite.fileExists("pkg/funcs/funcs.go"))
	suite.True(suite.fileExists("pkg/types/types.go"))

	suite.False(suite.fileExists("go.mod"))
	// Skip generated files and tests
	suite.False(suite.fileExists("pkg/funcs/func_test.go"))
	suite.False(suite.fileExists("pkg/types/mocks/MyInterface.go"))

	generatedMain := suite.readFile("main.go")
	expectedMain := `package main

var var1 = "somevalue"

var Var2 = "someOtherValue"

const Const1 = 0

const const2 = 0

func Foo(e bool) error {
	panic("stub")
}

type Embedme interface{}
`

	suite.Equal(expectedMain, generatedMain)
}

func (suite *GenTestSuite) TestGenerateStubsGoMod() {
	err := GenerateStubs(inputDir, []string{"./..."}, suite.outputDir, true, nil, nil)
	suite.NoError(err)

	suite.True(suite.fileExists("go.mod"))

	generatedGoMod := suite.readFile("go.mod")
	expectedGoMod := `module github.com/gostubpkg/testmod

go 1.21.1
`
	suite.Equal(expectedGoMod, generatedGoMod)
}

func (suite *GenTestSuite) TestGenerateStubsFuncsPackage() {
	err := GenerateStubs(inputDir, []string{"./pkg/funcs"}, suite.outputDir, true, nil, nil)
	suite.NoError(err)

	suite.True(suite.fileExists("pkg/funcs/funcs.go"))
	suite.True(suite.fileExists("go.mod"))
	suite.False(suite.fileExists("pkg/types/types.go"))
	suite.False(suite.fileExists("main.go"))

	generatedFuncs := suite.readFile("pkg/funcs/funcs.go")
	expectedFuncs := `package funcs

import "io"

func Bar[T1 any, T2 int](t1 []T1, t2 T2) T2 {
	panic("stub")
}

func Baz(pod interface{}, writer io.Writer, str string) error {
	panic("stub")
}

type Embedme interface{}
`

	suite.Equal(expectedFuncs, generatedFuncs)
}

func (suite *GenTestSuite) TestGenerateStubsTypesPackage() {
	err := GenerateStubs(inputDir, []string{"./pkg/types"}, suite.outputDir, false, nil, nil)
	suite.NoError(err)

	suite.True(suite.fileExists("pkg/types/types.go"))
	suite.False(suite.fileExists("go.mod"))
	suite.False(suite.fileExists("pkg/funcs/funcs.go"))
	suite.False(suite.fileExists("main.go"))

	generatedTypes := suite.readFile("pkg/types/types.go")
	expectedTypes := `package types

import (
	"io"
	"os"
)

type MyEmbeddedStruct struct{}

type MyInterface interface {
	GetPodName(pod interface{}) string
	getPodNamePrivate(pod interface{}) string
}

type MyStruct struct {
	MyEmbeddedStruct
	Embedme
	Name     string
	Num      int
	Pointer  *os.File
	IOReader io.Reader
	Pod      interface{}
}

type MyTypeAlias string

type MyTypeAlias2 MyStruct

type MyTypeAlias3 interface{}

func (s *MyStruct) GetPodName(pod interface{}) string {
	panic("stub")
}

type Embedme interface{}
`

	suite.Equal(expectedTypes, generatedTypes)
}

func (suite *GenTestSuite) TestGenerateStubsAllowImports() {
	err := GenerateStubs(inputDir, []string{"./..."}, suite.outputDir, false, []string{"k8s.io/api/core/v1"}, nil)
	suite.NoError(err)

	generatedFuncs := suite.readFile("pkg/funcs/funcs.go")
	expectedFuncs := `package funcs

import (
	"io"

	corev1 "k8s.io/api/core/v1"
)

func Bar[T1 any, T2 int](t1 []T1, t2 T2) T2 {
	panic("stub")
}

func Baz(pod *corev1.Pod, writer io.Writer, str string) error {
	panic("stub")
}

type Embedme interface{}
`

	suite.Equal(expectedFuncs, generatedFuncs)

	generatedTypes := suite.readFile("pkg/types/types.go")
	expectedTypes := `package types

import (
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
)

type MyEmbeddedStruct struct{}

type MyInterface interface {
	GetPodName(pod *corev1.Pod) string
	getPodNamePrivate(pod *corev1.Pod) string
}

type MyStruct struct {
	MyEmbeddedStruct
	corev1.PodSpec
	Name     string
	Num      int
	Pointer  *os.File
	IOReader io.Reader
	Pod      corev1.Pod
}

type MyTypeAlias string

type MyTypeAlias2 MyStruct

type MyTypeAlias3 corev1.Pod

func (s *MyStruct) GetPodName(pod *corev1.Pod) string {
	panic("stub")
}

type Embedme interface{}
`

	suite.Equal(expectedTypes, generatedTypes)
}

func (suite *GenTestSuite) TestGenerateStubsFunctionBodies() {
	err := GenerateStubs(inputDir, []string{"./..."}, suite.outputDir, false, nil, map[string]string{
		"funcs.Bar":                    `panic("i don't like generics")`,
		"types.(*MyStruct).GetPodName": `return "StubPodName"`,
	})
	suite.NoError(err)

	generatedFuncs := suite.readFile("pkg/funcs/funcs.go")
	suite.Contains(generatedFuncs, `func Bar[T1 any, T2 int](t1 []T1, t2 T2) T2 {
	panic("i don't like generics")
`)

	generatedTypes := suite.readFile("pkg/types/types.go")
	suite.Contains(generatedTypes, `func (s *MyStruct) GetPodName(pod interface{}) string {
	return "StubPodName"
}
`)
}

func (suite *GenTestSuite) filePath(filename string) string {
	return filepath.Join(suite.outputDir, module, filename)
}

func (suite *GenTestSuite) fileExists(filename string) bool {
	info, err := os.Stat(suite.filePath(filename))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (suite *GenTestSuite) readFile(filename string) string {
	file, err := os.ReadFile(suite.filePath(filename))
	suite.NoError(err)

	return string(file)
}

func TestGenTestSuite(t *testing.T) {
	suite.Run(t, new(GenTestSuite))
}
