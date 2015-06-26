Watch Veracross API for current and former students and update their status in AD.

# What it does
* **Former Students/Alumni**: 
Check LDAP and disable account. If the LDAP account is not found, it skips over the record.
* **Current Students**: 
Check and enable account. If the LDAP account is not found, logs a message.
Example log message:
`2015/06/26 10:46:06 Student FirstName LastName, CurrentGrade not found in LDAP Search`

When enabling the account, the attribute `useraccountcontrol` is set to `512` and when disabling the account, the attribute is switched to `514`.
These attributes are taken from Active Directory. I'm not sure about other LDAP systems, although the [go-ldap](https://github.com/go-ldap/ldap) library is compatible with any LDAP v3 server.

# Creation/Deletion of accounts
This utility does not handle creation and deletion of AD accounts. Every school has different policies for usernames, OUs etc. A script/monitoring system can be used to create accounts or alerts based on the log output.

# Command line flags
```
Usage of ./vmanage:
  -debug=false: debug ldap connection
  -recent=true: Only check recent changes
  -server=false: run as daemon, with http server
```
By default, vmanage will check the /recent endpoint in the Veracross API.
Use `-recent=false` to check every Veracross contact instead.

By default, vmanage will check for changes and exit.
`-server` will run vmanage continuously checking Veracross for changes every 15 minutes. 
Running vmanage in server mode also turns on a http server, that listens on port 8080 and returns the status of the server.
Example: 
```
curl http://localhost:8080/status

{"status":"OK"}
```

The status endpoint can be used by a monitoring tool(like nagios) to check that the server is running.

# Configuration options
All configuration is sourced from environment variables. See `sample_env` file for an example.

* `BASE_DN=dc=example,dc=net`
* `LDAP_SERVER=ldap.example.net`
* `LDAP_ADMIN=cn=Administrator,cn=Users,dc=example,dc=net`
`LDAP_ADMIN` is an optional parameter. Defaults to the example above.
* `LDAP_PASSWORD=myPassword`
* `VCAPI_USERNAME=api.username`
* `VCAPI_PASSWORD=apiPassword`
* `VCAPI_SCHOOLID=whitby`
* `VCAPI_VERSION=v2`
`VCAPI_VERSION` is an optional parameter. Defaults to the example above.
* `GRADUATION_YEAR=2023`
`GRADUATION_YEAR` is used to configure the range of Veracross students to check. For example, we create accounts for 1st through 8th graders, so in 2015, we would use 2023 as the graduation year.

On unix systems(Mac, Linux) you can source the environment variables from a file.
Example: `source sample_env`

# TODO
* Also check Faculty and Staff
* More status responses. Currently status always returns OK.

