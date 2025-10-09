![Baton Logo](./baton-logo.png)

# `baton-metabase` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-metabase.svg)](https://pkg.go.dev/github.com/conductorone/baton-metabase) ![main ci](https://github.com/conductorone/baton-metabase/actions/workflows/main.yaml/badge.svg)

`baton-metabase` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Prerequisites
This connector requires the open source version Metabase v0.49.x or v0.56.x to be downloaded in order to function correctly.
* Official release: https://github.com/metabase/metabase/releases/tag/v0.49.0
  https://github.com/metabase/metabase/releases/tag/v0.56.0-beta

* Docker Hub image: https://hub.docker.com/r/metabase/metabase/tags?page=1&name=v0.49.0
  https://hub.docker.com/r/metabase/metabase/tags?name=v0.56.0

  Example (Docker):
  v0.49.0:
  ```
  docker run -d -p 3000:3000 \
  --name metabase \
  metabase/metabase:v0.49.0
  ```
  v0.56.0:
  ```
  docker run -d -p 3000:3000 \
  --name metabase \
  metabase/metabase:v0.56.0-beta
  ```
The previous commands starts Metabase and exposes it on port 3000 of the server.
The connector requires the --metabase-base-url parameter, which should be set to the URL where this Metabase instance is accessible (e.g., https://metabase.customer.com for production).
For example:
If Metabase is running on a server with domain metabase.customer.com and port 443 (HTTPS), the base URL would be:
* --metabase-base-url https://metabase.customer.com
  For the previous case of docker commands, the base URL would be:
* --metabase-base-url http://localhost:3000

## Connector credentials
1. What credentials or information are needed to set up the connector? (For example, API key, client ID and secret, domain, etc.)

   Requires a base URL and an API Key. Args: --metabase-base-url, --metabase-api-key

   The required URL was defined in the connector requirements instructions
   To obtain the API key follow the next steps:
    1. In your Metabase address where the open source version was launched, click on the gear icon in the upper right section and click on admin settings:
       v0.49:
       ![1-049.png](1-049.png)
       v0.56:   
       ![1-056.png](1-056.png)

    2. Click on authentication and API keys, then click on manage:
       v0.49:
       ![2-049.png](2-049.png)
       v0.56:
       ![2-056.png](2-056.png)

    3. Click on create API key:
       v0.49:
       ![3-049.png](3-049.png)
       v0.56:
       ![3-056.png](3-056.png)

    4. Fill in the required fields:
        * Key name: Enter any descriptive name (e.g. baton-connector).
        * Group: Select the administrators group as this will allow you to synchronize all connector resources with the API key.
        * Create the API key.
          v0.49:
          ![4-049.png](4-049.png)
          v0.56:
          ![4-056.png](4-056.png)

    5. Save your API key as you will not be able to view it again:
       v0.49:
       ![5-049.png](5-049.png)
       v0.56:
       ![5-056.png](5-056.png)

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-metabase
baton-metabase
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-metabase:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-metabase/cmd/baton-metabase@main

baton-metabase

baton resources
```

# Data Model

`baton-metabase` will pull down information about the following resources:
- Users

`baton-metabase` does not specify supporting account provisioning or entitlement provisioning.

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-metabase` Command Line Usage

```
baton-metabase

Usage:
  baton-metabase [flags]
  baton-metabase [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --metabase-base-url string     The base URL of the Metabase instance. e.g., https://metabase.customer.com ($METABASE_BASE_URL)
      --metabase-api-key string      API key generated in Metabase for the connector ($METABASE_API_KEY)
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-metabase
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-metabase

Use "baton-metabase [command] --help" for more information about a command.
```
