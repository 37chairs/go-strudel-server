# Go Strudel WebSocket Server

Strudel only runs in browser, and I want to be able to make noise by running commands and such. So, I wrote a basic go server to serve the html and websocket
that allow me to send some basic commands. For my purposes, I probably won't do much more of the Strudel DSL, but who knows.

Main challenge here is that Strudel is more of a DSL than a library, so we have to kindof wrap it and run evals. Kinda gross.

```bash
go run main.go
```

The server will start on `http://localhost:8069` and the WebSocket endpoint will be available at `ws://localhost:8069/ws` (nice).
Open your browser and navigate to `http://localhost:8069`. The web client will automatically connect to the WebSocket server and provide:

#### Ruby Client

Don't be a caveman, use rbenv and get a decent ruby.

```bash
bundle install
ruby ruby_client.rb
```

Commands:

- `setcps <cps>` - Set the tempo (e.g., `setcps 1.2`)
- `play <pattern>` - Send a play command (e.g., `play s(\"bd*4\")`)
- `stop` - Send a stop command
- `quit` - Exit the client

#### Examples

```
> setcps 1
> play stack(s("bd*2, ~ sd"), s("hh*4").gain(0.4))
> stop
```

# License

MIT / do whatever you want, just don't do it with rust.
