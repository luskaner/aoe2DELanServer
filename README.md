# Age of Empires 2 Definitive Edition LAN Server

AoE2:DE LAN Server is a web server that allows you to play multiplayer **LAN** game modes without having an internet
connection **to the game server** (*aoe-api.worldsedgelink.com*) paving the way to how the original AoE2 worked plus many features new to HD and DE
versions.

**You will still need a way to bypass the *online-only* restriction that is imposed by the game to being connected to
Steam or Xbox Live depending on the version to fully play offline.**

## Features

- Co-Op Campaigns.
- Scenarios (including transferring the map):
    - Event Scenarios*.
    - Custom Scenarios.
- ... all other game modes available by creating a lobby **only with server as "Use Local Lan Server"**.
- Rematch.
- Restore game.
- Data mods.
- Invite player to lobby (including via link but only if game is already running).
- Player Search.
- Chatting (both in the lobby and in-game).
- Crossplay (cross-platform) Steam & Xbox (PC-only).

*\*Will change depending on the server version and might require an update.*

## Unsupported features

- Not compatible with Battle Server P2P (LAN):
    - Quick Play.
    - Ranked.
    - Spectate Games.
- Not possible as it would require internet and some access to the user profile:
    - Steam & Xbox Friends.
- Not implemented:
    - Changing player profile icon: the default will always be used.
    - Leaderboards: will appear empty.
    - Player stats: will appear empty.
    - Clans: all players are without clans. Browsing clan will appear empty and creating one will always result in
      error.
    - Lobby ban player: will appear like it works but doesn't.
    - Report player: will appear like it works but doesn't.

## System Requirements

### Server

- Windows: 10 or higher, Server 2016 or higher.
- MacOS: Catalina 10.15 or higher.
- GNU/Linux: *any supported distro, see the note below for details*.

Admin rights or firewall permission to listen to port 443 for https will likely be required (once or repeatedly)
depending on the operating system.

Note: For the full list see [minimum requirements for Go](https://go.dev/wiki/MinimumRequirements) 1.22.

### Launcher

- Windows: 10 or higher, (possibly Server 2016 or higher) all x86-64 (same as the game).
- If you allow it to handle the hosts file, local certificate or a custom game launcher which requires admin privileges, it will require rights elevation.

### Client

- Age of Empires 2 Definitive Edition - Steam or Microsoft Store.
- Up-to-date version of the game.

## Binaries

See the [releases page](https://github.com/luskaner/aoe2DELanServer/releases) for server and launcher binaries for
supported operating systems.

*Note: If you are using Windows Defender on Windows it may flag one or more executables as virus, this is a **false positive***.

### Verification

The verification process ensures that the files you download are the same as the ones that were uploaded by the
maintainer.

1. Check the release tag is verified with the committer's signature key (*as all commits must be*).
2. Download the ```..._checksums.txt``` and ```..._checksums.txt.sig``` files.
3. Import the [release public key](release_public.key) and import it to your keyring if you haven't already.
4. Verify the ```..._checksums.txt``` file with the ```..._checksums.txt.sig``` file.
5. Verify the SHA-256 checksum list inside ```..._checksums.txt``` with the downloaded archives.

## Installation

Both the launcher and server work out of the box without any installation. Just download the compressed archives,
decompress and run them.

## How it works

### Server

The server is simple web server that listens to the game's API requests. The server reimplements
the minimum required API surface to allow the game to work in LAN mode. It is completely safe as no data sent from the
client is stored or sent to any other server.

*Note: See the [server README](server/README.md) for more details.*

### Launcher

The launcher allows to easily play the game in LAN mode while allowing the official launcher to be used for online play.

It can do the following setup steps for you:

- Automatically start/stop the server or connect to an existing one automatically.
- (Optional) Use an isolated metadata and profile directories to avoid potential issues with the official game.
- (Optional) Modify the hosts file to redirect the game's API requests to the LAN server.
- (Optional) Install a self-signed certificate to allow the game to connect to the LAN server.
- Automatically find and start the game.

Afterwards, it reverses any changes to allow the official launcher to connect to the official servers.

*Note: See the [launcher README](launcher/README.md) for more details.*

## Simplest way to use it

1. **Download the asset `aoe2DELanServer_X_win_x86-64.zip`** from the latest
   release https://github.com/luskaner/aoe2DELanServer/releases
2. **Uncompress** it somewhere (it's fully portable and with no dependencies).
3. If not using the Steam or Microsoft Store launcher, **edit the [launcher/config.ini](launcher/config/config.ini) file**
   and modify
   the `Client.Executable` section to point to the game launcher path, e.g `C:\AoE2DE\launcher.exe` (no quotes needed).
   You will need to use a custom launcher for 100% offline play.
4. **Execute `launcher/launcher.exe`**: you will be asked for admin elevation and confirmation of other dialogs as
   needed, you
   will also need to allow the connections via the Microsoft Defender Firewall or any other.
5. **Repeat the above steps for every PC** you want to play in LAN with by running the `launcher.exe`, the first PC to
   launch
   it will host the "server" and the rest will auto-discover and connect to it.
6. In the game, when hosting a new lobby, just make sure to set the server to **Use Local Lan Server**. Setting it to
   public
   visibility is recommended.
7. **Invite friends** by searching them by name and sending an invite as needed. They can also search in the lobby
   browser
   by
   the lobby ID or just connect to it directly if public (sharing the link to join the lobby automatically only works if the game is already running).

## Local development

### System requirements

- [Go 1.22](https://go.dev/dl/).
- [Git](https://git-scm.com/downloads).
- [Task](https://taskfile.dev/installation/).
- [GoReleaser](https://goreleaser.com/).

### Debug

It is recommended to use an IDE such as [GoLand](https://www.jetbrains.com/go/) (free for academia)
or [Visual Studio Code](https://code.visualstudio.com/) (free) with
the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go).

Depending on the module you want to debug, you will need to run the corresponding task **before**:

- server: ```task debug-prepare-server```
    - genCert: ```task debug:prepare-server-genCert```
- launcher: ```task debug:prepare-launcher```

### Build

Run ```task build```.

### Release

1. Install [gpg2](https://docs.releng.linuxfoundation.org/en/latest/gpg.html) if needed.
2. Create a new sign-only GPG key pair (*RSA 4096-bit*) with a passphrase.
3. Copy .env.example to .env and set ```GPG_FINGERPRINT``` to the fingerprint of the key.
4. Finally run ```task release```

## Terms of Use

You and all the clients connecting to your server are only authorized to use this software if:

- Owning a **legal license** of Age of Empires 2 Definitive Edition.
- Not using this software to cheat/hack and, in general, respect all the game terms of service.
- Use this software for personal use.
- Use this software in a LAN environment.

Disclaimer: This software is not affiliated with Xbox Game Studios, Microsoft Corporation, Forgotten Empires LLC,
World's Edge LLC, or any other entity that is involved in the development of Age of Empires 2 Definitive Edition.
