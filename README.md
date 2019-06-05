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

## Running

Edit `config.ini` before running this:

```sh
./image-server config.ini
```

## Licensing

Licensed under the terms of the GNU Affero General Public License version 3 or,
at your option, any later version.
