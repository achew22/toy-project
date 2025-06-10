package helloworld_test

import (
	"testing"

	"github.com/achew22/toy-project/internal/server/servertest"
)

func TestHelloWorldService_Golden(t *testing.T) {
	servertest.RunGoldenStepTests(t)
}