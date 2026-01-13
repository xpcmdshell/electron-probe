package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
)

//go:embed scripts/*
var scriptFolder embed.FS

func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}

func getTarget(ctx context.Context, devt *devtool.DevTools, targetType string, urlFilter string) (*devtool.Target, error) {
	switch targetType {
	case "node":
		return devt.Get(ctx, devtool.Node)
	case "page":
		if urlFilter != "" {
			targets, err := devt.List(ctx)
			if err != nil {
				return nil, err
			}
			for _, t := range targets {
				if t.Type == devtool.Page && strings.Contains(t.URL, urlFilter) {
					return t, nil
				}
			}
			return nil, fmt.Errorf("no page target matching URL filter: %s", urlFilter)
		}
		return devt.Get(ctx, devtool.Page)
	case "auto":
		if pt, err := devt.Get(ctx, devtool.Node); err == nil {
			return pt, nil
		}
		if urlFilter != "" {
			targets, err := devt.List(ctx)
			if err != nil {
				return nil, err
			}
			for _, t := range targets {
				if t.Type == devtool.Page && strings.Contains(t.URL, urlFilter) {
					return t, nil
				}
			}
		}
		return devt.Get(ctx, devtool.Page)
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}
}

func listTargets(ctx context.Context, devt *devtool.DevTools) {
	targets, err := devt.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list targets: %v", err)
	}
	fmt.Printf("Available targets:\n")
	for i, t := range targets {
		fmt.Printf("  [%d] %s (%s)\n      URL: %s\n", i, t.Title, t.Type, t.URL)
	}
}

func main() {
	scriptPath := flag.String("script", "", "Path to JS script to evaluate in the target")
	evalExpr := flag.String("eval", "", "JS expression to evaluate directly (alternative to -script)")
	inspectTarget := flag.String("inspect-target", "", "V8 inspector/CDP listener (e.g., http://localhost:9222)")
	targetType := flag.String("target-type", "auto", "Target type: node, page, or auto (tries node first, then page)")
	urlFilter := flag.String("url-filter", "", "Filter page targets by URL substring (e.g., 'claude.ai')")
	list := flag.Bool("list", false, "List available targets and exit")
	flag.Parse()

	if *inspectTarget == "" {
		log.Fatalf("Must specify -inspect-target")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	devt := devtool.New(*inspectTarget)

	if *list {
		listTargets(ctx, devt)
		os.Exit(0)
	}

	if *scriptPath == "" && *evalExpr == "" {
		log.Fatalf("Must specify -script or -eval")
	}

	var scriptData []byte
	var err error
	if *evalExpr != "" {
		scriptData = []byte(*evalExpr)
	} else {
		scriptData, err = scriptFolder.ReadFile(*scriptPath)
		if err != nil {
			scriptData, err = os.ReadFile(*scriptPath)
			if err != nil {
				log.Fatalf("Failed to read script: %v", err)
			}
		}
	}

	pt, err := getTarget(ctx, devt, *targetType, *urlFilter)
	if err != nil {
		log.Fatalf("Failed to get target: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Target: %s (%s)\n", pt.Title, pt.URL)

	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	c := cdp.NewClient(conn)

	eval := runtime.NewEvaluateArgs(string(scriptData))
	eval.AwaitPromise = BoolAddr(true)
	eval.ReplMode = BoolAddr(true)
	reply, err := c.Runtime.Evaluate(context.Background(), eval)
	if err != nil {
		log.Fatalf("Evaluation error: %v", err)
	}

	if reply.ExceptionDetails != nil {
		log.Fatalf("Exception(line %d, col %d): %v\n", 
			reply.ExceptionDetails.LineNumber, 
			reply.ExceptionDetails.ColumnNumber, 
			reply.ExceptionDetails.Exception)
	}

	s, _ := strconv.Unquote(string((*reply).Result.Value))
	fmt.Printf("%s\n", s)
}
