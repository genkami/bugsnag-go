package bugsnag_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	bugsnag "github.com/bugsnag/bugsnag-go"
)

var testAPIKey = "166f5ad3590596f9aa8d601ea89af845"
var testEndpoint string

func ExampleAutoNotify() {
	config := bugsnag.Configuration{APIKey: testAPIKey}
	createAccount := func(ctx context.Context) {
		fmt.Println("Creating account and passing context around...")
	}
	ctx := bugsnag.StartSession(context.Background())
	defer bugsnag.AutoNotify(ctx, config)
	createAccount(ctx)
	// Output:
	// Creating account and passing context around...
}

func ExampleRecover() {
	panicFunc := func() {
		fmt.Println("About to panic")
		panic("Oh noes")
	}

	// Will recover when panicFunc panics
	func() {
		config := bugsnag.Configuration{APIKey: testAPIKey}
		defer bugsnag.Recover(bugsnag.StartSession(context.Background()), config)
		panicFunc()
	}()

	fmt.Println("Panic recovered")
	// Output:
	// About to panic
	// Panic recovered
}

func ExampleConfigure() {
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       "YOUR_API_KEY_HERE",
		ReleaseStage: "production",
		// See bugsnag.Configuration for other fields
	})
}

func ExampleHandler() {
	handleReq := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling HTTP request")
	}

	// Set up your http handlers as usual
	http.HandleFunc("/", handleReq)

	// use bugsnag.Handler(nil) to wrap the default http handlers
	// so that Bugsnag is automatically notified about panics.
	http.ListenAndServe(":1234", bugsnag.Handler(nil))
}

func ExampleHandler_customServer() {
	handleReq := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling GET")
	}

	// If you're using a custom server, set the handlers explicitly.
	http.HandleFunc("/", handleReq)

	srv := http.Server{
		Addr:        ":1234",
		ReadTimeout: 10 * time.Second,
		// use bugsnag.Handler(nil) to wrap the default http handlers
		// so that Bugsnag is automatically notified about panics.
		Handler: bugsnag.Handler(nil),
	}
	srv.ListenAndServe()
}

func ExampleHandler_customHandlers() {
	handleReq := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling GET")
	}

	// If you're using custom handlers, wrap the handlers explicitly.
	handler := http.NewServeMux()
	http.HandleFunc("/", handleReq)
	// use bugsnag.Handler(handler) to wrap the handlers so that Bugsnag is
	// automatically notified about panics
	http.ListenAndServe(":1234", bugsnag.Handler(handler))
}

func ExampleNotify() {
	ctx := context.Background()
	ctx = bugsnag.StartSession(ctx)
	_, err := net.Listen("tcp", ":80")

	if err != nil {
		bugsnag.Notify(ctx, err)
	}
}

func ExampleNotify_details() {
	ctx := context.Background()
	ctx = bugsnag.StartSession(ctx)
	_, err := net.Listen("tcp", ":80")

	if err != nil {
		bugsnag.Notify(ctx, err,
			// show as low-severity
			bugsnag.SeverityInfo,
			// set the context
			bugsnag.Context{String: "createlistener"},
			// pass the user id in to count users affected.
			bugsnag.User{Id: "123456789"},
			// custom meta-data tab
			bugsnag.MetaData{
				"Listen": {
					"Protocol": "tcp",
					"Port":     "80",
				},
			},
		)
	}
}

func ExampleOnBeforeNotify() {

	type Job struct {
		Retry     bool
		UserID    string
		UserEmail string
	}

	bugsnag.OnBeforeNotify(func(event *bugsnag.Event, config *bugsnag.Configuration) error {
		// Search all the RawData for any *Job pointers that we're passed in
		// to bugsnag.Notify() and friends.
		for _, datum := range event.RawData {
			if job, ok := datum.(*Job); ok {
				// don't notify bugsnag about errors in retries
				if job.Retry {
					return fmt.Errorf("bugsnag middleware: not notifying about job retry")
				}
				// add the job as a tab on Bugsnag.com
				event.MetaData.AddStruct("Job", job)
				// set the user correctly
				event.User = &bugsnag.User{Id: job.UserID, Email: job.UserEmail}
			}
		}

		// continue notifying as normal
		return nil
	})
}
