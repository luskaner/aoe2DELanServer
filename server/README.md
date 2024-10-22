# Server

[![Go Report Card](https://goreportcard.com/badge/github.com/luskaner/aoe2DELanServer/server)](https://goreportcard.com/report/github.com/luskaner/aoe2DELanServer/server)

The server module represents the core of the LAN Server. It is a simple web server that listens to the game's
API requests. The server reimplements the minimum required API surface to allow the game to work in LAN mode.

## Minimum system requirements

#### Stable

- Windows 10 (no S edition/mode).
- Windows Server 2016.
- Windows IoT.
- Linux: kernel 2.6.32 (see [here](https://go.dev/wiki/Linux) for more details).
- macOS: Catalina (v10.15).

Admin rights or firewall permission to listen on port 443 (https) will likely be required depending on the operating
system.

<details>
<summary>Experimental</summary>

- BSD-based (OpenBSD, DragonFly BSD, FreeBSD and NetBSD).
- Solaris-based (Solaris and Illumos).
- AIX.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.22.

</details>

## Configuration

### Certificate

You can use your own certificate by (re)placing the `cert.pem` and `key.pem` files in the `resources/certificates`
directory.
The easiest way to generate a self-signed certificate is by running the ``bin/genCert`` executable (more
info [here](../server-genCert) or you may leave
that to
the ```launcher``` if you are hosting and running the launcher on same PC.

### Main

The few configuration options are available in the [`config.toml`](resources/config/config.toml) file. The file is
self-explanatory and should be easy to understand.

### Login

The configuration file sent to the client upon login is [`login.json`](resources/config/age2/login.json). Some options
are
easy to understand while others might require researching.

### Cloud

The game connects to a static cloud to download assets. The server is configured to replace the original calls to
itself. The configuration file is [`cloudfilesIndex.json`](resources/config/age2/cloudfilesIndex.json) and the
corresponding
files reside in the [`cloud`](resources/responses/cloud) directory.

### Other static responses

The server also serves some static responses for the game to work. The files are located in
the [`responses`](resources/responses) base directory:

- [`Achievements`](resources/responses/age2/achievements.json): List of achievements.
- [`Leaderboards`](resources/responses/age2/leaderboards.json): List of leaderboards.
- [`Automatch maps`](resources/responses/age2/automatchMaps.json): List of maps for automatch.
- [`Challenges`](resources/responses/age2/challenges.json): List of challenges.
- [`Presence Data`](resources/responses/age2/presenceData.json): Presence data. Basically if a player is online, offline
  or
  away.
- [`Item Definitions`](resources/responses/age2/itemDefinitions.json): Definitions of items. Includes rewards,
  challenges and
  other items.
- [`Item Bundle Items`](resources/responses/age2/itemBundleItems.json): Grouping of items into bundles.

*Note: These files might require updates to work with future game versions.*

## Command Line

CLI is available with similar options as the configuration. You can see the available options with
`server -h`.

## API endpoints

For documentation on how what each endpoints does, please refer
to [LibreMatch documentation](https://wiki.librematch.org/rlink/game/start). Other endpoints are mostly
self-explanatory.

## Docker

See [Docker](../server-docker) for information.

## Exit Codes

* [Base codes](../common/errors.go).
* [Server codes](internal/errors.go).
