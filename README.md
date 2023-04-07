# Go Condition Checker [![Build Status](https://github.com/xgfone/go-checker/actions/workflows/go.yml/badge.svg)](https://github.com/xgfone/go-checker/actions/workflows/go.yml) [![GoDoc](https://pkg.go.dev/badge/github.com/xgfone/go-checker)](https://pkg.go.dev/github.com/xgfone/go-checker) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-checker/master/LICENSE)


Provide a checker to check whether a condition is ok periodically by the checker config strategy.


## Install
```shell
$ go get -u github.com/xgfone/go-checker
```


## Example
```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xgfone/go-checker"
)

func main() {
	vipcond := checker.NewVipCondition("192.168.1.2", "")
	vipchecker := checker.NewChecker("vipchecker", vipcond, func(id string, ok bool) {
		if ok {
			fmt.Println("192.168.1.2 is bound")
		} else {
			fmt.Println("192.168.1.2 is not bound")
		}
	})

	// The checker uses checker.DefaultConfig by default,
	// but we can reset it to a new one.
	vipchecker.SetConfig(checker.Config{})

	// For Test
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel() // We can use the cancel function or call vipchecker.Stop to stop the checker.

	// Start the checker to check the condition periodically.
	go vipchecker.Start(ctx)

	// We can call the Ok method to get the ok status of the checker.
	// Also, we can use the callback function above to monitor it in real-time.
	fmt.Println(vipchecker.Ok())

	// For test, wait to end.
	<-ctx.Done()
}
```
