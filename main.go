package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type Timeout struct {
	time.Duration
}

func (t Timeout) String() string { return t.Duration.String() }
func (t *Timeout) Set(value string) (err error) {
	t.Duration, err = time.ParseDuration(value)
	return
}

func main() {
	var (
		cmd, user, privKey string
		passwordRequired   bool

		timeout = Timeout{10 * time.Second}
	)

	flag.StringVar(&cmd, "c", "w", "command")
	flag.BoolVar(&passwordRequired, "p", false, "password is required (flag)")
	flag.StringVar(&user, "u", "", "username")
	flag.StringVar(&privKey, "i", "", "private key path")
	flag.Var(&timeout, "t", "timeout")

	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	key, err := os.ReadFile(privKey)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout.Duration,
	}

	if signer, err := ssh.ParsePrivateKey(key); err != nil {
		log.Println(privKey, err)
	} else {
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if passwordRequired {
		fmt.Print("Enter SSH password: ")

		passphrase, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		// Create the Signer for this private key.
		if signer, err := ssh.ParsePrivateKeyWithPassphrase(key, passphrase); err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		} else {
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		}
	}

	fmt.Println()

	done := make(chan struct{})
	defer close(done)

	hosts := flag.Args()
	for _, host := range hosts {
		go func() {
			defer func() {
				done <- struct{}{}
			}()

			t := time.Now()
			client, err := ssh.Dial("tcp", host, config)
			if err != nil {
				log.Println("Failed to dial: ", err)
				return
			}
			defer func() {
				if err := client.Close(); err != nil {
					log.Println(err)
					return
				}
			}()

			// var cmdOutput strings.Builder // prevent mix of output of goroutines
			session, err := client.NewSession()
			if err != nil {
				log.Println("Failed to create session: ", err)
				return
			}
			defer func() {
				if err := session.Close(); err != nil && err != io.EOF {
					log.Println(err)
					return
				}
			}()

			out, err := session.CombinedOutput(cmd)
			if err != nil {
				log.Println("Failed to run:", err.Error())
				return
			}
			fmt.Printf("%s ❭❭❭ %s (time: %s)\n", host, cmd, time.Since(t).Truncate(100*time.Millisecond))
			fmt.Println(string(out))
		}()
	}

	for range len(hosts) {
		<-done
	}
}
