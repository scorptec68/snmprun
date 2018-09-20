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
	"time"

	"github.com/PromonLogicalis/asn1"
	"github.com/PromonLogicalis/snmp"
)

var logger *log.Logger

// Convert OID in string format to OID in uint slice format
func strToOID(str string) (oid asn1.Oid, err error) {
	str = strings.TrimPrefix(str, ".") // remove leading dot
	subStrings := strings.Split(str, ".")
	oid = make(asn1.Oid, len(subStrings))
	for i, componentStr := range subStrings {
		x, err := strconv.ParseUint(componentStr, 10, 32)
		if err != nil {
			return nil, err
		}
		oid[i] = uint(x)
	}
	return oid, nil
}

func convertBitsetToOctetStr(bitset BitsetMap) string {
	var maxK uint
	// get highest key in the set
	for k := range bitset {
		if k > maxK {
			maxK = k
		}
	}

	numBytes := maxK/8 + 1
	byteArr := make([]byte, numBytes)
	for k := range bitset {
		bytePos := k / 8
		bitPos := 7 - k%8
		byteArr[bytePos] |= 1 << bitPos
	}
	return string(byteArr)
}

func addOIDFunc(agent *snmp.Agent, interp *Interpreter, strOid string) {
	if len(strOid) == 0 {
		logger.Println("Empty oid")
		return
	}
	oid, err := strToOID(strOid)
	if err != nil {
		logger.Printf("Bad oid %v (%s) - should not happen\n", oid, strOid)
		return
	}

	agent.AddRoManagedObject(
		oid,
		func(oid asn1.Oid) (interface{}, error) {
			oidStr := oid.String()
			fmt.Printf("callback: oid: %s\n", oidStr)
			fmt.Printf("oid values: %v\n", interp.oid2Values)
			val, found := interp.GetValueForOid(oidStr)
			if !found {
				return nil, errors.New("Illegal Value")
			}
			switch val.valueType {
			case ValueBoolean:
				return val.boolVal, nil
			case ValueInteger:
				fmt.Printf("found int: %d\n", val.intVal)
				return val.intVal, nil
			case ValueString:
				return val.stringVal, nil
			case ValueBitset:
				return convertBitsetToOctetStr(val.bitsetVal), nil
			case ValueNone:
				return nil, errors.New("Illegal Value")
			}
			return nil, errors.New("Illegal Value")
		})
}

func initSNMPServer(interp *Interpreter) (agent *snmp.Agent, conn *net.UDPConn, err error) {
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

	fmt.Printf("oid2Values: %v\n", interp.oid2Values)
	for oidStr := range interp.oid2Values {
		addOIDFunc(agent, interp, oidStr)
	}

	return agent, conn, err
}

// Read from a channel about OID requests
func runSNMPServer(agent *snmp.Agent, conn *net.UDPConn, quit chan bool, wg *sync.WaitGroup) {
	const readTimeoutSecs = 5

	defer wg.Done()

	// Serve requests
	for {

		// stop if told to finish up
		select {
		case <-quit:
			return
		default:
			// Do other stuff
		}

		// read incoming PDU
		buffer := make([]byte, 1024)
		conn.SetReadDeadline(time.Now().Add(readTimeoutSecs * time.Second))
		n, source, err := conn.ReadFrom(buffer)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				// error but not a network error or a network error other than timeout
				// handle non-timeout error
				logger.Printf("Failed to read buffer: %s", err)
				os.Exit(1)
			}
			// timeout => test for quit or try read again
			continue
		}

		// process PDU
		buffer, err = agent.ProcessDatagram(buffer[:n])
		if err != nil {
			logger.Println(err)
			continue
		}

		// respond with a new PDU
		_, err = conn.WriteTo(buffer, source)
		if err != nil {
			logger.Printf("Failed to write buffer: %s", err)
			os.Exit(1)
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Print("Missing filename to run\n")
		os.Exit(1)
	}
	filename := os.Args[1]

	f, err := os.OpenFile(filename+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	logger = log.New(f, "snmpsim", log.LstdFlags)

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

	interp := new(Interpreter)
	interp.Init(program)

	agent, conn, err := initSNMPServer(interp)
	if err != nil {
		fmt.Printf("Failed to init snmp server: %s\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	quitServer := make(chan bool)
	// SNMP server running in background
	go runSNMPServer(agent, conn, quitServer, &wg)

	// now run program to set the OID values
	err = interp.InterpProgram(program)
	if err != nil {
		logger.Printf("Interpreting error: %s\n", err)
	}
	quitServer <- true

	wg.Wait()
}
