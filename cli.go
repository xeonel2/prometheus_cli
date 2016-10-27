// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"bytes"
	"os"
	"strconv"
	"strings"
	"time"
	"gopkg.in/yaml.v2"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
)

var (
	server   = flag.String("server", "", "URL of the Prometheus server to query")
	timeout  = flag.Duration("timeout", time.Minute, "Timeout to use when querying the Prometheus server")
	useCSV   = flag.Bool("csv", true, "Whether to format output as CSV")
	csvDelim = flag.String("csvDelimiter", ";", "Single-character delimiter to use in CSV output")
	config = flag.String("config", "uptime.yml", "Config file path")
)

type conf struct {
	Server string `yaml:"server"`
	Timezone string `yaml:"timezone"`
	EmailTo string `yaml:"emailto"`
	EmailFrom string `yaml:"emailfrom"`
	EmailSubject string `yaml:"emailsubject"`
	SmtpHost string `yaml:"smtphost"`
	SmtpPort string `yaml:"smtpport"`
	SmtpUser string `yaml:"smtpuser"`
	SmtpPwd string `yaml:"smtppwd"`
	ShowCount bool `yaml:"showcount"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	EndpointName string `yaml:"name"`
	EndpointFailedQuery string `yaml:"failed"`
	EndpointSuccessQuery string `yaml:"success"`
}

func (c *conf) getConf(configfile string) *conf {

	yamlFile, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr, "")
	os.Exit(1)
}

func queryToString(r QueryResponse) string {
	if *useCSV {
		return fmt.Sprint(r.ToCSV(rune((*csvDelim)[0])))
	}
	return fmt.Sprint(r.ToText())
}


func printQueryResponse(r QueryResponse) {
	fmt.Print(queryToString(r))
}

func query(c *Client, endpoint Endpoint, ShowCount bool) {
	//fmt.Println("called query for", qry)
	efailqry := endpoint.EndpointFailedQuery
	esuccqry := endpoint.EndpointSuccessQuery
	ename := endpoint.EndpointName
	var xx2 float64= 0
	var xx5 float64= 0
	if (efailqry=="" || esuccqry=="") {
		die("Error in config. Check uptime.yml")
	}
	sresp, err := c.Query(esuccqry)
	if err != nil {
		die("Error querying server: %s", err)
	}
	if(sresp == nil || len(queryToString(sresp)) == 0 ){
		xx2=0
	}else{
		xx2, err=strconv.ParseFloat(strings.Split(queryToString(sresp),";")[1], 64)
		if err != nil {
			die("Error parsing query: %s", err)
		}
	}

	fresp, err := c.Query(efailqry)
	if err != nil {
		die("Error querying server: %s", err)
	}
	if(fresp == nil || len(queryToString(fresp)) == 0 ){
		xx5=0
	}else{
		xx5, err=strconv.ParseFloat(strings.Split(queryToString(fresp),";")[1], 64)
		if err != nil {
			die("Error parsing query: %s", err)
		}
	}
	if(xx2==0 && xx5==0){
		fmt.Println(ename)
		fmt.Println("Endpoint not used!")
		MailBuffer.WriteString("\n<td>")
		MailBuffer.WriteString(ename)
		MailBuffer.WriteString("\n</td>")
		if(ShowCount) {
			MailBuffer.WriteString("\n<td>")
			MailBuffer.WriteString("0")
			MailBuffer.WriteString("\n</td>")
			MailBuffer.WriteString("\n<td>")
			MailBuffer.WriteString("0")
			MailBuffer.WriteString("\n</td>")
		}
		MailBuffer.WriteString("\n<td>")
		MailBuffer.WriteString("100%")
		MailBuffer.WriteString("\n</td>")

	}else {
		fmt.Println(ename)
		fmt.Println("2xx:", xx2)
		fmt.Println("5xx:", xx5)
		fmt.Println("uptime:", (xx2 / (xx2 + xx5)) * 100, "%")
		MailBuffer.WriteString("\n<td>")
		MailBuffer.WriteString(ename)
		MailBuffer.WriteString("\n</td>")
		if(ShowCount) {
			MailBuffer.WriteString("\n<td>")
			MailBuffer.WriteString(fmt.Sprint(xx2))
			MailBuffer.WriteString("\n</td>")
			MailBuffer.WriteString("\n<td>")
			MailBuffer.WriteString(fmt.Sprint(xx5))
			MailBuffer.WriteString("\n</td>")
		}
		MailBuffer.WriteString("\n<td>")
		MailBuffer.WriteString(fmt.Sprint((xx2 / (xx2 + xx5)) * 100))
		MailBuffer.WriteString("%\n")
		MailBuffer.WriteString("\n</td>")
	}

}

func queryRange(c *Client) {
	if flag.NArg() != 4 && flag.NArg() != 5 {
		flag.Usage()
		die("Wrong number of range query arguments")
	}
	end, err := strconv.ParseFloat(flag.Arg(2), 64)
	if err != nil {
		flag.Usage()
		die("Invalid end timestamp '%s'", flag.Arg(2))
	}
	rangeSec, err := strconv.ParseUint(flag.Arg(3), 10, 64)
	if err != nil {
		flag.Usage()
		die("Invalid query range '%s'", flag.Arg(3))
	}
	var step uint64
	if flag.NArg() == 5 {
		step, err = strconv.ParseUint(flag.Arg(4), 10, 64)
		if err != nil {
			flag.Usage()
			die("Invalid query step '%s'", flag.Arg(4))
		}
	} else {
		step = rangeSec / 250
	}
	if step < 1 {
		step = 1
	}

	resp, err := c.QueryRange(flag.Arg(1), end, rangeSec, step)
	if err != nil {
		die("Error querying server: %s", err)
	}

	printQueryResponse(resp)
}

func metrics(c *Client) {
	if flag.NArg() != 1 {
		flag.Usage()
		die("Too many arguments")
	}

	if metrics, err := c.Metrics(); err != nil {
		die("Error querying server: %s", err)
	} else {
		for _, m := range metrics {
			fmt.Println(m)
		}
	}
}

func usage() {
	//fmt.Fprintf(os.Stderr, "Usage:\n")
	//fmt.Fprintf(os.Stderr, "\t%s [flags] query <expression>\n", os.Args[0])
	//fmt.Fprintf(os.Stderr, "\t%s [flags] query_range <expression> <end_timestamp> <range_seconds> [<step_seconds>]\n", os.Args[0])
	//fmt.Fprintf(os.Stderr, "\t%s [flags] metrics\n", os.Args[0])
	//fmt.Printf("\nFlags:\n")
	//flag.PrintDefaults()
}

var MailBuffer *bytes.Buffer
func main() {
	var con conf
	flag.Parse()
	con.getConf(*config)
	var ShowCount=con.ShowCount
	var ql=len(con.Endpoints)
	if con.Server == "" {
		die("Server name not present. Check uptime.yml")
	}
	if ql==0 {
		die("Endpoint not present. Check uptime.yml")
	}

	loc ,e := time.LoadLocation(con.Timezone)
	if e!=nil{}
	istdate := fmt.Sprint((time.Now().In(loc)).Format("2006-01-02"))

	MailBuffer = bytes.NewBufferString("<html>")
	MailBuffer.WriteString("\n<head>")
	MailBuffer.WriteString(fmt.Sprintf("\n<h1><b>Uptime Percentile for %s:</b></h1>\n",istdate))
	MailBuffer.WriteString("\n<style>")
	MailBuffer.WriteString("\ntable {")
	MailBuffer.WriteString("\n    font-family: arial, sans-serif;")
	MailBuffer.WriteString("\n    border-collapse: collapse;")
	MailBuffer.WriteString("\n    width: 100%;")
	MailBuffer.WriteString("\n}")
	MailBuffer.WriteString("\n")
	MailBuffer.WriteString("\ntd, th {")
	MailBuffer.WriteString("\n    border: 1px solid #dddddd;")
	MailBuffer.WriteString("\n    text-align: left;")
	MailBuffer.WriteString("\n    padding: 8px;")
	MailBuffer.WriteString("\n}")
	MailBuffer.WriteString("\n")
	MailBuffer.WriteString("\ntr:nth-child(even) {")
	MailBuffer.WriteString("\n    background-color: #dddddd;")
	MailBuffer.WriteString("\n}")
	MailBuffer.WriteString("\n</style>\n")
	MailBuffer.WriteString("\n</head>")
	MailBuffer.WriteString("\n<body>")
	MailBuffer.WriteString("\n<table>")
	MailBuffer.WriteString("\n<tr>")
	MailBuffer.WriteString("\n<th>Endpoint</th>")
	if(ShowCount) {
		MailBuffer.WriteString("\n<th>2XX</th>")
		MailBuffer.WriteString("\n<th>5XX</th>")
	}
	MailBuffer.WriteString("\n<th>Uptime</th>")
	MailBuffer.WriteString("\n</tr>")
	c := NewClient(con.Server, *timeout)
	for _,element := range con.Endpoints {
		MailBuffer.WriteString("\n<tr>")
		query(c,element,ShowCount)
		MailBuffer.WriteString("\n</tr>")
	}

	MailBuffer.WriteString("\n</table>")
	MailBuffer.WriteString("\n</body>")
	MailBuffer.WriteString("\n</html>")
	SubjectBuffer := bytes.NewBufferString(con.EmailSubject)
	m := gomail.NewMessage()
	m.SetAddressHeader("From", con.EmailFrom, "Uptime from Prometheus")
	m.SetAddressHeader("To", con.EmailTo, con.EmailTo)
	SubjectBuffer.WriteString(fmt.Sprintf(" for %s", istdate ))
	m.SetHeader("Subject", SubjectBuffer.String())
	m.SetBody("text/html", MailBuffer.String())
	porty, err := strconv.Atoi(con.SmtpPort)
	if err != nil {
		die("SMTP Error in config(uptime.yml):%s",err)
	}
	d := gomail.NewDialer(con.SmtpHost, porty, con.SmtpUser, con.SmtpPwd)
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
	fmt.Println(MailBuffer.String())

}
