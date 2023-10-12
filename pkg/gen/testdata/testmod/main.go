package main

import "fmt"

var (
	var1 = "somevalue"
	Var2 = "someOtherValue"
)

const Const1, const2 = 0, 1

// I am a comment, I will be gone in the stub!

func main() {
	fmt.Println("Hello, stub!")
}

func Foo(e bool) error {
	if !e {
		return nil
	}

	return fmt.Errorf("error")
}
