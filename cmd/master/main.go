package main

import (
	"context"
	"fmt"

	"github.com/mtstnt/vrun/internal/core"
	"github.com/mtstnt/vrun/internal/runner"
)

func main() {
	task := runner.Task{
		Config: runner.Config{
			Dockerfile: `FROM python:3.12.0-slim
CMD ["sleep", "infinity"]
`,
			CompileCommand: "",
			RunCommand:     "python3 main.py",
		},
		FileSystem: runner.FileMap{
			"main.py": []byte(`
import hello
a = int(input())
b = int(input())
print(hello.add(a, b))
			`),
			"hello/__init__.py": []byte(`
def add(a, b): return a + b
		`),
		},
		TestCases: []runner.TestCase{
			{
				Input:  "1\n2",
				Output: "3",
			},
			{
				Input:  "3\n99",
				Output: "102",
			},
		},
	}

	ctx := context.Background()

	// TODO: Mock this part for easier testing.
	// Create interface just to do stuff with Docker Client.
	client, err := core.GetDockerClient()
	if err != nil {
		panic(err)
	}

	result, err := runner.Run(client, ctx, task)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", result)
}
