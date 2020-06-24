<!--
SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>

SPDX-License-Identifier: AGPL-3.0-or-later
-->

image-manager
=============

[![License](https://img.shields.io/badge/license-AGPLv3.0-blue.svg)](https://www.gnu.org/licenses/agpl-3.0.en.html)
[![GitHub release](https://img.shields.io/github/release/liri-infra/image-manager.svg)](https://github.com/liri-infra/image-manager)
[![GitHub issues](https://img.shields.io/github/issues/liri-infra/image-manager.svg)](https://github.com/liri-infra/image-manager/issues)
[![CI](https://github.com/liri-infra/image-manager/workflows/CI/badge.svg?branch=develop)](https://github.com/liri-infra/image-manager/actions?query=workflow%3ACI)

`image-manager` serves an image repository.

A build server sends ISO images and checksum files to image-manager, which archives them in
the correct spot based on their channel.

`image-manager` provides three subcommands:

  * **gentoken**: Generate an API token (more on that later).
  * **server**: An HTTP server that lets you upload files.
  * **client**: An HTTP client that uploads files.

## Dependencies

You need Go installed.

On Fedora:

```sh
sudo dnf install -y golang
```

Download all the Go dependencies:
  
```sh
go mod download
```

## Build

Build with:

```sh
make
```

## Install

Install with:

```sh
make install
```

The default prefix is `/usr/local` but you can specify another one:

```sh
make install PREFIX=/usr
```

And you can also relocate the binaries, this is particularly
useful when building packages:

```
...

%install
make install DESTDIR=%{buildroot} PREFIX=%{_prefix}

...
```

## Configuration file

Format:

```yaml
storage: <PATH TO STORAGE LOCATION>
channels:
  - name: <NAME>
    path: <PATH RELATIVE TO STORAGE LOCATION>
    cleanup: <BOOLEAN>
  - ...
tokens:
  - token: <TOKEN>
    created: <TIMESTAMP>
  - ...
```

## Token

All requests to the API require a token. You can generate one with:

```sh
image-manager gentoken [--config=<FILENAME>]
```

This command will generate a new token and store it in the YAML file `<FILENAME>`.
The file name is `image-manager.yaml` by default (that is when `--config` is not passed).

If you instead wants to use Docker type something like:

```sh
docker run --rm -it \
  -v $(pwd)/image-manager.yaml:/etc/image-manager.yaml \
  liriorg/image-manager \
  gentoken -c /etc/image-manager.yaml
```

## Server

Start the server with:

```sh
image-manager server [--config=<FILENAME>] [--path=<PATH>] [--verbose] --address=[<ADDR>]
```

This command will start the HTTP server.

The tokens are validated against the configuration file, see the previous
chapter for more information.

Replace `<PATH>` with the path to your archive.

Replace `<ADDR>` with the host name and port to bind, by default it's ":8080"
which means port `8080` on `localhost`.

Pass `--verbose` to print more messages.

If you instead wants to use Docker type something like:

```sh
docker run --rm -it \
  -v $(pwd)/image-manager.yaml:/etc/image-manager.yaml \
  -v $(pwd)/archive:/var/archive \
  -p 8080:8080 \
  liriorg/image-manager \
  receive -c /etc/image-manager.yaml -p /var/archive
```

## Client

Start the client with:

```sh
image-manager client [--token=<TOKEN>] [--address=<ADDR>] [--channel=<CHANNEL>] [[--file=<FILE>], ...] [--verbose]
```

This command will upload one or more files to the `<ADDR>` server, using the `<TOKEN>` API token.

Replace `<CHANNEL>` with the channel you want to upload to.

Replace `<FILE>` with the file to be uploaded.
You can pass `--file=<FILE>` multiple times.

Pass `--verbose` to print more messages.

If you instead wants to use Docker type something like:

```sh
docker run --rm -it \
  liriorg/image-manager \
  client --token=<TOKEN> \
    --channel=<CHANNEL> \
    --file=<ISO_FILE> \
    --file=<CHECKSUM_FILE>
```

## Licensing

Licensed under the terms of the GNU Affero General Public License version 3 or,
at your option, any later version.
