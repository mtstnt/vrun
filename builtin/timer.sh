# !/bin/sh

compileCmd="{{ .CompileCommand }}"
runCmd="{{ .RunCommand }}"

# time (
timeout 1s sh <<EOF
    echo "==== Program Output ===="
    echo "2" | $runCmd
EOF
# )

echo "==== Code: $?"
