package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/sftp"
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

type Upload struct {
	src, dst string
}

func (u Upload) String() string { return fmt.Sprintf("%s:%s", u.src, u.dst) }
func (u *Upload) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) < 2 {
		return errors.New("malformed upload path, use: /local/path:/host/path")
	}
	u.src, u.dst = parts[0], parts[1]
	if len(u.src) == 0 || len(u.dst) == 0 {
		return fmt.Errorf("path cannot be empty, parsed local path: %q, parsed host path: %q", u.src, u.dst)
	}
	f, err := os.Open(u.src)
	defer f.Close()
	return err
}

func main() {
	var (
		cmd, user, privKey string
		passwordRequired   bool

		timeout = Timeout{10 * time.Second}

		readHostsFromFile string
		upload            Upload
		params            []string
	)

	flag.StringVar(&cmd, "c", "w", "command")
	flag.BoolVar(&passwordRequired, "p", false, "password is required (flag)")
	flag.StringVar(&user, "u", "", "username")
	flag.StringVar(&privKey, "i", "", "private key path")
	flag.Var(&timeout, "t", "timeout")
	flag.StringVar(&readHostsFromFile, "f", "", "read hosts from file")
	flag.Var(&upload, "upload", "upload file to host (format: /local/path:/host/path)")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		params = append(params, f.Name)
	})

	if slices.Contains(params, "c") && slices.Contains(params, "upload") {
		log.Fatal("command and upload are mutually exclusive parameters")
	}

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

	hosts := flag.Args()

	if readHostsFromFile != "" {
		f, err := os.Open(readHostsFromFile)
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			hosts = append(hosts, strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			log.Println(err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(hosts))

	for _, host := range hosts {
		go func() {
			defer wg.Done()

			var outBuf, errBuf, logBuf bytes.Buffer
			logger := log.New(&logBuf, "", log.LstdFlags|log.Lshortfile)

			t := time.Now()
			defer func() {
				fmt.Printf("%s ❭❭❭ %s (time: %s)\n", host, cmd, time.Since(t).Truncate(100*time.Millisecond))
				if outBuf.Len() > 0 {
					fmt.Println(outBuf.String())
				}
				if errBuf.Len() > 0 {
					fmt.Println(errBuf.String())
				}
				if logBuf.Len() > 0 {
					fmt.Println(logBuf.String())
				}
			}()

			if !strings.Contains(host, ":") {
				host += ":22"
			}
			if idx := strings.Index(host, "@"); idx >= 0 {
				config.User = host[:idx]
				host = host[idx+1:]
			}

			client, err := ssh.Dial("tcp", host, config)
			if err != nil {
				logger.Println(err)
				return
			}
			defer func() {
				if err := client.Close(); err != nil {
					logger.Println(err)
					return
				}
			}()

			if slices.Contains(params, "c") {
				// var cmdOutput strings.Builder // prevent mix of output of goroutines
				session, err := client.NewSession()
				if err != nil {
					logger.Println("Failed to create session: ", err)
					return
				}
				defer func() {
					if err := session.Close(); err != nil && err != io.EOF {
						logger.Println(err)
						return
					}
				}()

				// todo: session.Stdin = os.Stdin
				session.Stdout = &outBuf
				session.Stderr = &errBuf

				if err = session.Run(cmd); err != nil {
					logger.Println(err)
					return
				}
			}
			if slices.Contains(params, "upload") {
				sftpClient, err := sftp.NewClient(client)
				if err != nil {
					logger.Println(err)
					return
				}
				defer sftpClient.Close()

				dstFile, err := sftpClient.Create(upload.dst)
				if err != nil {
					logger.Println(err)
					return
				}

				srcFile, err := os.Open(upload.src)
				if err != nil {
					logger.Println(err)
					return
				}

				n, err := io.Copy(dstFile, srcFile)
				if err != nil {
					logger.Println(err)
					return
				}
				fmt.Fprintf(&outBuf, "%d bytes written to file %s", n, upload.dst)
			}
		}()
	}
	wg.Wait()
}
