![main](https://github.com/alphauslabs/bluectl/workflows/main/badge.svg)

**(work-in-progress)**

`bluectl` is the official command line interface for Alphaus services.

To install using [brew](https://brew.sh/), run the following command:

```bash
$ brew tap alphauslabs/tap # only once
$ brew install bluectl
```

By default, this tool will look for the following environment variables for [authentication](https://alphauslabs.github.io/blueapi/authentication/apikey.html):

```bash
# For Ripple users:
ALPHAUS_RIPPLE_CLIENT_ID
ALPHAUS_RIPPLE_CLIENT_SECRET

# For Wave users:
ALPHAUS_WAVE_CLIENT_ID
ALPHAUS_WAVE_CLIENT_SECRET
```

You can also use the `--client-id` and `--client-secret` flags to set the values explicitly.

Run `bluectl -h` or `bluectl <subcommand> -h` to know more about the available subcommands and flags.
