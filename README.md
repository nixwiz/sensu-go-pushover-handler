[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/nixwiz/sensu-go-pushover-handler)
![Go Test](https://github.com/nixwiz/sensu-go-pushover-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/nixwiz/sensu-go-pushover-handler/workflows/goreleaser/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/nixwiz/sensu-go-pushover-handler)](https://goreportcard.com/report/github.com/nixwiz/sensu-go-pushover-handler)

## Sensu Go Pushover Handler

- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Sensu Go](#sensu-go)
    - [Asset registration](#asset-registration)
    - [Handler definition](#handler-definition)
  - [Sensu Core](#sensu-core)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

### Overview

The Senso Go Pushover Handler is a [Sensu Event Handler][1] for sending incident notifications to [Pushover][5].

### Files

N/A

## Usage examples

### Help

```
The Sensu Go Pushover handler for sending notifications.

Usage:
  sensu-go-pushover-handler [flags]

Flags:
  -C, --criticalPriority uint    The priority for Critical status messages (default 1)
  -h, --help                     help for sensu-go-pushover-handler
  -b, --messageBody string       The message body, in token substitution format (default "{{.Check.Output}}")
  -m, --messageTitle string      The message title, in token substitution format (default "{{.Entity.Name}}/{{.Check.Name}}")
  -a, --pushoverAPIURL string    The Pushover API URL (default "https://api.pushover.net/1/messages")
  -s, --messageSound string      The sound for the message (default "pushover")
  -O, --okPriority uint          The priority for OK status messages (default 0)
  -t, --pushoverToken string     The Pushover API token
  -u, --pushoverUserKey string   The Pushover API user key
  -U, --unknownPriority uint     The priority for Unknown status messages (default 1)
  -W, --warningPriority uint     The priority for Warning status messages (default 0)
```

## Configuration
### Sensu Go
#### Asset registration

Assets are the best way to make use of this plugin. If you're not using an asset, please consider doing so! If you're using sensuctl 5.13 or later, you can use the following command to add the asset: 

`sensuctl asset add nixwiz/sensu-go-pushover-handler`

If you're using an earlier version of sensuctl, you can download the asset definition from [this project's Bonsai asset index page][7] or one of the existing [releases][3] or create an executable script from this source.

To build from source, from the local path of the sensu-go-pushover-handler repository:
```
go build -o /usr/local/bin/sensu-go-pushover-handler main.go
```

#### Handler definition

```yaml
api_version: core/v2
type: Handler
metadata:
  namespace: default
  name: pushover
spec:
  type: pipe
  command: sensu-go-pushover-handler -b http://sensu-backend.example.com:3000
  filters:
  - is_incident
  - not_silenced
  runtime_assets:
  - nixwiz/sensu-go-pushover-handler
  secrets:
  - name: SENSU_PUSHOVER_TOKEN
    secret: pushover#token
  - name: SENSU_PUSHOVER_USERKEY
    secret: pushover#userkey
  timeout: 10
```

**Security Note**: The Pushover Token and UserKey should always be treated as
security sensitive configuration options and in this example, they are loaded
into the handler configuration as environment variables using [secrets][10].
Command arguments are commonly readable from the process table by other
unpriviledged users on a system (ex: ps and top commands), so it's a better
practice to read in sensitive information via environment variables or
configuration files on disk. The --pushoverToken and --pushoverUserKey flags
are provided as an override for testing purposes.

### Sensu Core

N/A

## Installation from source

### Sensu Go

See the instructions above for [asset registration][9].

### Sensu Core

Install and setup plugins on [Sensu Core][8].

## Additional notes

### Pushover Configuration

This handler makes use of Pushover's [standard API][2] mechanism. This means creating an application token, as well as
a user API key.

### Use of tokens

For defining the message title and body, [tokens][4] from the [events attributes][6] are used.

### Environment Variables and Annotations

|Variable|Setting|Annotation|
|--------------------|-------|------|
|SENSU_PUSHOVER_TOKEN| same as -t / --pushoverToken|sensu.io/plugins/pushover/config/pushoverToken|
|SENSU_PUSHOVER_USERKEY|same as -u / --pushoverUserKey|sensu.io/plugins/pushover/config/pushoverUserKey|
|N/A|same as -b / --messageBody|sensu.io/plugins/pushover/config/messageBody|
|N/A|same as -m / --messageTitle|sensu.io/plugins/pushover/config/messageTitle|

**Note**: The command line arguments take precedence over the environment variables above.

**Note**: Annotations take precedence over command line arguments above.

## Contributing

N/A

[1]: https://docs.sensu.io/sensu-go/latest/reference/handlers/#how-do-sensu-handlers-work
[2]: https://pushover.net/api
[3]: https://github.com/nixwiz/sensu-go-pushover-handler/releases
[4]: https://docs.sensu.io/sensu-go/latest/reference/tokens/#sensu-token-specification
[5]: https://github.com/sensu/sensu-email-handler
[6]: https://docs.sensu.io/sensu-go/latest/reference/events/#attributes
[7]: https://bonsai.sensu.io/assets/nixwiz/sensu-go-pushover-handler
[8]: https://docs.sensu.io/sensu-core/latest/installation/installing-plugins/
[9]: #asset-registration
[10]: https://docs.sensu.io/sensu-go/latest/reference/secrets/
