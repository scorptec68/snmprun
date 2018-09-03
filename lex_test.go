package main

import "fmt"

func printTokens(l *lexer) {
	for {
		it := l.nextItem()
		fmt.Println("item:", it)
		if it.typ == itemEOF {
			break
		}
	}
}

func ExampleLexing1() {

	// test input
	inputStr := `var
       x: integer
    endvar
    run
       x = 3
    endrun`

	printTokens(lex("test", inputStr))
	// Output:
	// item: <var>
	// item: "\n" (type 19)
	// item: "x" (type 21)
	// item: ":" (type 18)
	// item: <integer>
	// item: "\n" (type 19)
	// item: <endvar>
	// item: "\n" (type 19)
	// item: <run>
	// item: "\n" (type 19)
	// item: "x" (type 21)
	// item: "=" (type 3)
	// item: "3" (type 1)
	// item: "\n" (type 19)
	// item: <endrun>
	// item: EOF
	//

}

func ExampleLexing2() {

	// test input
	inputStr := `var
       this-boy : string
    endvar
    run
       this-boy = "hi there"
    endrun`

	printTokens(lex("test", inputStr))
	// Output:
	// item: <var>
	// item: "\n" (type 19)
	// item: "this-boy" (type 21)
	// item: ":" (type 18)
	// item: <string>
	// item: "\n" (type 19)
	// item: <endvar>
	// item: "\n" (type 19)
	// item: <run>
	// item: "\n" (type 19)
	// item: "this-boy" (type 21)
	// item: "=" (type 3)
	// item: "hi there" (type 2)
	// item: "\n" (type 19)
	// item: <endrun>
	// item: EOF
	//
}

func ExampleLexing3() {
	// try out the readme example

	// test input
	inputStr := `	var
	   test: string
	   x: integer
	   y: integer
	   i: integer
	   doit: boolean
	   this-boy: boolean
	endvar

	run
	    this-boy = true
	    x = 3 + 1000000 + 2 +
	          4 + 6
	    x = 3
		test = "hello"

		y = x + 3
		print("y =  " + str(x))

		if y > 3
			test = "super"
		elseif y < 3
		    test = "wonder"
		else
		    test = "duper"
		endif

		print(test)
		test = test - "r"
		print(test)

		loop 10
		    i = i + 1
		endloop

		print("i = " + str(i))
	endrun`

	printTokens(lex("test", inputStr))
	// Output:
	// item: <var>
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: ":" (type 18)
	// item: <string>
	// item: "\n" (type 19)
	// item: "x" (type 21)
	// item: ":" (type 18)
	// item: <integer>
	// item: "\n" (type 19)
	// item: "y" (type 21)
	// item: ":" (type 18)
	// item: <integer>
	// item: "\n" (type 19)
	// item: "i" (type 21)
	// item: ":" (type 18)
	// item: <integer>
	// item: "\n" (type 19)
	// item: "doit" (type 21)
	// item: ":" (type 18)
	// item: <boolean>
	// item: "\n" (type 19)
	// item: "this-boy" (type 21)
	// item: ":" (type 18)
	// item: <boolean>
	// item: "\n" (type 19)
	// item: <endvar>
	// item: "\n" (type 19)
	// item: <run>
	// item: "\n" (type 19)
	// item: "this-boy" (type 21)
	// item: "=" (type 3)
	// item: <true>
	// item: "\n" (type 19)
	// item: "x" (type 21)
	// item: "=" (type 3)
	// item: "3" (type 1)
	// item: "+" (type 12)
	// item: "1000000" (type 1)
	// item: "+" (type 12)
	// item: "2" (type 1)
	// item: "+" (type 12)
	// item: "4" (type 1)
	// item: "+" (type 12)
	// item: "6" (type 1)
	// item: "\n" (type 19)
	// item: "x" (type 21)
	// item: "=" (type 3)
	// item: "3" (type 1)
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: "=" (type 3)
	// item: "hello" (type 2)
	// item: "\n" (type 19)
	// item: "y" (type 21)
	// item: "=" (type 3)
	// item: "x" (type 21)
	// item: "+" (type 12)
	// item: "3" (type 1)
	// item: "\n" (type 19)
	// item: <print>
	// item: "(" (type 16)
	// item: "y =  " (type 2)
	// item: "+" (type 12)
	// item: "str" (type 21)
	// item: "(" (type 16)
	// item: "x" (type 21)
	// item: ")" (type 17)
	// item: ")" (type 17)
	// item: "\n" (type 19)
	// item: <if>
	// item: "y" (type 21)
	// item: ">" (type 8)
	// item: "3" (type 1)
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: "=" (type 3)
	// item: "super" (type 2)
	// item: "\n" (type 19)
	// item: <elseif>
	// item: "y" (type 21)
	// item: "<" (type 7)
	// item: "3" (type 1)
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: "=" (type 3)
	// item: "wonder" (type 2)
	// item: "\n" (type 19)
	// item: <else>
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: "=" (type 3)
	// item: "duper" (type 2)
	// item: "\n" (type 19)
	// item: <endif>
	// item: "\n" (type 19)
	// item: <print>
	// item: "(" (type 16)
	// item: "test" (type 21)
	// item: ")" (type 17)
	// item: "\n" (type 19)
	// item: "test" (type 21)
	// item: "=" (type 3)
	// item: "test" (type 21)
	// item: "-" (type 13)
	// item: "r" (type 2)
	// item: "\n" (type 19)
	// item: <print>
	// item: "(" (type 16)
	// item: "test" (type 21)
	// item: ")" (type 17)
	// item: "\n" (type 19)
	// item: <loop>
	// item: "10" (type 1)
	// item: "\n" (type 19)
	// item: "i" (type 21)
	// item: "=" (type 3)
	// item: "i" (type 21)
	// item: "+" (type 12)
	// item: "1" (type 1)
	// item: "\n" (type 19)
	// item: <endloop>
	// item: "\n" (type 19)
	// item: <print>
	// item: "(" (type 16)
	// item: "i = " (type 2)
	// item: "+" (type 12)
	// item: "str" (type 21)
	// item: "(" (type 16)
	// item: "i" (type 21)
	// item: ")" (type 17)
	// item: ")" (type 17)
	// item: "\n" (type 19)
	// item: <endrun>
	// item: EOF
}

func ExampleLexing4() {

	// test input
	inputStr := `var
		model 1.3.6.1.6.7.7 integer [1 = 'start', 2 = 'finish']
		endvar`

	printTokens(lex("test", inputStr))
	// Output:
	// item: <var>
	// item: "\n" (type new line)
	// item: "model" (type identifier)
	// item: "1.3.6.1.6.7.7" (type OID)
	// item: <integer>
	// item: "[" (type [)
	// item: "1" (type int literal)
	// item: "=" (type =)
	// item: "start" (type alias)
	// item: "," (type ,)
	// item: "2" (type int literal)
	// item: "=" (type =)
	// item: "finish" (type alias)
	// item: "]" (type ])
	// item: "\n" (type new line)
	// item: <endvar>
	// item: EOF
}
