# snmp-simulator
An SNMP simulator with a language to specify the state changes for OIDs.


Example snmp program:

	var
	    printer-state 1.3.3.2.1.1.1 integer [1 = 'printing', 2 = 'idle', 3 = 'error']
	    page-count 1.3.5.56.6.6 counter
	    error-state 1.3.6.1.2.1.25.3.5.1.2.1 bitset [bit1 = 'no paper', bit2 = 'paper jam']
	    model 1.3.5.6.6.7.7 string
	    cat 1.3.5.6.6.7.8.8 string ["tabby" = 'tabby']
	endvar
	
	run
	    model = "toshiba2555c"
	    page-count = 5
	    printer-state = 'idle'
	    error-state = []
	
	    // now the core stuff
	    loop
	        delay 10 secs
	        printer-state = 'idle' 
	        printer-state = 2
	        error-state = ['no paper', 'paper jam'] 
	        error-state = [bit1, bit6]
	
	        loop for 10
	          printer-state = 'printing'
	          page-count = page-count + 1
	          delay 20 secs
	        endloop
	
	    endloop
	endrun

Grammar using BNF notation

	<var-declaration> ::= var <var-list> endvar
	<var-list> :: <var-defn> | <var-list>
	
	<run-declaration> ::= run <statement-list> endrun
	<statement-list> ::= <statement> | <statement-list>
	
	<var-defn> ::= <identifier> <oid> <type> | <identifier> <oid> <type> <aliasset>
	<type> ::= integer | string | oid | bitset | counter | guage | timeticks | ipaddress | null
	
	<aliasset> ::= [ <atomicvalue> = <alias>, <atomicvalue> = <alias>, .... ]
	<atomicvalue> ::= <bitpos-token> <number-token> | <string>
	<value> ::= <atomicvalue> | <set> 
	
	<oid> ::= <oidtree | inet.<oidtree>
	<oidtree> ::= <number-token> | <number-token>.<oidtree>
	
	<string> ::= "chars-token"
	<alias> ::= 'chars-token'
	<set> ::= [ <itemlist> ]
	<itemlist> ::= <item> | <item>, <itemlist>
	<item> ::= <value> | <alias>
	<bitpos_token> := bit[0-64]
	
	<statement> := <assignment> | <loop> | <conditional> | <delay> | <print> | <comment>
	
	<assignment> ::= <identifier> = <expression>
	<loop> ::= loop [for <integer>] <statement-list> endloop
	
	<conditional> ::= if <condition> then <statement-list> endif
	<expression> ::= <value> | <operation>
	<operation> ::= <expression> + <expression> | <expression> - <expression> | <expression> * <expression> | <expression> / <expression>
	<delay> ::= delay <expression> <unit>
	<unit> ::= secs | msecs | hours | days | weeks
	<comment> ::= // <chars-to-newline-token>


