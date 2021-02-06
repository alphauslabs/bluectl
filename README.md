![main](https://github.com/alphauslabs/bluectl/workflows/main/badge.svg)

`bluectl` is the official command line interface for Alphaus services.

To install using [brew](https://brew.sh/), run the following command:

```bash
$ brew install mobingi/tap/bluectl
```

By default, this tool will look for the following environment variables for authentication:

```bash
ALPHAUS_CLIENT_ID
ALPHAUS_CLIENT_SECRET
```

You can also use the `--client-id` and `--client-secret` flags to set the values explicitly.

Run `bluectl -h` or `bluectl <subcommand> -h` to explore the available subcommands and flags.
