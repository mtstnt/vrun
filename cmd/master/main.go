package main

import (
	"context"
	"fmt"

	"github.com/mtstnt/vrun/internal/runner"
)

func main() {
	task := runner.Task{
		Config: runner.Config{
			ImageName:      "python:3.11.1-debian",
			CompileCommand: "",
			RunCommand:     "python main.py",
		},
		FileSystem: runner.SourceFileMap{
			"main.py": []byte(`
import hello
hello.hello()
			`),
			"hello/__init__.py": []byte(`
def hello(): print("hello world")
		`),
		},
		TestCases: []string{
			"hello world",
		},
	}

	ctx := context.Background()

	result, err := runner.Run(ctx, task)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", result)
}
