# Queue Module for Tinh Tinh

<div align="center">
<img alt="GitHub Release" src="https://img.shields.io/github/v/release/tinh-tinh/queue">
<img alt="GitHub License" src="https://img.shields.io/github/license/tinh-tinh/queue">
<a href="https://codecov.io/gh/tinh-tinh/queue" > 
    <img src="https://codecov.io/gh/tinh-tinh/queue/graph/badge.svg?token=4MXQ4VGU2R"/> 
</a>
<a href="https://pkg.go.dev/github.com/tinh-tinh/queue"><img src="https://pkg.go.dev/badge/github.com/tinh-tinh/queue.svg" alt="Go Reference"></a>
</div>

<div align="center">
    <img src="https://avatars.githubusercontent.com/u/178628733?s=400&u=2a8230486a43595a03a6f9f204e54a0046ce0cc4&v=4" width="200" alt="Tinh Tinh Logo">
</div>

## Overview

The Queue module provides a robust, Redis-based job queue for the Tinh Tinh framework, supporting job scheduling, rate limiting, retries, concurrency, delayed jobs, priorities, and more.

## Install

```bash
go get -u github.com/tinh-tinh/queue/v2
```

## Features

- **Redis-Based:** Robust persistence and distributed processing.
- **Delayed Jobs:** Schedule jobs to run after a delay.
- **Cron Scheduling:** Schedule and repeat jobs using cron patterns.
- **Rate Limiting:** Control job processing rate.
- **Retries:** Automatic retry on failure.
- **Priority:** Job prioritization.
- **Concurrency:** Multiple workers per queue.
- **Pause/Resume:** Temporarily stop and resume job processing.
- **Crash Recovery:** Recovers jobs after process crashes.
- **Remove on Complete/Fail:** Clean up jobs after handling.

## Quick Start

### 1. Register the Module

```go
import "github.com/tinh-tinh/queue/v2"

queueModule := queue.ForRoot(&queue.Options{
    Connect: &redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    },
    Workers: 3,
    RetryFailures: 3,
})
```

Or via factory:

```go
queueModule := queue.ForRootFactory(func(ref core.RefProvider) *queue.Options {
    return &queue.Options{ /* ... */ }
})
```

### 2. Register and Inject Queues

```go
userQueueModule := queue.Register("user") // uses default/global options

// In your service or controller:
userQueue := queue.Inject(module, "user")
```

## Contributing

We welcome contributions! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or need help, you can:
- Open an issue in the GitHub repository
- Check our documentation
- Join our community discussions
