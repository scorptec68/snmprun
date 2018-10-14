# SNMPrun simulator
An SNMP simulator with a language to specify the state changes for OIDs.

##Hello world program:

```
var
   hello: .1.2.3 string
endvar

run
   hello = "G'day, mate."
   print hello
   sleep 10 secs
endrun
```

Running the SNMP simulator and running the snmpwalk client program.

```
server>$ sudo ./snmprun examples/hello.sim
G'day, mate.

client>$ snmpwalk -c public -v1 localhost .1
iso.2.3 = STRING: "G'day, mate."
End of MIB
```

##Variables program

```
var
    str: .1.1.1 string
    i1: .1.1.2 integer
    i2: .1.2.2 integer
    cnt: .1.1.3 counter
    bool: .1.1.4 boolean
    oid1: .1.1.5 oid
    ticks: .1.1.6 timeticks
    g: .1.1.7 guage
    ip: .1.1.8 ipaddress
    bits: .1.1.9 bitset [ 1 = 'good', 2 = 'bad', 3 = 'ugly']
endvar
run
    str = "hi " + "there"
    print "str = " + str

    i1 = 1 
    i2 = 2
    i2 = 3 * i2 - i1
    print "int = " + strInt(i2)

    cnt = 4
    print "cnt = " + strCnt(cnt)

    bool = true
    print "bool = " + strBool(bool)

    oid1 = .1.1
    oid1 = oid1 + .1
    print "oid = " + strOid(oid1)

    ticks = 1000
    print "ticks = " + strTimeticks(ticks)

    g = 42
    print "guage = " + strGuage(g)

    ip = 127.0.0.1
    print "ip = " + strIpaddress(ip)

    bits = [ 'good', 'ugly']
    bits = bits - [ 'ugly']
    bits = bits + [ 'bad']
    print "bits = " + strBitset(bits)

endrun
```

Running SNMP simulator:

```
$ sudo ./snmprun examples/variables.sim 
str = hi there
int = 5
cnt = 4
bool = true
oid = .1.1.2
ticks = 1000
guage = 42
ip = 127.0.0.1
bits = {1, 3}
```

##Printer printing pages with errors program

```
var
  printer-status: 2.1.25.3.5.1.1.1 integer [1 = 'other', 2 = 'p-unknown', 3 = 'idle', 4 = 'printing',
                                        5 = 'warmup', ]
  device-status: 2.1.25.3.2.1.5.1 integer [1 = 'd-unknown', 2 = 'running', 3 = 'warning', 
                                       4 = 'testing', 5 = 'down' ]
  error-state: 2.1.25.3.5.1.2.1 bitset [0 = 'low paper', 1 = 'no paper', 2 = 'low toner',
                                   3 = 'no toner', 4 = 'door open', 5 = 'jammed',
                                   6 = 'offline', 7 = 'service requested', 
                                   8 = 'input tray missing',
                                   9 = 'output tray missing',
                                   10 = 'marker supply missing',
                                   11 = 'output near full',
                                   12 = 'output full',
                                   13 = 'input tray empty',
                                   14 = 'overdue prvent maint',]
  host-time: 2.1.25.1.1.0 timeticks
  sys-object: 2.1.1.2.0 oid
  device-desc1: 2.1.25.3.2.1.3.1 string
  marker-count: 2.1.43.10.2.1.4.1.1 counter
  do-color: boolean
  host: .1.3.6.1.2.1.4.20.1.1.10.100.63.22 ipaddress
  tosh-color:  4.1.1129.2.3.50.1.3.21.6.1.3.1.1 counter
endvar

run
    device-desc1 = "Toshiba 2555c"
    tosh-color = 400
    sys-object = .1.3.6.1.4.1.1129.2.3.45.1
    host-time = 21851051
    host = 192.168.1.1
    printer-status = 'idle'
    device-status = 'running'
    marker-count = 1042
    error-state = ['output near full', 'low toner']

    print "setting print-state to " + strInt(printer-status)
    print "setting device-state to " + strInt(device-status)

    sleep 2 secs
    error-state = error-state - ['low toner']

    sleep 2 secs
    printer-status = 'printing'

    loop times 5
        marker-count = marker-count + 1

        // every 2nd page is color
        if do-color
           tosh-color = tosh-color + 1
           do-color = false
        else
           do-color = true
        endif

        sleep 2 secs
    endloop

    error-state = error-state + ['low paper']
    printer-status = 'idle'

    sleep 10 secs
endrun
```

Running simulator example:
```
client>$ while true; do date; snmpwalk -c public -v1 localhost .1; sleep 2; done
Sat 13 Oct 2018 15:34:26 AEDT
SNMPv2-MIB::sysObjectID.0 = OID: SNMPv2-SMI::enterprises.1129.2.3.45.1
IP-MIB::ipAdEntAddr.10.100.63.22 = IpAddress: 192.168.1.1
HOST-RESOURCES-MIB::hrSystemUptime.0 = Timeticks: (21851051) 2 days, 12:41:50.51
HOST-RESOURCES-MIB::hrDeviceDescr.1 = STRING: Toshiba 2555c
HOST-RESOURCES-MIB::hrDeviceStatus.1 = INTEGER: running(2)
HOST-RESOURCES-MIB::hrPrinterStatus.1 = INTEGER: idle(3)
HOST-RESOURCES-MIB::hrPrinterDetectedErrorState.1 = Hex-STRING: 20 10 
SNMPv2-SMI::mib-2.43.10.2.1.4.1.1 = Counter32: 1042
SNMPv2-SMI::enterprises.1129.2.3.50.1.3.21.6.1.3.1.1 = Counter32: 400
End of MIB
Sat 13 Oct 2018 15:34:28 AEDT
SNMPv2-MIB::sysObjectID.0 = OID: SNMPv2-SMI::enterprises.1129.2.3.45.1
IP-MIB::ipAdEntAddr.10.100.63.22 = IpAddress: 192.168.1.1
HOST-RESOURCES-MIB::hrSystemUptime.0 = Timeticks: (21851051) 2 days, 12:41:50.51
HOST-RESOURCES-MIB::hrDeviceDescr.1 = STRING: Toshiba 2555c
HOST-RESOURCES-MIB::hrDeviceStatus.1 = INTEGER: running(2)
HOST-RESOURCES-MIB::hrPrinterStatus.1 = INTEGER: idle(3)
HOST-RESOURCES-MIB::hrPrinterDetectedErrorState.1 = Hex-STRING: 00 10 
SNMPv2-SMI::mib-2.43.10.2.1.4.1.1 = Counter32: 1042
SNMPv2-SMI::enterprises.1129.2.3.50.1.3.21.6.1.3.1.1 = Counter32: 400
End of MIB
Sat 13 Oct 2018 15:34:30 AEDT
SNMPv2-MIB::sysObjectID.0 = OID: SNMPv2-SMI::enterprises.1129.2.3.45.1
IP-MIB::ipAdEntAddr.10.100.63.22 = IpAddress: 192.168.1.1
HOST-RESOURCES-MIB::hrSystemUptime.0 = Timeticks: (21851051) 2 days, 12:41:50.51
HOST-RESOURCES-MIB::hrDeviceDescr.1 = STRING: Toshiba 2555c
HOST-RESOURCES-MIB::hrDeviceStatus.1 = INTEGER: running(2)
HOST-RESOURCES-MIB::hrPrinterStatus.1 = INTEGER: printing(4)
HOST-RESOURCES-MIB::hrPrinterDetectedErrorState.1 = Hex-STRING: 00 10 
SNMPv2-SMI::mib-2.43.10.2.1.4.1.1 = Counter32: 1043
SNMPv2-SMI::enterprises.1129.2.3.50.1.3.21.6.1.3.1.1 = Counter32: 400
End of MIB
Sat 13 Oct 2018 15:34:32 AEDT
SNMPv2-MIB::sysObjectID.0 = OID: SNMPv2-SMI::enterprises.1129.2.3.45.1
IP-MIB::ipAdEntAddr.10.100.63.22 = IpAddress: 192.168.1.1
HOST-RESOURCES-MIB::hrSystemUptime.0 = Timeticks: (21851051) 2 days, 12:41:50.51
HOST-RESOURCES-MIB::hrDeviceDescr.1 = STRING: Toshiba 2555c
HOST-RESOURCES-MIB::hrDeviceStatus.1 = INTEGER: running(2)
HOST-RESOURCES-MIB::hrPrinterStatus.1 = INTEGER: printing(4)
HOST-RESOURCES-MIB::hrPrinterDetectedErrorState.1 = Hex-STRING: 00 10 
SNMPv2-SMI::mib-2.43.10.2.1.4.1.1 = Counter32: 1044
SNMPv2-SMI::enterprises.1129.2.3.50.1.3.21.6.1.3.1.1 = Counter32: 401
End of MIB
```

##Grammar using BNF notation
```
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
```

