package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/PromonLogicalis/asn1"
	"github.com/PromonLogicalis/snmp"
)

// Convert string to OID
func strToOID(str string) (oid asn1.Oid, err error) {
	subStrings := strings.Split(str, ".")
	oid = make(asn1.Oid, len(subStrings))
	for i, component := range subStrings {
		x, err := strconv.ParseUint(component, 10, 32)
		if err != nil {
			return nil, err
		}
		oid[i] = uint(x)
	}
	return oid, nil
}

func addOIDFunc(agent *snmp.Agent, val Value) {
	oid, err := strToOID(val.oid)
	if err != nil {
		fmt.Println("Bad oid - shouldn't happen")
	}

	agent.AddRoManagedObject(
		oid,
		func(oid asn1.Oid) (interface{}, error) {
			switch val.valueType {
			case ValueBoolean:
				return val.boolVal, nil
			case ValueInteger:
				return val.intVal, nil
			case ValueString:
				return val.stringVal, nil
			case ValueNone:
				return nil, errors.New("Illegal Value")
			}
			return nil, errors.New("Illegal Value")
		})
}

func initSNMPServer() (agent *snmp.Agent, conn *net.UDPConn, err error) {
	agent = snmp.NewAgent()

	// Set the read-only and read-write communities
	agent.SetCommunities("public", "private")

	// Bind to an UDP port
	addr, err := net.ResolveUDPAddr("udp", ":161")
	if err != nil {
		return nil, nil, err
	}
	conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return nil, nil, err
	}

	return agent, conn, err
}

func processValues(agent *snmp.Agent, values <-chan Value, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		value := <-values
		if value.valueType == ValueNone {
			break
		}
		fmt.Printf("Value %v was read.\n", value)
		addOIDFunc(agent, value)
	}
}

// Read from a channel about OID requests
func runSNMPServer(agent *snmp.Agent, conn *net.UDPConn,
	values <-chan Value, wg *sync.WaitGroup) {

	defer wg.Done()

	// Serve requests
	for {
		buffer := make([]byte, 1024)
		n, source, err := conn.ReadFrom(buffer)
		if err != nil {
			fmt.Errorf("Failed to read buffer: %s", err)
			os.Exit(1)
		}

		// Problem is that interpreter can produce a bunch of values
		// and we won't process them until we get a request
		// to our snmp server

		buffer, err = agent.ProcessDatagram(buffer[:n])
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = conn.WriteTo(buffer, source)
		if err != nil {
			fmt.Errorf("Failed to write buffer: %s", err)
			os.Exit(1)
		}
	}
}

// Program will run and will modify variables.
func runProgram(prog *Program, values chan<- Value, wg *sync.WaitGroup) {

	defer wg.Done()
	interp := new(Interpreter)
	err := interp.InterpProgram(prog, values)
	if err != nil {
		fmt.Printf("Interpreting error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Print("Missing filename to run")
		os.Exit(1)
	}
	filename := os.Args[1]

	inputBuf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Unable to read file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lex(filename, string(inputBuf))

	parser := NewParser(l)
	program, err := parser.ParseProgram()
	if err != nil {
		fmt.Printf("Parsing error: %s\n", err)
		os.Exit(1)
	}

	//PrintProgram(program, 0)
	//os.Exit(0)

	agent, conn, err := initSNMPServer()
	if err != nil {
		fmt.Printf("Failed to init snmp server: %s\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(3)

	// could use a channel to communicate - yeah
	// channel can emit an item when we have an assignment to an oid variable
	values := make(chan Value)
	go runProgram(program, values, &wg)

	//TODO: what do I do about 2 goroutines accessing agent
	// how do I synchronize?
	go processValues(agent, values, &wg)
	go runSNMPServer(agent, conn, values, &wg)

	wg.Wait()
}
