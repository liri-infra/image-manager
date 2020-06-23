<!--
SPDX-FileCopyrightText: 2019 Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>

SPDX-License-Identifier: AGPL-3.0-or-later
-->

image-manager
=============

[![License](https://img.shields.io/badge/license-AGPLv3.0-blue.svg)](https://www.gnu.org/licenses/agpl-3.0.en.html)
[![GitHub release](https://img.shields.io/github/release/liri-infra/image-manager.svg)](https://github.com/liri-infra/image-manager)
[![Build Status](https://travis-ci.org/liri-infra/image-manager.svg?branch=develop)](https://travis-ci.org/liri-infra/image-manager)
[![GitHub issues](https://img.shields.io/github/issues/liri-infra/image-manager.svg)](https://github.com/liri-infra/image-manager/issues)

image-manager serves an image repository.
CI jobs build images and send them to image-manager, which archives them in
the correct spot.

image-manager is written with Go.

## Build

Build the server:

```sh
make
```

Build and push the Docker container:

```sh
sudo make push
```

## Server

The server has only one (optional) argument: the configuration file name
which is `config.ini` by default.

Edit `config.ini` and then type this to run the server:

```sh
./image-server config.ini
```

### Configuration

See `config.ini` for an example.
There is also a users JSON database in `users.json`.

## Client

You can use the `image-server-client` Python program as a client.

It requires:

 * requests
 * requests-toolbelt

### Create a token

All APIs are protected by a JWT token, before calling any API you must create a token.

You can let the client read the password from a file:

```sh
./image-server-client create-token myusername password.txt
```

or from standard input:

```sh
echo mypassword | ./image-server-client create-token myusername
```

The client also lets you write the password and terminate with Ctrl+D,
just run it with:

```sh
./image-server-client create-token myusername
```

then type the password and press Ctrl+D.

The `token` is printed to standard output.

## Licensing

Licensed under the terms of the GNU Affero General Public License version 3 or,
at your option, any later version.
