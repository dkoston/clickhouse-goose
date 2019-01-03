package main

import (
	"github.com/jessevdk/go-flags"
	"log"
	"net"
	"strings"
	"errors"
	"os"
	"fmt"
	"os/exec"
)

type CommandLineOptions struct {
	DBAddr     string `long:"db_addr" description:"clickhouse db connection address" env:"DB_ADDR" default:"tcp://localhost:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=localhost:9001,localhost:9002"`
	GooseEnv 	string `long:"goose_env" description:"goose environment to execute against" env:"GOOSE_ENV" default:"development"`
	Verbose bool `long:"verbose" description:"verbose output" env:"VERBOSE"`
}

const name = "clickhouse-goose"
const version = "0.0.1"

func main() {
	var opts CommandLineOptions
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatalf("[%s] Unable to parse command line options: %v", name, err)
	}

	hostNamesOrIps, connString, err := ExtractHostsFromConnectionString(opts.DBAddr)
	if err != nil {
		log.Fatal(err)
	}
	ips := TranslateHostArrayToIPs(hostNamesOrIps)
	if opts.Verbose {
		log.Printf("IPS: %v", ips)
		log.Printf("Conn String: %s", connString)
	}

	for i := 0; i < len(ips); i++ {
		RunGoose(ips[i], connString, opts.GooseEnv, opts.Verbose)
	}

	// Put DB_ADDR back
	os.Setenv("DB_ADDR", opts.DBAddr)
	dbAddr := os.Getenv("DB_ADDR")
	if opts.Verbose {
		log.Printf("DB_ADDR: %s", dbAddr)
	}
}

func isIP(ip string) bool {
	ip = strings.Trim(ip, " ")
	r := net.ParseIP(ip)

	if r != nil {
		return true
	}
	return false
}

// tcp://localhost:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=localhost:9001,localhost:9002

func ExtractHostsFromConnectionString(dbAddr string) (hosts []string, connString string, err error) {
	// tcp://clickhouse1:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=clickhouse2:9000,clickhouse3:9000

	var hostnamesOrIps []string
	connString = "tcp://%s?"

	parts := strings.Split(dbAddr, "?")

	if len(parts) != 2 {
		return hostnamesOrIps, connString, errors.New("invalid connection string. multiple or no ? signs")
	}

	firstHostParts := strings.Split(parts[0], "//")

	if len(firstHostParts) != 2 {
		return hostnamesOrIps, connString, errors.New("invalid connection string. Invalid primary host")
	}

	firstHostname := firstHostParts[1]
	if !strings.Contains(firstHostname, ":") {
		return hostnamesOrIps, connString, errors.New("invalid connection string. No port defined for primary host")
	}
	hostnamesOrIps = append(hostnamesOrIps, firstHostname)

	queryStringParts := strings.Split(parts[1], "&")

	var nonAltHostParts []string

	// find alt hosts
	for i := 0; i < len(queryStringParts); i++ {
		if strings.HasPrefix(queryStringParts[i], "alt_hosts=") {
			altHostsParts := strings.Split(dbAddr, "alt_hosts=")
			if len(altHostsParts) > 1 {
				altHosts := strings.Split(altHostsParts[1], ",")

				for j := 0; j < len(altHosts); j++ {
					hostnamesOrIps = append(hostnamesOrIps, altHosts[j])
				}
			}
		} else {
			nonAltHostParts = append(nonAltHostParts, queryStringParts[i])
		}
	}

	connString += strings.Join(nonAltHostParts, "&")

	return hostnamesOrIps, connString, nil
}


func TranslateHostArrayToIPs(addrArray []string) []string {
	for i := 0; i < len(addrArray); i++ {
		addrArray[i] = TranslateHostToIP(addrArray[i])
	}
	return addrArray
}

func TranslateHostToIP(addr string) string {
	parts := strings.Split(addr, ":")
	hostnameOrIP := parts[0]

	if hostnameOrIP == "localhost" {
		parts[0] = "127.0.0.1"
		hostnameOrIP = "127.0.0.1"
	}

	if !isIP(hostnameOrIP) {
		addr, err := net.LookupIP(hostnameOrIP)
		if err != nil {
			log.Fatalf("Invalid Hostname: %v", err)
		}
		parts[0] = addr[0].String()
	}
	return strings.Join(parts, ":")
}

func RunGoose(ipAndPort string, connString string, gooseEnv string, verbose bool) {
	// export DB_ADDR
	os.Setenv("DB_ADDR", fmt.Sprintf(connString, ipAndPort))
	dbAddr := os.Getenv("DB_ADDR")
	if verbose {
		log.Printf("DB_ADDR: %s", dbAddr)
	}
	// run goose
	cmd := exec.Command("goose", "-env", gooseEnv, "up" )
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("failed running `goose up` failed with: %s\n", err)
	}
	if verbose {
		log.Printf("Goose:\n%s\n", string(out))
	}

}