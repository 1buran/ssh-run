# ssh-run - Run ssh command on multiple hosts

This script runs given command on all given hosts concurrency (goroutines used).
It has very very basic ssh functionality and was created for cases when you need just to run
some command on park of servers to get quickly some info e.g. version of installed php
or free disk space etc.

Installation:
```
go install github.com/1buran/run-ssh@latest
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
