package main

import (
	"fmt"
	"os"
)

func runProgram(progStr string) {
	l := lex("test", progStr)

	parser := NewParser(l)
	program, err := parser.ParseProgram()
	if err != nil {
		fmt.Printf("Parsing error: %s\n", err)
		os.Exit(1)
	}
	varInits := make(map[string]string)

	interp := new(Interpreter)
	interp.Init(program, varInits)
	interp.InterpProgram(program)
	if err != nil {
		fmt.Printf("Interpreting error: %s\n", err)
		os.Exit(1)
	}
}

func ExampleInterp1() {
	prog := `
  var
	i: integer
  endvar
  run
	  loop times 10
	    i = i + 1
	    print("hello" + " " + strInt(i))
	  endloop
  endrun`
	runProgram(prog)
	// Output:
	// hello 1
	// hello 2
	// hello 3
	// hello 4
	// hello 5
	// hello 6
	// hello 7
	// hello 8
	// hello 9
	// hello 10
}

func ExampleInterp2() {
	prog := `
var
  i: integer
endvar
run
    loop i < 10
		i = i + 1
        print("hello" + " " + strInt(i))
    endloop
endrun`
	runProgram(prog)
	// Output:
	// hello 1
	// hello 2
	// hello 3
	// hello 4
	// hello 5
	// hello 6
	// hello 7
	// hello 8
	// hello 9
	// hello 10
}

func ExampleInterp3() {
	prog := `
	var
  i: integer
endvar
run
    loop 
        if i = 10
			exit
        endif
		i = i + 1
        print("hello" + " " + strInt(i))
    endloop
endrun`
	runProgram(prog)
	// Output:
	// hello 1
	// hello 2
	// hello 3
	// hello 4
	// hello 5
	// hello 6
	// hello 7
	// hello 8
	// hello 9
	// hello 10
}

func ExampleInterp4() {
	prog := `
var
  i: integer
endvar
run
    loop 
        print(strInt(i))
        if 0 <= i & i < 5
			print("small")
		elseif 5 <= i & i < 10
			print("medium")
		elseif 10 <= i & i <= 15
			print("large")
        endif
        if i = 15
			exit
        endif
		i = i + 1
    endloop
endrun`
	runProgram(prog)
	// 0
	// small
	// 1
	// small
	// 2
	// small
	// 3
	// small
	// 4
	// small
	// 5
	// medium
	// 6
	// medium
	// 7
	// medium
	// 8
	// medium
	// 9
	// medium
	// 10
	// large
	// 11
	// large
	// 12
	// large
	// 13
	// large
	// 14
	// large
	// 15
}
