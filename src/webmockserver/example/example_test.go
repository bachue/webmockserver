package example_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	proto "webmockserver/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var mockClient proto.WebMockClient
var httpClient http.Client

func TestServer(t *testing.T) {
	assertions := &proto.WebMockAssertions{
		Assertions: []*proto.WebMockAssertion{
			&proto.WebMockAssertion{
				Host:        "localhost",
				MatchHost:   true,
				Method:      "GET",
				MatchMethod: true,
				Path:        "/call/method/getSomething",
				MatchPath:   true,
				AtLeast:     1,
				AtMost:      1,
				ReturnHeaders: map[string]*proto.StringArray{
					"Content-Type": &proto.StringArray{Strings: []string{"text/plain"}},
				},
				ReturnBody: []byte("ok"),
			},
			&proto.WebMockAssertion{
				Host:        "localhost",
				MatchHost:   true,
				Method:      "POST",
				MatchMethod: true,
				Path:        "/call/method/changeSomething",
				MatchPath:   true,
				Headers: map[string]*proto.StringArray{
					"Content-Type": &proto.StringArray{Strings: []string{"text/plain"}},
				},
				MatchHeaders: true,
				Parameters: map[string]*proto.StringArray{
					"query":  &proto.StringArray{Strings: []string{"success_query"}},
					"limit":  &proto.StringArray{Strings: []string{"20"}},
					"offset": &proto.StringArray{Strings: []string{"0"}},
				},
				MatchParameters:  true,
				AtLeast:          1,
				AtMost:           2,
				ReturnStatusCode: 201,
				ReturnHeaders: map[string]*proto.StringArray{
					"Content-Type": &proto.StringArray{Strings: []string{"text/plain"}},
				},
				ReturnBody: []byte("done"),
			},
			&proto.WebMockAssertion{
				Host:        "localhost",
				MatchHost:   true,
				Method:      "PATCH",
				MatchMethod: true,
				Path:        "/call/method/changeSomething",
				MatchPath:   true,
				Headers: map[string]*proto.StringArray{
					"Content-Type": &proto.StringArray{Strings: []string{"application/json"}},
				},
				Body:             []byte("{\"status\":\"ok\"}"),
				MatchBody:        true,
				AtLeast:          2,
				AtMost:           2,
				ReturnStatusCode: 400,
				ReturnHeaders: map[string]*proto.StringArray{
					"Content-Type": &proto.StringArray{Strings: []string{"text/plain"}},
				},
				ReturnBody: []byte("status should be one of success, error or fatal"),
			},
		},
	}
	if _, err := mockClient.Set(context.Background(), assertions); err != nil {
		t.Errorf("Failed to set assertions: %s\n", err)
	}
	defer mockClient.Done(context.Background(), new(proto.Void))

	if resp, err := httpClient.Get("http://localhost:12204/call/method/getSomething"); err != nil {
		t.Errorf("Failed to call GET method: %s\n", err)
	} else if resp.StatusCode != 200 {
		t.Errorf("GET method returns %d\n", resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("GET method reads body failed: %s\n", err)
	} else if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("GET method returns wrong headers: %v\n", resp.Header)
	} else if bytes.Compare(body, []byte("ok")) != 0 {
		t.Errorf("GET method returns wrong body: %s\n", body)
	}

	if resp, err := httpClient.Post("http://localhost:12204/call/method/changeSomething?limit=20&offset=0&query=success_query", "text/plain", nil); err != nil {
		t.Errorf("Failed to call POST method: %s\n", err)
	} else if resp.StatusCode != 201 {
		t.Errorf("POST method returns %d\n", resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("POST method reads body failed: %s\n", err)
	} else if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("POST method returns wrong headers: %v\n", resp.Header)
	} else if bytes.Compare(body, []byte("done")) != 0 {
		t.Errorf("POST method returns wrong body: %s\n", body)
	}

	if req, err := http.NewRequest(
		"PATCH",
		"http://localhost:12204/call/method/changeSomething",
		strings.NewReader("{\"status\":\"ok\"}")); err != nil {
		t.Errorf("Failed to construct PATH request\n")
	} else if resp, err := httpClient.Do(req); err != nil {
		t.Errorf("Failed to call PATCH method: %s\n", err)
	} else if resp.StatusCode != 400 {
		t.Errorf("PATCH method returns %d\n", resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Errorf("PATCH method reads body failed: %s\n", err)
	} else if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("PATCH method returns wrong headers: %v\n", resp.Header)
	} else if bytes.Compare(body, []byte("status should be one of success, error or fatal")) != 0 {
		t.Errorf("PATCH method returns wrong body: %s\n", body)
	}

	if resp, err := httpClient.Post("http://localhost:12204/call/method/changeSomething?limit=20&offset=20&query=success_query", "text/plain", nil); err != nil {
		t.Errorf("Failed to call POST method: %s\n", err)
	} else if resp.StatusCode != 404 {
		t.Errorf("PATCH method returns %d\n", resp.StatusCode)
	}

	if results, err := mockClient.Done(context.Background(), new(proto.Void)); err != nil {
		t.Errorf("Failed to export results: %s\n", err)
	} else if results.Success {
		t.Errorf("results should be failed\n")
	} else if len(results.AssertionResults) != 3 {
		t.Errorf("assertions should be 3\n")
	} else if !results.AssertionResults[0].Success {
		t.Errorf("the first assertion should be successful\n")
	} else if !results.AssertionResults[1].Success {
		t.Errorf("the second assertion should be successful\n")
	} else if results.AssertionResults[2].Success {
		t.Errorf("the last assertion should be wrong\n")
	}
}

func TestMain(m *testing.M) {
	if cmd, err := startServer(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		startHTTPClient()
		if err = startMockClient(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		exitCode := m.Run()
		if err = stopServer(cmd); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(exitCode)
	}
}

func startServer() (*exec.Cmd, error) {
	cmd := exec.Command("../../../server", "--grpc-host", "127.0.0.1", "--grpc-port", "12203", "--http-host", "127.0.0.1", "--http-port", "12204")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	time.Sleep(100 * time.Millisecond)
	return cmd, err
}

func stopServer(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}

func startMockClient() error {
	if conn, err := grpc.Dial("127.0.0.1:12203", grpc.WithInsecure(), grpc.WithTimeout(1*time.Second)); err != nil {
		return err
	} else {
		mockClient = proto.NewWebMockClient(conn)
		return nil
	}
}

func startHTTPClient() {
	httpClient = http.Client{
		Transport: http.DefaultTransport,
		Timeout:   1 * time.Second,
	}
}
