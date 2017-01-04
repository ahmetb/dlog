# dlog

Go library to parse the binary Docker Logs stream into plain text.

`dlog` offers a single method: `NewReader(r io.Reader) io.Reader`. You are
supposed to give the response body of the `/containers/<id>/logs`. The returned
reader strips off the log headers and just gives the plain text to be used.

You can get the logs stream from [go-dockerclient][gocl]'s `Logs()`[gocl-logs] method,
or by calling the endpoint via the UNIX socket directly.

See [`example_test.go`][./example_test.go] for an example usage.

[gocl]: https://github.com/fsouza/go-dockerclient
[gocl-logs]: https://godoc.org/github.com/fsouza/go-dockerclient#Client.Logs

-----

Licensed under Apache 2.0. Copyright 2016 [Ahmet Alp Balkan][ab].

[ab]: https://ahmetalpbalkan.com/
