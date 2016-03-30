package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var sshAuthSock = os.Getenv("SSH_AUTH_SOCK")

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	localHome := flag.String("localhome", "", "local home directory of your git repositories")
	remoteHome := flag.String("remotehome", "", "remote home directory of your git repositories")
	user := flag.String("user", "vagrant", "user to log in as")
	flag.Parse()
	if len(flag.Args()) == 0 {
		log.Fatal("must provide host")
	}
	host := flag.Args()[0]
	if len(host) == 0 {
		log.Fatal("must provide host")
	}
	if len(*localHome) == 0 {
		log.Fatal("must provide local-home")
	}
	if len(*remoteHome) == 0 {
		log.Fatal("must provide remote-home")
	}

	agentUnixSock, err := net.Dial("unix", sshAuthSock)
	panicIfErr(err)
	defer agentUnixSock.Close()

	agent := agent.NewClient(agentUnixSock)

	signers, err := agent.Signers()
	panicIfErr(err)

	// An SSH client is represented with a ClientConn. Currently only
	// the "password" authentication method is supported.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of AuthMethod via the Auth field in ClientConfig.
	config := &ssh.ClientConfig{
		User: *user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}
	client, err := ssh.Dial("tcp", host+":22", config)
	panicIfErr(err)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var body struct {
			Cmd []string `json:"cmd"`
			Wd  string   `json:"wd"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			rw.WriteHeader(500)
			log.Printf("can't parse: %s", err.Error())
			rw.Write([]byte(err.Error()))
		}

		escaped := []string{}
		for _, el := range body.Cmd {
			escaped = append(escaped, "'"+strings.Replace(el, *localHome, *remoteHome, -1)+"'")
		}

		dir := strings.Replace(body.Wd, *localHome, *remoteHome, -1)
		cmd := "cd " + dir + " && git " + strings.Join(escaped, " ")

		stdout, stderr, err := executeCmd(client, string(cmd))

		out := map[string]interface{}{
			"stdout": stdout,
			"stderr": stderr,
		}
		if err == nil {
			out["code"] = 0
		} else if exitErr, ok := err.(*ssh.ExitError); ok {
			out["code"] = exitErr.ExitStatus()
		} else {
			rw.WriteHeader(500)
		}

		if out["code"] != 0 {
			log.Printf("input %v", body)
			log.Printf("output %v", out)
		}
		json.NewEncoder(rw).Encode(out)
	})
	log.Println("running in daemon mode on port 12345")
	log.Fatal(http.ListenAndServe(":12345", nil))
}

func executeCmd(client *ssh.Client, cmd string) (stdout, stderr string, err error) {
	session, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)
	return stdoutBuf.String(), stderrBuf.String(), err
}
