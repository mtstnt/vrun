# !/bin/sh

compileCmd="{{ .Config.CompileCommand }}"
runCmd="{{ .Config.RunCommand }}"

for testcase in ../tests/*.txt; do
    tcContent="$(cat $testcase)"
    timeout 1s sh <<EOF
        echo "==== Program Output ===="
        echo "TestCase No: $testcase"
        echo "TestCase Content: $tcContent"
        echo "====================="
        echo "$tcContent" | $runCmd
EOF
done

echo "==== Status Code: $?"
