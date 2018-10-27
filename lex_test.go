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
	// item: "\n" (type new line)
	// item: "x" (type identifier)
	// item: ":" (type :)
	// item: <integer>
	// item: "\n" (type new line)
	// item: <endvar>
	// item: "\n" (type new line)
	// item: <run>
	// item: "\n" (type new line)
	// item: "x" (type identifier)
	// item: "=" (type =)
	// item: "3" (type int literal)
	// item: "\n" (type new line)
	// item: <endrun>
	// item: EOF
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
	// item: <var>
	// item: "\n" (type new line)
	// item: "this-boy" (type identifier)
	// item: ":" (type :)
	// item: <string>
	// item: "\n" (type new line)
	// item: <endvar>
	// item: "\n" (type new line)
	// item: <run>
	// item: "\n" (type new line)
	// item: "this-boy" (type identifier)
	// item: "=" (type =)
	// item: "hi there" (type string literal)
	// item: "\n" (type new line)
	// item: <endrun>
	// item: EOF
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
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: ":" (type :)
	// item: <string>
	// item: "\n" (type new line)
	// item: "x" (type identifier)
	// item: ":" (type :)
	// item: <integer>
	// item: "\n" (type new line)
	// item: "y" (type identifier)
	// item: ":" (type :)
	// item: <integer>
	// item: "\n" (type new line)
	// item: "i" (type identifier)
	// item: ":" (type :)
	// item: <integer>
	// item: "\n" (type new line)
	// item: "doit" (type identifier)
	// item: ":" (type :)
	// item: <boolean>
	// item: "\n" (type new line)
	// item: "this-boy" (type identifier)
	// item: ":" (type :)
	// item: <boolean>
	// item: "\n" (type new line)
	// item: <endvar>
	// item: "\n" (type new line)
	// item: <run>
	// item: "\n" (type new line)
	// item: "this-boy" (type identifier)
	// item: "=" (type =)
	// item: <true>
	// item: "\n" (type new line)
	// item: "x" (type identifier)
	// item: "=" (type =)
	// item: "3" (type int literal)
	// item: "+" (type +)
	// item: "1000000" (type int literal)
	// item: "+" (type +)
	// item: "2" (type int literal)
	// item: "+" (type +)
	// item: "4" (type int literal)
	// item: "+" (type +)
	// item: "6" (type int literal)
	// item: "\n" (type new line)
	// item: "x" (type identifier)
	// item: "=" (type =)
	// item: "3" (type int literal)
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: "=" (type =)
	// item: "hello" (type string literal)
	// item: "\n" (type new line)
	// item: "y" (type identifier)
	// item: "=" (type =)
	// item: "x" (type identifier)
	// item: "+" (type +)
	// item: "3" (type int literal)
	// item: "\n" (type new line)
	// item: <print>
	// item: "(" (type ()
	// item: "y =  " (type string literal)
	// item: "+" (type +)
	// item: "str" (type identifier)
	// item: "(" (type ()
	// item: "x" (type identifier)
	// item: ")" (type ))
	// item: ")" (type ))
	// item: "\n" (type new line)
	// item: <if>
	// item: "y" (type identifier)
	// item: ">" (type >)
	// item: "3" (type int literal)
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: "=" (type =)
	// item: "super" (type string literal)
	// item: "\n" (type new line)
	// item: <elseif>
	// item: "y" (type identifier)
	// item: "<" (type <)
	// item: "3" (type int literal)
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: "=" (type =)
	// item: "wonder" (type string literal)
	// item: "\n" (type new line)
	// item: <else>
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: "=" (type =)
	// item: "duper" (type string literal)
	// item: "\n" (type new line)
	// item: <endif>
	// item: "\n" (type new line)
	// item: <print>
	// item: "(" (type ()
	// item: "test" (type identifier)
	// item: ")" (type ))
	// item: "\n" (type new line)
	// item: "test" (type identifier)
	// item: "=" (type =)
	// item: "test" (type identifier)
	// item: "-" (type -)
	// item: "r" (type string literal)
	// item: "\n" (type new line)
	// item: <print>
	// item: "(" (type ()
	// item: "test" (type identifier)
	// item: ")" (type ))
	// item: "\n" (type new line)
	// item: <loop>
	// item: "10" (type int literal)
	// item: "\n" (type new line)
	// item: "i" (type identifier)
	// item: "=" (type =)
	// item: "i" (type identifier)
	// item: "+" (type +)
	// item: "1" (type int literal)
	// item: "\n" (type new line)
	// item: <endloop>
	// item: "\n" (type new line)
	// item: <print>
	// item: "(" (type ()
	// item: "i = " (type string literal)
	// item: "+" (type +)
	// item: "str" (type identifier)
	// item: "(" (type ()
	// item: "i" (type identifier)
	// item: ")" (type ))
	// item: ")" (type ))
	// item: "\n" (type new line)
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
