# Monokit2

Monokit2 is a monitoring and alerting system that supports various plugins for monitoring different services and systems.

It is designed to be extensible, allowing users to add new plugins as needed.

[Plugin development for Monokit2](https://github.com/monobilisim/monokit2/blob/main/PLUGIN-DEVELOPMENT.md)

## Installation
The project is in beta state and installation instructions will be provided soon.

Currently, you can clone config files from example config folder to `/etc/mono`.

Download the latest monokit2 development version from [release](https://github.com/monobilisim/monokit2/releases/tag/devel)

Copy monokit2 binary to `/usr/local/bin/monokit2`

Make sure monokit2 binary is set to executable and run as root user.

Because this is a beta version, you may encounter bugs and issues.

We suggest using `monokit2 reset` to start from empty db and config files for each time you update monokit2 development binary.

### Alerts supported plugins and modules:
- osHealth:
  - disk: Zulip, Redmine
  - system load: Zulip, Redmine
  - memory: Zulip
  - zfs: Zulip, Redmine
  - systemd: Zulip
