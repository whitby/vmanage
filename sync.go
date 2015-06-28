package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/go-ldap/ldap"
	"github.com/whitby/vcapi"
)

// Search LDAP by cn filter
func searchByName(l *ldap.Conn, name string) (*ldap.SearchResult, error) {
	filter := fmt.Sprintf("(cn=%v)", ReplaceAccents(name))
	search := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil)

	sr, err := l.Search(search)
	if err != nil {
		return nil, err
	}
	switch {
	case len(sr.Entries) == 0:
		return sr, ErrNoResults
	case len(sr.Entries) > 1:
		return sr, ErrTooManyResults
	}
	return sr, nil
}

// searchDisable disables an LDAP account
func searchAndDisable(l *ldap.Conn, name string) error {
	sr, err := searchByName(l, name)
	if err != nil {
		return err
	}
	entry := sr.Entries[0]
	return modify(l, entry, AccountDisabled)
}

// searchAndEnable enables an LDAP account
func searchAndEnable(l *ldap.Conn, name string) error {
	sr, err := searchByName(l, name)
	if err != nil {
		return err
	}
	entry := sr.Entries[0]
	return modify(l, entry, AccountEnabled)
}

// modify enables or disables an LDAP account
func modify(l *ldap.Conn, entry *ldap.Entry, action string) error {
	useraccountcontrol := entry.Attributes[1].Values[0]
	if useraccountcontrol != action {
		modify := ldap.NewModifyRequest(entry.DN)
		modify.Replace("useraccountcontrol", []string{action})
		if err := l.Modify(modify); err != nil {
			log.Println("ERROR: %s\n", err.Error())
			return err
		}
		cn := entry.Attributes[0].Values[0]
		logMod(cn, action)
	}
	return nil
}

// logMod logs modifications to an account
func logMod(cn string, action string) {
	switch action {
	case AccountEnabled:
		log.Printf("Enabled LDAP Account for %v", cn)
	case AccountDisabled:
		log.Printf("Disabled LDAP Account for %v", cn)
	}
}

func initLdap() (*ldap.Conn, error) {
	l, err := ldap.DialTLS(
		"tcp",
		fmt.Sprintf("%s:%d", ldapServer, ldapPort),
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return nil, err
	}
	err = l.Bind(user, passwd)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func formerStudents(client *vcapi.Client, opt *vcapi.ListOptions) ([]vcapi.Alumni, error) {
	if recent {
		return client.Alumni.Recent(opt)
	}
	return client.Alumni.List(opt)
}

func currentStudents(client *vcapi.Client, opt *vcapi.ListOptions) ([]vcapi.Student, error) {
	if recent {
		return client.Students.Recent(opt)
	}
	return client.Students.List(opt)
}

// Disable LDAP Accounts for alumni of a certain graduation year.
// Ignores any account not found in LDAP directory
func disableFormerStudents(client *vcapi.Client, l *ldap.Conn, year int) {
	defer wg.Done()
	opt := &vcapi.ListOptions{
		Params: vcapi.Params{
			"graduation_year": fmt.Sprintf("%v", year),
			"option":          "2",
		}}
	for {
		alumni, err := formerStudents(client, opt)
		if err != nil {
			log.Println("Encountered an error when fetching alumni")
			log.Println(err)
			return
		}

		for _, a := range alumni {
			name := fmt.Sprintf("%v %v", a.FirstName, a.LastName)
			err := searchAndDisable(l, name)
			if err != nil {
				switch err {
				case ErrNoResults:
				default:
					log.Fatal(err)
				}
			}

		}
		opt.Page++
		if opt.NextPage == 0 {
			break
		}
	}
}

// Enable all current students' LDAP accounts.
// Ignores accounts with graduation year after the year set in the configuration.
// Logs any accounts not found in directory.
func enableCurrentStudents(client *vcapi.Client, l *ldap.Conn) {
	defer wg.Done()
	opt := &vcapi.ListOptions{
		Params: vcapi.Params{
			"option": "0",
		}}
	for {
		students, err := currentStudents(client, opt)
		if err != nil {
			log.Println("Encountered an error when fetching students")
			log.Println(err)
			return
		}

		for _, s := range students {
			// Even though 'option=0' should limit the results to current students,
			// The Veracross V2 API seems to ignore that option...
			// Checking the role just in case.
			if s.GraduationYear <= graduationYear && s.Role == "Student" {
				name := fmt.Sprintf("%v %v", s.FirstName, s.LastName)
				err := searchAndEnable(l, name)
				if err != nil {
					switch err {
					case ErrNoResults:
						log.Printf("Student %v, %v not found in LDAP Search", name, s.CurrentGrade)
					default:
						log.Fatal(err)
					}
				}
			}
		}
		opt.Page++
		if opt.NextPage == 0 {
			break
		}
	}
}
