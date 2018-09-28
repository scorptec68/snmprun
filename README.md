# snmp-simulator
An SNMP simulator with a language to specify the state changes for OIDs.


Example snmp program:

	var
	    printer-state 1.3.3.2.1.1.1 integer [1 = 'printing', 2 = 'idle', 3 = 'error', ]
	    page-count 1.3.5.56.6.6 counter
	    error-state 1.3.6.1.2.1.25.3.5.1.2.1 bitset [1 = 'no paper', 2 = 'paper jam', ]
	    model 1.3.5.6.6.7.7 string
	    cat 1.3.5.6.6.7.8.8 string
		x integer [1 = 'happy', 2 = 'sad']
		str string
		y boolean
	endvar
	
	run
	    model = "toshiba2555c"
	    page-count = 5
	    printer-state = 'idle'
	    error-state = []
		x = 'sad'
	
	    // now the core stuff
	    loop
	        delay 10 secs
	        printer-state = 'idle' 
	        printer-state = 2
	        error-state = ['no paper', 'paper jam'] 
	        error-state = [1, 6]
	
	        loop for 10
	          printer-state = 'printing'
	          page-count = page-count + 1
	          delay 20 secs
	        endloop
	
	    endloop
	endrun

Grammar using BNF notation
    <program> ::= <var-declaration> <run-declaration>

	<var-declaration> ::= var <var-list> endvar
	<var-list> :: <var-defn> | <var-list>
	
	<run-declaration> ::= run <statement-list> endrun
	<statement-list> ::= <statement> | <statement-list>
	
	<var-defn> ::= <identifier> <oid> <type> | <identifier> <oid> <type> <aliasset>
	<type> ::= integer | string | oid | bitset | counter | guage | timeticks | ipaddress
	
	<aliasset> ::= [ <number-token> = <alias>, <number-token> = <alias>, .... ]
	<atomicvalue> ::= <bitpos-token> <number-token> | <string>
	<value> ::= <atomicvalue> | <set> 
	
	<oid> ::= <oidtree | .<oidtree>
	<oidtree> ::= <number-token> | <number-token>.<oidtree>
	
	<string> ::= "chars-token"
	<alias> ::= 'chars-token'
	<set> ::= [ <itemlist> ]
	<itemlist> ::= <item> | <item>, <itemlist>
	<item> ::= <value> | <alias>
	<bitpos_token> := <number-token>
	
	<statement> := <assignment> | <loop> | <conditional> | <delay> | <print> | <comment>
	
	<assignment> ::= <identifier> = <expression>
	<loop> ::= loop [for <integer>] <statement-list> endloop
	
	<conditional> ::= if <condition> then <statement-list> endif
	<expression> ::= <int-expression> | <bool-expression> | <oid-expression> |  <str-expression> | <addr-expression> | <bitset-expression>

    <int-expression> ::= <int-term> + <int-expression> | <int-term> - <int-expression> | <int-term>
	<int-term> ::= <int-factor> * <int-term> | <int-factor> / <int-term>
	<int-factor> ::== <identifier> | <int-literal> | <alias> | - <int-factor> | ( <int-expression> )

    <str-expression> ::= <str-term> + <str-expression> | <str-term>
	<str-term> ::= <identifier> | <str-literal> | ( <str-expression>) |
					strInt(<int-expression>) | strBool(<bool-expression>) |
					strCounter(<int-expression>) | strOid(<oid-expression>) |
					strAddr(<addr-expression>) | strBitset(<bitset-expression>)

    <bool-expression> ::= <bool-term> \| <bool-expression> | <bool-term>
	<bool-term> ::= <bool-factor> & <bool-term> | <bool-factor>
	<bool-factor> ::= <identifier> | <bool-literal> | ( <bool-expression> ) 
	<bool-literal> ::= true | false

    <oid-expression> ::= <oid-term> + <oid-expression> | <oid-term>
	<oid-term> ::= <identifier> | <oid-literal> | ( <oid-expression> )

    <addr-expression> ::= <identifier> | <addr-literal>

    <bitset-expression> ::= <bitset-term> + <bitset-expression | <bitset-term> - <bitset-expression> | <bitset-term>
	<bitset-term> ::= <identifier> | <bitset-literal> | ( <bitset-expression> )

	<unit> ::= secs | msecs | hours | days | weeks
	<comment> ::= // <chars-to-newline-token>


