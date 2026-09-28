# golabel

Label printer app for the Epson TM-T20III with a simple web interface. Designed to run on [gokrazy](https://gokrazy.org/) — pure Go, no cgo.

Started as a port of `test2` in TestEscPos.

## Status

Working and in production on a gokrazy host (`hello`).

## Links

- **Source**: https://git.bytestone.uk/hum3/golabel
- **Issues**: https://git.bytestone.uk/hum3/golabel/issues

## Usage

`golabel` exposes an HTTP interface for printing labels. It is included in the `hello` gokrazy build (see `~/gokrazy/hello/config.json`).

## Development

```sh
task check    # fmt, vet, test
task test     # tests only
```

The module is consumed from the gokrazy build via a local `replace` directive pointing at `/home/hum3/minor/golabel`.
