# gostubpkg

gostubpkg is a tool for generating stubs of Go packages.

## Goals and non-goals

In the context of gostubpkg, _stubbing_ means replacing the package with a dummy implementation that provides the same API as the original package.
This tool is intended to be used to slim down the size of Go binaries by replacing dependencies with stubs,
specifically when dealing with [WebAssembly](https://webassembly.org/) binaries.
Therefore, this tool is not intended to be used for mocking or testing, as there are [other tools](https://github.com/avelino/awesome-go#testing) that are better suited for these purposes.

## Installation

You can install gostubpkg using the following command:

```shell
go get -u github.com/kubewarden/gostubpkg
```

## Usage

```shell
gostubpkg [flags] <patterns>...

Flags:
  -a, --allow-imports strings            Specify this flag multiple times to add external imports
                                         that will not be removed from the generated stubs.
                                         Example: -a k8s.io/api/core/v1
  -c, --config string                    config file (default "gostubpkg.yaml")
  -f, --function-bodies stringToString   Specify this flag multiple times to add a type mapping.
                                         Example: -f "cmd.Execute"='println("hello world")' -f "yourpkg.(*YourType).YourMethod"='return nil' (default [])
  -m, --generate-go-mod                  Generate the go.mod file in the root of the stub package
  -h, --help                             help for gostubpkg
  -i, --input-dir string                 Specify the directory in which to run the build system's query tool that provides information about the packages (default $PWD)
  -o, --output-dir string                Specify the output directory for the stubs (default $PWD)
  -v, --verbose count                    Increase output verbosity. Example: --verbose=2 or -vv
```

### Generate stubs for all packages

```shell
gostubpkg -m -i /path/to/module -o /path/to/output ./...
```

This will generate stubs and a `go.mod` file for all packages in the specified input directory.
All the functions in the stubs will panic when called, and all the external imports will be removed.
External types will be replaced with `interface{}` in struct fields, type aliases, and function signatures.

Private functions, private struct fields, private struct methods, generated files, and test files will be ignored.
Private types and interfaces will be kept in the stubs since they could be embedded in public types.

For instance, this code:

```go
package yourpkg

import (
    "fmt"
    "io"

    corev1 "k8s.io/api/core/v1"
)

type YourType struct {
    Pod corev1.Pod
}

type yourOtherType struct {
    Writer io.Writer
    Foo()
}

type YourAlias corev1.Pod

func Foo(pod *corev1.Pod) {
    fmt.Println(pod.Name)
}

func bar(pod *corev1.Pod) string  {
    fmt.Println(pod.Name)
}
```

Will be replaced with:

```go
package yourpkg

import (
    "fmt"
    "io"
)

type YourType struct {
    Pod interface{}
}

type yourOtherType struct {
    Writer io.Writer
    Foo()
}

type YourAlias interface{}

func Foo(pod interface{}) {
    panic("stub")
}
```

### Generate stubs for specific packages and allow external imports

```shell
gostubpkg -i /path/to/your/code -o /path/to/output ./yourpkg --allow-imports k8s.io/api/core/v1
```

This command generates stubs for packages in the specified input directory while allowing external imports from `k8s.io/api/core/v1`.

For instance, this code snippet:

```go
package yourpkg

import (
    corev1 "k8s.io/api/core/v1"
)

type YourType struct {
    Pod *corev1.Pod
}

func (t *YourType) SetPod(pod *corev1.Pod) {
    t.Pod = pod
}
```

Will be replaced with:

```go
package yourpkg

import (
    corev1 "k8s.io/api/core/v1"
)

type YourType struct {
    Pod corev1.Pod
}

func (t *YourType) SetPod(pod *corev1.Pod) {
    panic("stub")
}
```

### Custom function bodies

Sometimes you may want to specify custom function bodies for the stubs.

```shell
gostubpkg -f "yourpkg.Foo"='println("hello stub")' -f "yourpkg.(*YourType).YourMethod"='return nil'
```

This code snippet:

```go
package yourpkg

func Foo() {
    fmt.Println("hello world")
}

type YourType struct {}

func (t *YourType) YourMethod() error {
    return fmt.Errorf("error")
}
```

Will be replaced with:

```go
package yourpkg

func Foo() {
    println("hello stub")
}

type YourType struct {}

func (t *YourType) YourMethod() error {
    return nil
}
```

## Configuration

gostubpkg supports a configuration file in YAML format.
The configuration file can be specified using the `--config` flag.

Example `gostubpkg.yaml`:

```yaml
allow-imports:
  - k8s.io/api/core/v1

generate-go-mod: true

function-bodies:
  cmd.Execute: 'println("hello world")'
  yourpkg.(*YourType).YourMethod: "return nil"
```
