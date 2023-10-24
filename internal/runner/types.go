package runner

// FileMap is a map of [relative path]: [file content]
// Used to dynamically specify file with its directory structure
// as they are being sent via API.
type FileMap map[string][]byte

type OutputCheckStrategy int

const (
	StringComparison = iota
	FloatingPointComparison
)

// TestCase stores the given input and expected output.
// The OutputCheckStrategy is used to determine how a testcase should be checked.
type TestCase struct {
	Input               string
	Output              string
	OutputCheckStrategy int
}

// Config is the configuration for the environment on which the code is going to be run.
type Config struct {
	Dockerfile     string
	CompileCommand string
	RunCommand     string
	MemoryLimit    int
}

// Task is the main unit that is being constructed and sent for execution.
type Task struct {
	Config     Config
	FileSystem FileMap
	TestCases  []TestCase
}

// WIP: Result is the object that reports the results of code execution.
type Result struct {
	CorrectTestcases int
}
