package main

import (
	"github.com/go-test/deep"
	"testing"
)

// helper
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Test_GatherHostnamesOrIps(t *testing.T) {
	type testCase struct {
		DBAddr         string
		ConnString     string
		HostnamesOrIps []string
		Valid          bool
	}

	var testCases = []testCase{
		{
			"tcp://clickhouse1:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=clickhouse2:9000,clickhouse3:9000",
			"tcp://%s?database=marketdata&read_timeout=5&write_timeout=5",
			[]string{"clickhouse1:9000", "clickhouse2:9000", "clickhouse3:9000"},
			true,
		},
		{
			"tcp://clickhouse1:9000?database=marketdata&read_timeout=5&write_timeout=5",
			"tcp://%s?database=marketdata&read_timeout=5&write_timeout=5",
			[]string{"clickhouse1:9000"},
			true,
		},
		{
			"tcp://10.1.23.45:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=clickhouse2:9000,10.0.0.1:9000",
			"tcp://%s?database=marketdata&read_timeout=5&write_timeout=5",
			[]string{"10.1.23.45:9000", "clickhouse2:9000", "10.0.0.1:9000"},
			true,
		},
		{
			"tcp://10.1.23.45:9000?write_timeout=5&database=testing&alt_hosts=clickhouse2:9000,10.0.0.1:9000,10.0.0.2:9200,10.0.0.3:4555,10.0.0.4:5656",
			"tcp://%s?write_timeout=5&database=testing",
			[]string{"10.1.23.45:9000", "clickhouse2:9000", "10.0.0.1:9000", "10.0.0.2:9200", "10.0.0.3:4555", "10.0.0.4:5656"},
			true,
		},
		{
			"tcp://davekoston.com:9000?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=localhost:9000",
			"tcp://%s?database=marketdata&read_timeout=5&write_timeout=5",
			[]string{"davekoston.com:9000", "localhost:9000"},
			true,
		},
		{
			"tcp://fred?database=marketdata&read_timeout=5&write_timeout=5&alt_hosts=clickhouse2:9000,10.0.0.1:9000,10.0.0.2:9200,10.0.0.3:4555,10.0.0.4:5656",
			"tcp://%s?database=marketdata&read_timeout=5&write_timeout=5",
			[]string{"10.1.23.45", "clickhouse2", "10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"},
			false,
		},
		{
			"tcp://clickhouse1:9000",
			"tcp://%s?",
			[]string{"clickhouse1:9000"},
			false,
		},
		{
			"tcp:/clickhouse1:9000",
			"tcp:/%s?",
			[]string{"clickhouse1:9000"},
			false,
		},
	}

	for i := 0; i < len(testCases); i++ {
		hostnamesOrIps, connString, err := ExtractHostsFromConnectionString(testCases[i].DBAddr)

		if !testCases[i].Valid {
			if err == nil {
				t.Errorf("Expected test case (%d) to be invalid. %v", i, hostnamesOrIps)
			}
			continue
		}

		if diff := deep.Equal(hostnamesOrIps, testCases[i].HostnamesOrIps); diff != nil {
			t.Errorf("Unable to gather hostnames/IPs. test case (%d). Diff: %v. Found: %v", i, diff, hostnamesOrIps)
		}

		if connString != testCases[i].ConnString {
			t.Errorf("Unable to get connString. Wanted(%s) Found(%s)", testCases[i].ConnString, connString)
		}
	}
}

func Test_isIP(t *testing.T) {
	type IPTest struct {
		HostOrIP string
		IP       bool
	}

	var testCases = []IPTest{
		{"www.google.com", false},
		{"foundationdb", false},
		{"127.0.0.1", true},
		{"localhost", false},
		{"10.253.1.24", true},
	}

	for i := 0; i < len(testCases); i++ {
		ip := isIP(testCases[i].HostOrIP)

		if ip != testCases[i].IP {
			t.Errorf("Expected %s to be IP: %t, Got: %t", testCases[i].HostOrIP, testCases[i].IP, ip)
		}
	}
}

func Test_TranslateHostArrayToIPs(t *testing.T) {
	type HostArrayTest struct {
		HostArray []string
		Expected  [][]string
	}

	one := make([][]string, 1)
	one = append(one, []string{"127.0.0.1:4500:tcp"})

	two := make([][]string, 1)
	two = append(one, []string{"127.0.0.1:9999"})

	three := make([][]string, 1)
	three = append(three, []string{"104.27.130.63:1234", "127.0.0.1:9999"})
	three = append(three, []string{"104.27.131.63:1234", "127.0.0.1:9999"})

	var testCases = []HostArrayTest{
		{[]string{"127.0.0.1:4500:tcp"}, one},
		{[]string{"127.0.0.1:9999"}, two},
		{[]string{"davekoston.com:1234", "127.0.0.1:9999"}, three},
	}

	for i := 0; i < len(testCases); i++ {
		translated := TranslateHostArrayToIPs(testCases[i].HostArray)

		found := false

		for j := 0; j < len(testCases[i].Expected); j++ {
			if diff := deep.Equal(translated, testCases[i].Expected[j]); diff == nil {
				found = true
			}
		}

		if !found {
			t.Errorf("Expected %s to be translated to : %s, Got: %s", testCases[i].HostArray, testCases[i].Expected, translated)
		}
	}
}

func Test_TranslateHostToIP(t *testing.T) {
	type HostTest struct {
		Host     string
		Expected []string
	}

	var testCases = []HostTest{
		{"localhost:1234", []string{"127.0.0.1:1234"}},
		{"127.0.0.1:4500:tcp", []string{"127.0.0.1:4500:tcp"}},
		{"127.0.0.1:9999", []string{"127.0.0.1:9999"}},
		{"davekoston.com:1234", []string{"104.27.130.63:1234", "104.27.131.63:1234"}},
		{"davekoston.com:4321:tcp", []string{"104.27.130.63:4321:tcp", "104.27.131.63:4321:tcp"}},
	}

	for i := 0; i < len(testCases); i++ {
		translated := TranslateHostToIP(testCases[i].Host)

		if !stringInSlice(translated, testCases[i].Expected) {
			t.Errorf("Expected %s to be translated to : %s, Got: %s", testCases[i].Host, testCases[i].Expected, translated)
		}
	}
}
