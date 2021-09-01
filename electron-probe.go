package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
)

//go:embed scripts/*
var scriptFolder embed.FS

// Working around a syntax limitation
func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}

func main() {
	scriptPath := flag.String("script", "", "Path to JS script to evaluate in the target")
	inspectTarget := flag.String("inspect-target", "", "V8 inspector listener")
	flag.Parse()
	if *inspectTarget == "" {
		log.Fatalf("Must specify inspector target")
	}
	if *scriptPath == "" {
		log.Fatalf("Must specify script payload")
	}

	scriptData, err := scriptFolder.ReadFile(*scriptPath)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	devt := devtool.New(*inspectTarget)
	pt, err := devt.Get(ctx, devtool.Node)
	if err != nil {
		panic(err)
	}
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := cdp.NewClient(conn)

	eval := runtime.NewEvaluateArgs(string(scriptData))
	eval.AwaitPromise = BoolAddr(true)
	eval.ReplMode = BoolAddr(true)
	reply, err := c.Runtime.Evaluate(context.Background(), eval)
	if err != nil {
		panic(err)
	}

	if reply.ExceptionDetails != nil {
		// Dump the exception details if the script run was unsuccessful
		log.Fatalf("Exception(line %d, col %d): %v\n", reply.ExceptionDetails.LineNumber, reply.ExceptionDetails.ColumnNumber, reply.ExceptionDetails.Exception)
	}

	// discarding the error result, failure doesn't matter.
	// This will just handle cases where string results come
	// back doubled escaped, causing parsing issues in follow-up
	// tools like `jq`
	s, _ := strconv.Unquote(string((*reply).Result.Value))

	fmt.Printf("%s\n", s)
}
