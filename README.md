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
  -f string
    	read hosts from file
  -i string
    	private key path
  -p	password is required (flag)
  -t value
    	timeout (default 10s)
  -u string
    	username
  -upload value
    	upload file to host (format: /local/path:/host/path)
```

Use `-u` to set username.

Use `-p` to enable prompt of entering password of ssh private key.

Use `-i` to specify a path of ssh private key.

Use `-c` to specify a commnad which will be executed over ssh.

Use `-f` to specify a path to file contains list of hosts (line by line).

Use `-t` option to set timout of ssh connection,
the format as golang [time.ParseDuration](https://pkg.go.dev/time#ParseDuration) expected e.g.:
- `20s`: 20 seconds
- `300s` or `5m`: 5 minutes
- `1h35m`: 1 hours 35 minutes

Use `-upload` to upload a file on hosts, format: `/local/path:/host/path`.

## Examples of usage

Basic params `-u`, `-c`, `-i`:

```
$ ssh-run [params] -i ~/.ssh/id_rsa -u admin -c "uptime" host1 host2:4566 root@host3 john@host4:3434 ...
```

this will execute `uptime` command on hosts:
- `host1`: connect to default ssh port - `22`, use username - `admin`
- `host2`: connect to custom ssh port - `4566`, use username - `admin`
- `host3`: connect to default ssh port - `22`, use username - `root`
- `host4`: connect to custom ssh port - `3434`, use username - `john`

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

Use `-p` to enter passphrase of ssh private key:

```
$ ssh-run -i ~/.ssh/aws-ec.pem -c "uname -r" -u admin -p 10.10.10.14 10.10.10.12
Enter SSH password:

10.10.10.12:22 ❭❭❭ uptime (time: 1.1s)
 18:32:12 up 84 days,  2:17,  0 user,  load average: 0.00, 0.00, 0.00

10.10.10.14:22 ❭❭❭ uptime (time: 2.1s)
 18:32:12 up 31 days,  2:17,  0 user,  load average: 0.00, 0.00, 0.00

```
