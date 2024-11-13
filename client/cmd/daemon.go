package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/robfig/cron"
	flag "github.com/spf13/pflag"
)

type daemon struct {
	daemon bool

	cron string
}

func (d *daemon) PersistentFlags(flagSet *flag.FlagSet) {
	flagSet.BoolVarP(&d.daemon, "daemon", "d", false, "Run the daemon, default is false")
	flagSet.StringVarP(&d.cron, "cron", "c", "0 0/5 * * *", "The cron expression to run the daemon, default is every 5 minutes")
}

func (d *daemon) isDaemon() bool {
	return os.Getppid() == 1
}

func (d *daemon) isTrue() bool {
	return d.daemon
}

func (d *daemon) isNeedDaemonize() bool {
	return d.isTrue()
}

// daemonize daemonize the process
func (d *daemon) daemonize(ctx context.Context) error {
	name := os.Args[0]
	arg := os.Args[1:]

	if !d.isDaemon() {
		d.run(name, arg...)
		slog.Info("Daemonize the process")
		// Exit the parent process, make the child process as a daemon
		os.Exit(0)
		return nil
	}

	d.cronRun(ctx, name, d.trimDaemonArgs(arg)...)
	return nil
}

func (d *daemon) cronRun(ctx context.Context, name string, arg ...string) {
	c := cron.New()
	c.AddFunc(d.cron, func() {
		d.run(name, arg...)
	})

	c.Start()
	defer c.Stop()

	// Run the command immediately
	d.run(name, arg...)

	slog.Info("Start cron job")
	<-ctx.Done()
	slog.Info("Stop cron job")
}

func (d *daemon) run(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err := cmd.Start()
	if err != nil {
		slog.Error("Failed to start the daemon",
			"error", err,
		)
	}
}

// trimDaemonArgs trim the daemon args, remove the daemon flag and its value
// such as --daemon, --daemon true, --daemon=true, -d, -d=true, -d true
func (d *daemon) trimDaemonArgs(args []string) []string {
	var result []string
	skipNext := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if skipNext {
			skipNext = false
			continue
		}

		if arg == "--daemon" || arg == "-d" {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				skipNext = true
			}
		} else if strings.HasPrefix(arg, "--daemon=") || strings.HasPrefix(arg, "-d=") {
			continue
		} else {
			result = append(result, arg)
		}
	}

	return result
}
