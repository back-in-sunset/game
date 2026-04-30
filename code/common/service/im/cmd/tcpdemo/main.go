package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"im/internal/goimx"
)

type scopeRef struct {
	TenantID    string `json:"tenant_id"`
	ProjectID   string `json:"project_id"`
	Environment string `json:"environment"`
}

func main() {
	var (
		addr        = flag.String("addr", "127.0.0.1:8091", "tcp server address")
		token       = flag.String("token", "", "jwt token")
		domain      = flag.String("domain", "platform", "im domain")
		tenantID    = flag.String("tenant", "", "tenant id")
		projectID   = flag.String("project", "", "project id")
		environment = flag.String("env", "", "environment")
		action      = flag.String("action", "", "optional command action")
		dataJSON    = flag.String("data", "", "optional command data json")
	)
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "token is required")
		os.Exit(2)
	}

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	adapter := goimx.NewAdapter()
	reader := bufio.NewReader(conn)

	loginBody, err := json.Marshal(map[string]any{
		"token":  *token,
		"domain": *domain,
		"scope": scopeRef{
			TenantID:    *tenantID,
			ProjectID:   *projectID,
			Environment: *environment,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	loginWire, err := adapter.Encode(goimx.OpAuth, 1, loginBody, &bytes.Buffer{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := conn.Write(loginWire); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	frame, err := adapter.Decode(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("login reply op=%d seq=%d body=%s\n", frame.Op, frame.Seq, string(frame.Body))

	if *action != "" {
		cmd := map[string]any{"action": *action}
		if *dataJSON != "" {
			var data any
			if err := json.Unmarshal([]byte(*dataJSON), &data); err != nil {
				fmt.Fprintln(os.Stderr, "invalid -data json:", err)
				os.Exit(1)
			}
			cmd["data"] = data
		}
		body, err := json.Marshal(cmd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		wire, err := adapter.Encode(goimx.OpServerPush, 2, body, &bytes.Buffer{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if _, err := conn.Write(wire); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		frame, err = adapter.Decode(reader)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("command reply op=%d seq=%d body=%s\n", frame.Op, frame.Seq, string(frame.Body))
	}

	heartbeatWire, err := adapter.Encode(goimx.OpHeartbeat, 3, nil, &bytes.Buffer{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := conn.Write(heartbeatWire); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	frame, err = adapter.Decode(reader)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("heartbeat reply op=%d seq=%d body=%s\n", frame.Op, frame.Seq, string(frame.Body))
}
