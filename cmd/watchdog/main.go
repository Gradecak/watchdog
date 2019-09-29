package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gradecak/watchdog/pkg/api"
	"github.com/gradecak/watchdog/pkg/dispatcher"
	"github.com/gradecak/watchdog/pkg/dispatcher/queue"
	"github.com/gradecak/watchdog/pkg/events"
	"github.com/gradecak/watchdog/pkg/events/aggregator"
	"github.com/gradecak/watchdog/pkg/events/listeners/nats"
	"github.com/gradecak/watchdog/pkg/policy/gdpr"
	"github.com/gradecak/watchdog/pkg/policy/gdpr/provenance"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const (
	DISPATCH_QUEUE_SIZE  = 100000
	DEFAULT_NATS_URL     = "127.0.0.1"
	DEFAULT_NATS_CLUSTER = "test-cluster"
	DEFAULT_DB_URL       = "127.0.0.1:3306"
	DEFAULT_DB_USER      = "root"
	DEFAULT_DB_PASS      = "12345"
)

func runApp(ctx context.Context, c *cli.Context) error {
	logrus.Info("Setting up app...")
	prov, err := provenance.NewDBProv(&provenance.DbConf{
		Init: c.Bool("init-db"),
		User: c.String("db-user"),
		Pass: c.String("db-pass"),
		Db:   "watchdog",
		URL:  c.String("db-url"),
	})
	if err != nil {
		panic(err)
	}
	policy := gdpr.NewPolicy(prov)
	policyAPI := api.NewPolicyAPI(policy)
	dispatchAPI := setupDispatcher(dispatcher.NewDispatchQueue(DISPATCH_QUEUE_SIZE), policyAPI)

	if c.Bool("metrics") {
		go serveMetricsServer()
	}

	//main loop of the system
	runEventAggregator(dispatchAPI, &nats.Config{
		Cluster:  c.String("nats-cluster"),
		Client:   "watchdog",
		URL:      c.String("nats-url"),
		Prefixes: []string{"CONSENT", "PROVENANCE"},
	})

	return nil
}

func serveMetricsServer() {
	logrus.Info("Starting prometheus server...")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":5352", nil)
}

func setupDispatcher(q queue.DispatchQueue, p *api.Policy) *api.Dispatcher {
	dispatcher := dispatcher.NewDispatcher(q, p)
	go dispatcher.Run()
	return api.NewDispatcherAPI(q)
}

func runEventAggregator(d *api.Dispatcher, natsConf *nats.Config) {
	listener, err := nats.New(natsConf)
	if err != nil {
		panic(err.Error())
	}
	// start collecting events
	sys := aggregator.NewEventAggregator(d, []events.Listener{listener})
	sys.Run()
}

func main() {
	ctx, cancelFn := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Kill, syscall.SIGTERM)
	//shutdown logic
	go func() {
		for sig := range c {
			logrus.Infof("Recieved %v signal", sig)
			go func() {
				time.Sleep(30 * time.Second)
				logrus.Warn("Shutdown deadline exceeded, forcing shutdown")
				os.Exit(0)
			}()
			cancelFn()
			break
		}
	}()

	cliApp := newCli()
	cliApp.Action = func(c *cli.Context) error {
		return runApp(ctx, c)
	}
	cliApp.Run(os.Args)
}

func newCli() *cli.App {
	app := cli.NewApp()
	app.Flags = append([]cli.Flag{
		cli.BoolFlag{
			Name: "init-db",
		},
		cli.StringFlag{
			Name:  "nats-url",
			Value: DEFAULT_NATS_URL,
		},
		cli.StringFlag{
			Name:  "nats-cluster",
			Value: DEFAULT_NATS_CLUSTER,
		},
		cli.StringFlag{
			Name:  "db-url",
			Value: DEFAULT_DB_URL,
		},
		cli.StringFlag{
			Name:  "db-user",
			Value: DEFAULT_DB_USER,
		},
		cli.StringFlag{
			Name:  "db-pass",
			Value: DEFAULT_DB_PASS,
		},
		cli.BoolFlag{
			Name: "metrics",
		},
	})
	return app
}
