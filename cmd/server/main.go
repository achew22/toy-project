package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	ctx := context.Background()
	run(ctx, os.Args)
}
func run(ctx context.Context, args []string) {
	fmt.Println("Server is running with args:", args)
}
