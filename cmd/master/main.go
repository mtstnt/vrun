package main

import (
	"context"
	"fmt"

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

	result, err := runner.Run(ctx, task)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", result)
}
