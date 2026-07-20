# feedrrr

feedrrr is an RSS feed polling daemon that periodically checks configured RSS feeds on a cron schedule and sends new items to configurable notification sinks using [Shoutrrr](https://github.com/nicholas-fedor/shoutrrr).

## Features
- crontab-based syntax for jobs
- multiple notification sinks and aliases using [Shoutrrr](https://github.com/nicholas-fedor/shoutrrr). See its [documentation page](https://shoutrrr.nickfedor.com/latest/services/overview/) for supported services.
- plaintext or source (usually HTML) formatting in delivery

## Installation

### Docker
You may either use a docker image:

```bash
ghcr.io/das-kaesebrot/feedrrr
```

Example command to run it:

```
user@machine:~$ docker run --rm -it -v /path/to/config.yml:/etc/feedrrr/config.yml ghcr.io/das-kaesebrot/feedrrr
```

### Go install

Or go's install syntax:
```
go install dev.kaesebrot.eu/go/feedrrr/cmd/feedrrr@latest
```

## Build and install

### Build from source

Clone the repository and build the binary:

```
git clone https://github.com/das-kaesebrot/feedrrr.git
cd feedrrr
go build -o feedrrr ./cmd/feedrrr
./feedrrr -c config.yml
```

## Configuration

feedrrr is configured via a YAML configuration file. By default, it looks for `config.yml` in the following locations (in order of precedence):

1. `/etc/feedrrr/config.yml`
2. `$XDG_CONFIG_HOME/feedrrr/config.yml`
3. `$HOME/.config/feedrrr/config.yml`
4. `./config.yml` (current directory)

See [`config.example.yml`](config.example.yml) for a full example.

### Configuration

### Environment variables

| Variable | Description | Default |
|---|---|---|
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |

All configuration keys can be overridden via environment variables prefixed with `FEEDRRR_` (e.g. `FEEDRRR_JOBS`).

### CLI flags

| Flag | Short | Description |
|---|---|---|
| `--config` | `-c` | Explicitly override the configuration file used |
| `--loglevel` | `-l` | Log level |

#### YAML config
##### Jobs

Each job defines an RSS feed to poll and where to send notifications.

```yaml
jobs:
  my-feed:
    # RSS/Atom feed URL (required)
    source: https://example.com/feed.xml
    
    # Cron expression using either 5 or 6-elements (with seconds) syntax (required)
    schedule: "*/30 * * * *"
    
    # List of sink aliases or Shoutrrr URLs (required)
    sinks:
      - my-telegram
    
    # Convert article content to plain text (optional, default: false)
    plaintext: true
    
    # Title prefix (optional)
    prefix: "[New Post]"

    # Change detection mode (optional, default: guid)
    # can be either "guid" or "pubdate"
    # pubdate: articles published (specifically, the pubdate value) between last cronjob run and current cron job run will be detected as new
    # guid: articles that have appeared after an article with the guid seen during last cronjob run will be detected as new
    change_mode: guid
```

The `schedule` field accepts standard 5-field cron expressions (`minute hour dom month dow`) as well as 6-field expressions with seconds (`second minute hour dom month dow`). Timezone-aware schedules are supported via a `TZ=` or `CRON_TZ=` prefix (e.g. `TZ=Europe/Berlin 0 9 * * *`).

##### Sinks

Sinks map aliases to one or more Shoutrrr URLs [(reference)](https://shoutrrr.nickfedor.com/latest/services/overview/) and/or one or more aliases.

```yaml
sinks:
  my-telegram: telegram://bottoken?chats=chatid
  my-email: smtp://user:password@smtp.example.com:587/?from=alerts@example.com&to=me@example.com
  both-combined:
    - my-telegram
    - my-email
```
