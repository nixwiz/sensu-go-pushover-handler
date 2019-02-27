# sensu-go-pushover-handler
The Senso Go Pushover Handler is a [Sensu Event Handler][1] for sending incident
notifications to Pushover.

This handler reuses concepts found in the [sensu-email-handler][5].

## Installation
Create an executable script from this source or download one of the existing [releases][3].

From the local path of the sensu-go-pushover-handler repository:

```
go build -o /usr/local/bin/sensu-pushover-handler main.go
```

## Sensu Configuration

Example Sensu Go definition:

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "pushover"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-pushover-handler",
        "timeout": 10,
        "env_vars": [
            "PUSHOVER_TOKEN=a0b1c2d3e4f5g6h7i8j9k0l1m2n3o4",
            "PUSHOVER_USERKEY=a0b1c2d3e4f5g6h7i8j9k0l1m"
        ],
        "filters": [
            "is_incident",
            "not_silenced"
        ]
    }
}

```

## Pushover Configuration

This handler makes use of Pushover's [standard API][2] mechanism.  This means creating an application token, as well as
a user API key.

## Usage Examples

#### Help
```
The Sensu Pushover handler for sending notifications

Usage:
  sensu-pushover-handler [flags]

Flags:
  -C, --criticalPriority int8    The priority for Critical status messages (default 1)
  -h, --help                     help for sensu-pushover-handler
  -b, --messageBody string       The message body, in token substitution format (default "{{.Check.Output}}")
  -m, --messageTitle string      The message title, in token substitution format (default "{{.Entity.Name}}/{{.Check.Name}}")
  -O, --okPriority int8          The priority for OK status messages (default 0)
  -t, --pushoverToken string     The Pushover API token, if not in env PUSHOVER_TOKEN
  -u, --pushoverUserKey string   The Pushover User Key, if not in env PUSHOVER_USERKEY
  -U, --unknownPriority int8     The priority for Unknown status messages (default 1)
  -W, --warningPriority int8     The priority for Warning status messages (default 0)
```

#### Use of tokens

For defining the message title and body, [tokens][4] from the [events attributes][6] are used.

#### Environment Variables
|Variable|Setting|
|--------------------|-------|
|PUSHOVER_TOKEN| same as -t / --pushoverToken|
|PUSHOVER_USERKEY|same as -u / --pushoverUserKey|

**Note:**  The command line arguments take precedence over the environment variables above.

[1]: https://docs.sensu.io/sensu-go/5.2/reference/handlers/#how-do-sensu-handlers-work
[2]: https://pushover.net/api
[3]: https://github.com/nixwiz/sensu-go-pushover-handler/releases
[4]: https://docs.sensu.io/sensu-go/5.2/reference/tokens/#sensu-token-specification
[5]: https://github.com/sensu/sensu-email-handler
[6]: https://docs.sensu.io/sensu-go/5.2/reference/events/#attributes
