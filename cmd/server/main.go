package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if e, ok := err.(*ExitError); ok {
			os.Exit(e.Code)
		} else {
			os.Exit(1)
		}
	}
}

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	return e.Err.Error()
}

func run(ctx context.Context, args []string) error {
	fmt.Println("Server is running with args:", args)
	// Simulate an error for demonstration
	return &ExitError{Code: 2, Err: fmt.Errorf("simulated error")}
}
