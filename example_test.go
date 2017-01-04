package dlog_test

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"

	"github.com/ahmetalpbalkan/dlog"
)

func ExampleNewReader() {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
		Name: "/var/run/docker.sock",
		Net:  "unix"})
	if err != nil {
		log.Fatalf("cannot connect docker socket: %v", err)
	}

	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn, nil
		}}}

	url := "http://-/containers/CONTAINER_NAME/logs?stdout=1&stderr=1&follow=1"
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unexpected status code: %s", resp.Status)
	}

	// At this point we have a logs stream, here is how to read each log line
	// from container:
	r := dlog.NewReader(resp.Body)
	s := bufio.NewScanner(r)
	for s.Scan() {
		log.Println(s)
	}
	if err := s.Err(); err != nil {
		log.Fatalf("read error: %v", err)
	}
}
