# Integration tests

Tests that run the real server with a real Postgres (docker) and real
WebSocket clients. Planned for stages 3 and 4:

- three clients see the same timer within 0.5 s
- host reconnect restores the timer
- anonymous users can use everything; login never blocks
- backup and restore round trip

Use the `integration` build tag so `go test ./...` stays fast:

    go test -tags integration ./test/integration/...
