//Following Resources were found helpful in completing this task

//	https://tour.golang.org/
//	https://github.com/vladimirvivien/go-networking/tree/master/currency/servertxt0
//	https://golangcode.com/how-to-read-a-csv-file-into-a-struct/
//

// This program implements COVID-19 related data lookup service
// over TCP or Unix Data Socket. It loads data stored in a CSV
// file downloaded fromand uses a simple
// text-based protocol to interact with the user and send
// the data.
//
// Users send Date/Region search requests as a textual command in the form:
//
// <Region/Date> e.g. <Punjab> or <2020-03-14> or <KP>
//
// When the server receives the request, it is parsed and is then used
// to search the related data. The search result is then printed
// line-by-line back to the client.

// Testing:
// Netcat or telnet can be used to test this server by connecting and
// sending command using the format described above.

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var D []Data //Slice of Data Structs containing all the information

type Data struct {
	TestPos    string // # ofPositive Tests
	TestPer    string // # of Tests Performed
	Date       string //Date
	Discharged string // # of Patients Discharged
	Expired    string // # of Patients Expired
	Admitted   string // # of Patients Admitted
	Region     string //Region
}

func main() {
	var addr string
	var network string
	flag.StringVar(&addr, "e", ":4040", "service endpoint [ip addr or socket path]")
	flag.StringVar(&network, "n", "tcp", "network protocol [tcp,unix]")
	flag.Parse()

	// Open the file
	csvfile, err := os.Open("TimeSeries_KeyIndicators.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r, err := csv.NewReader(csvfile).ReadAll()
	for i, d := range r {
		if i > 0 {
			info := Data{
				TestPos:    d[0],
				TestPer:    d[1],
				Date:       d[2],
				Discharged: d[3],
				Expired:    d[4],
				Admitted:   d[5],
				Region:     d[6],
			}
			D = append(D, info)
		}
	}

	// validate supported network protocols
	switch network {
	case "tcp", "tcp4", "tcp6", "unix":
	default:
		log.Fatalln("unsupported network protocol:", network)
	}

	// create a listener for provided network and host address
	ln, err := net.Listen(network, addr)
	if err != nil {
		log.Fatal("failed to create listener:", err)
	}
	defer ln.Close()
	log.Println("**** Corona Virus Data [Pakistan] ***")
	log.Printf("Service started: (%s) %s\n", network, addr)
	// connection-loop - handle incoming requests
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			if err := conn.Close(); err != nil {
				log.Println("failed to close listener:", err)
			}
			continue
		}
		log.Println("Connected to", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

//Look for Given Input in the Data
func Find(param string) []Data {
	var res []Data

	for _, d := range D {
		if d.Date == param || d.Region == param {
			res = append(res, d)

		}
	}
	return res
}

func handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("error closing connection:", err)
		}
	}()

	if _, err := conn.Write([]byte("Connected to Server...!\nUsage: Type Region or Date\n[Punjab,Balochistan,ICT,KP,2020-04-04 etc.]\n")); err != nil {
		log.Println("error writing:", err)
		return
	}

	// loop to stay connected with client until client breaks connection
	for {
		// buffer for client command
		cmdLine := make([]byte, (1024 * 4))
		n, err := conn.Read(cmdLine)
		if n == 0 || err != nil {
			log.Println("connection read error:", err)
			return
		}

		param := string(cmdLine[:n-1])
		result := Find(param)
		if len(result) == 0 {
			if _, err := conn.Write([]byte("Invalid Input / Nothing found\n")); err != nil {
				log.Println("failed to write:", err)
			}
			continue
		}
		// send info line by line to the client with fmt.Sprintf()
		for _, cur := range result {
			_, err := conn.Write([]byte(
				fmt.Sprintf(
					"Positive: %s, Performed: %s, Date: %s, Discharged: %s, Expired: %s, Admitted: %s, Region: %s,\n",
					cur.TestPos, cur.TestPer, cur.Date, cur.Discharged, cur.Expired, cur.Admitted, cur.Region,
				),
			))
			if err != nil {
				log.Println("failed to write response:", err)
				return
			}
		}

	}
}
