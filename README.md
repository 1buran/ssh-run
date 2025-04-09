# ssh-run - Run ssh command on multiple hosts

This script runs given command on all given hosts concurrently (goroutines used).
It has very very basic ssh functionality and was created for cases when you need just to run
some command on park of servers to get quickly some info e.g. version of installed php
or free disk space etc.

Installation:
```
go install github.com/1buran/ssh-run@latest
```

To print help run `ssh-run -h`:

```
Usage of ssh-run:
  -c string
    	command (default "w")
  -i string
    	private key path
  -p	password is required (flag)
  -t value
    	timeout (default 10s)
  -u string
    	username
```

Use `-u` to set username.

Use `-p` to enable prompt of entering password of ssh private key.

Use `-i` to specify a path of ssh private key.

Use `-c` to specify a commnad which will be executed over ssh.

Use `-t` option to set timout of ssh connection,
the format as golang [time.ParseDuration](https://pkg.go.dev/time#ParseDuration) expected e.g.:
- `20s`: 20 seconds
- `300s` or `5m`: 5 minutes
- `1h35m`: 1 hours 35 minutes

## Examples of usage

Specify username as part of host string:

```
$ ssh-run -i ~/.ssh/aws-ec.pem -c "uname -r" admin@10.10.10.14

10.10.10.14:22 ❭❭❭ uname -r (time: 1s)
6.1.0-30-cloud-amd64

```
the same as:

```
$ ssh-run -i ~/.ssh/aws-ec.pem -c "uname -r" -u admin 10.10.10.14

10.10.10.14:22 ❭❭❭ uname -r (time: 1s)
6.1.0-30-cloud-amd64

```

Multiple hosts:

```
$ ssh-run -i ~/.ssh/aws-ec.pem -c "uname -r" -u admin 10.10.10.14 10.10.10.12 10.10.10.24

10.10.10.12:22 ❭❭❭ uptime (time: 1.1s)
 18:32:12 up 84 days,  2:17,  0 user,  load average: 0.00, 0.00, 0.00

10.10.10.24:22 ❭❭❭ uptime (time: 2.0s)
 18:32:12 up 4 days,  2:17,  0 user,  load average: 0.00, 0.00, 0.00

10.10.10.14:22 ❭❭❭ uptime (time: 2.1s)
 18:32:12 up 31 days,  2:17,  0 user,  load average: 0.00, 0.00, 0.00

```
