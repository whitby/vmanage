package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/whitby/vcapi"
)

// AD Account Enabled
const AccountEnabled = "512"

// AD Account Disabled
const AccountDisabled = "514"

// Status returns status over HTTP
type Status struct {
	Status string `json:"status"`
}

func (s *Status) Set(status string) {
	s.Status = status
}

var (
	wg             sync.WaitGroup
	ldapServer     string
	ldapPort       uint16 = 636
	baseDN         string
	attributes     = []string{"dn", "cn", "useraccountcontrol"}
	user           string
	passwd         string
	recent         bool
	debug          bool
	server         bool
	config         *vcapi.Config
	graduationYear int
	status         Status = Status{"OK"}
	// ErrNoResults no results from LDAP Search
	ErrNoResults = errors.New("Search did not return any results")
	// ErrTooManyResults = more than one result from LDAP Search
	ErrTooManyResults = errors.New("Search returned too many results")
)

func init() {
	ldapServer = os.Getenv("LDAP_SERVER")
	baseDN = os.Getenv("BASE_DN")
	user = os.Getenv("LDAP_ADMIN")
	if user == "" {
		user = fmt.Sprintf("cn=Administrator,cn=Users,%v", baseDN)
	}
	passwd = os.Getenv("LDAP_PASSWORD")
	flag.BoolVar(&recent, "recent", true, "Only check recent changes")
	flag.BoolVar(&debug, "debug", false, "debug ldap connection")
	flag.BoolVar(&server, "server", false, "run as daemon, with http server")
	// vcapi Configuration
	config = &vcapi.Config{
		Username: os.Getenv("VCAPI_USERNAME"), // API Username
		Password: os.Getenv("VCAPI_PASSWORD"), // API Password
		SchoolID: os.Getenv("VCAPI_SCHOOLID"), // Client, school name
	}
	apiversion := os.Getenv("VCAPI_VERSION")
	if apiversion != "" {
		// defaults to "v2"
		config.APIVersion = apiversion
	}
	y, err := strconv.Atoi(os.Getenv("GRADUATION_YEAR"))
	if err != nil {
		log.Fatal(err)
	}
	graduationYear = y

	flag.Parse()
}

func syncVCtoAD() {
	l, err := initLdap()
	if err != nil {
		status.Set("LDAP_CONN_FAILURE")
		log.Println("ERROR: %s\n", err.Error())
	}
	defer l.Close()

	// debug sets ldap to debug mode.
	if debug {
		l.Debug = true
	}

	// Create a new client with the above configuration.
	client := vcapi.NewClient(config)

	// Number of goroutines for wait group
	years := graduationYear - time.Now().Year()
	wg.Add(years + 2)
	log.Println("Checking VC and AD for changes.")

	// Enable all current students LDAP accounts
	go enableCurrentStudents(client, l)

	// Loop through all former students and disable LDAP accounts
	for y := graduationYear; y >= time.Now().Year(); y-- {
		go disableFormerStudents(client, l, y)
	}

	// wait for goroutines to sync up
	wg.Wait()
	status.Set("OK")
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsn, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsn)
}

func serve() {
	http.HandleFunc("/status", statusHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	if server {
		syncVCtoAD()
		go serve()
		tickChan := time.NewTicker(time.Minute * 15).C
		for {
			select {
			case <-tickChan:
				syncVCtoAD()
			}
		}
	} else {
		syncVCtoAD()
	}
}
