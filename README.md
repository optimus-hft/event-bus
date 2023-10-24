# Event Bus
![pipeline](https://github.com/optimus-hft/event-bus/actions/workflows/go-ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/optimus-hft/event-bus/branch/main/graph/badge.svg)](#)
[![Go Report Card](https://goreportcard.com/badge/github.com/optimus-hft/event-bus)](https://goreportcard.com/report/github.com/optimus-hft/event-bus)
[![Go Reference](https://pkg.go.dev/badge/github.com/optimus-hft/event-bus.svg)](https://pkg.go.dev/github.com/optimus-hft/event-bus)

## GoLang Library for implementing event-driven architecture
EventBus can be used to implement event-driven architectures in Golang, Each bus can have multiple subscribers to different topics.

## Features
+ Listening on events using channels or callbacks with the ability to cancel subscriptions at any time.
+ Once subscriptions, After receiving the first event, Subscription is automatically cancelled.
+ Non-blocking publishing, Yet event ordering is guaranteed for each subscriber.

## Getting Started
### Installation
```
go get github.com/optimus-hft/event-bus
```

### Usage

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	eventbus "github.com/optimus-hft/event-bus"
)

func main() {
	bus := eventbus.New[int]()
	channel, unsub := bus.Subscribe("t1")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		for {
			select {
			case event := <-channel:
				fmt.Println("event", event)
			case <-ctx.Done():
				unsub()
				return
			}
		}
	}()

	bus.Publish("t1", 1)
	bus.Publish("t1", 2)
	bus.Publish("t1", 3)

	time.Sleep(time.Second)
}
```

## Contributing
Pull requests and bug reports are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
This project is licensed under the MIT License.