# Feature Parity with <https://github.com/monobilisim/monokit>

Base monokit2 + monokit_lib roadmap. Each plugin tracks its own features in
`plugins/<name>/ROADMAP.md` (index at the bottom).

## core

- [X] __Config__ (`/etc/mono/*.yml`, global.yml + per-plugin files wired in lib)
  - [ ] Config structs + load cases for the new plugin configs (win, opnsense, redis, rabbitmq, es, vault, k8s, mail, traefik, pritunl, wppconnect, upcheck, glb, ssh-notifier, valkey; mongodb section in db.yml)
  - [ ] Environment variable expansion (`${VAR}`) in config values
  - [ ] Config feature-gating by file presence (auto-detection)
- [X] __Logging__ (zerolog to configured file)
  - [ ] Log level from environment (monokit1 `MONOKIT_LOGLEVEL`)
  - [ ] Log rotation (size/backups/retention/compression)
  - [ ] Local log viewer command (filters: time range, component, level, field matchers, wildcard matching)
- [-] __Alarms__
  - [X] Zulip webhook alarms with interval + limit throttling (SQLite-backed)
  - [X] Up/down state tracking per plugin/module
  - [ ] Zulip bot API sender (config exists, not implemented)
  - [ ] Custom stream/topic per alarm (webhook URL rewriting)
  - [ ] Microsoft Teams / Power Automate webhooks (Adaptive Card payload)
  - [ ] SMTP email alarms (STARTTLS, emoji shortcode replacement)
  - [ ] Recovery window (delay recovery alarms, cancel on relapse)
  - [ ] 24h re-alarm for persistent down states
- [-] __Redmine__
  - [X] Issue create/update/close with 6h dedupe window, interval + limit throttling
  - [X] News creation
  - [ ] Percentage-delta tracking (note only on increase, e.g. disk usage)
  - [ ] Issue link appended to alarm messages
  - [ ] Uninstall/reset trigger on Redmine project closure (monokit1 `uninstall`)
- [X] __SQLite state DB__ (GORM, auto-migrated schema in lib)
- [X] __Plugin runner__ (binary-per-plugin, exit-code propagation, `-d` dependency probe, version probe)
- [X] __TUI__ (plugin list/install/update)
- [-] __Updater__
  - [X] Self-update + plugin update from GitHub releases (checksums, backup/rollback, devel prerelease)
  - [X] Major-version guard (non-major only)
  - [ ] Force flag / explicit version selection
- [ ] __Daemon / scheduler__ (run plugins on interval; monokit1 daemon with component auto-detection, `--once`, OS service install incl. Windows service)
- [X] __shutdownNotifier__ (poweron/poweroff notifications, systemd unit)
- [X] __reset__ command
- [ ] __Client/Server API__ (monokit1 `common/api`: host inventory server, host registration + hostkey auth, config sync, enable/disable components, wants-update-to, health data posting, log shipping, domains/multi-tenancy, Keycloak SSO, AWX, Cloudflare, Valkey cache, Swagger)
  - [ ] Decide scope for monokit2 (server may stay a separate project)
- [ ] __Windows support__ for base (config/log/db paths, service)

## plugins

| Plugin | Status | Roadmap |
|---|---|---|
| osHealth | active | [plugins/osHealth/ROADMAP.md](plugins/osHealth/ROADMAP.md) |
| mysqlHealth | active | [plugins/mysqlHealth/ROADMAP.md](plugins/mysqlHealth/ROADMAP.md) |
| mariadbHealth | active | [plugins/mariadbHealth/ROADMAP.md](plugins/mariadbHealth/ROADMAP.md) |
| pgsqlHealth | active | [plugins/pgsqlHealth/ROADMAP.md](plugins/pgsqlHealth/ROADMAP.md) |
| ufwApply | active | [plugins/ufwApply/ROADMAP.md](plugins/ufwApply/ROADMAP.md) |
| winHealth | scaffold | [plugins/winHealth/ROADMAP.md](plugins/winHealth/ROADMAP.md) |
| opnsenseHealth | scaffold | [plugins/opnsenseHealth/ROADMAP.md](plugins/opnsenseHealth/ROADMAP.md) |
| mongodbHealth | scaffold | [plugins/mongodbHealth/ROADMAP.md](plugins/mongodbHealth/ROADMAP.md) |
| redisHealth | scaffold | [plugins/redisHealth/ROADMAP.md](plugins/redisHealth/ROADMAP.md) |
| valkeyHealth | scaffold | [plugins/valkeyHealth/ROADMAP.md](plugins/valkeyHealth/ROADMAP.md) |
| rmqHealth | scaffold | [plugins/rmqHealth/ROADMAP.md](plugins/rmqHealth/ROADMAP.md) |
| esHealth | scaffold | [plugins/esHealth/ROADMAP.md](plugins/esHealth/ROADMAP.md) |
| vaultHealth | scaffold | [plugins/vaultHealth/ROADMAP.md](plugins/vaultHealth/ROADMAP.md) |
| k8sHealth | scaffold | [plugins/k8sHealth/ROADMAP.md](plugins/k8sHealth/ROADMAP.md) |
| zimbraHealth | scaffold | [plugins/zimbraHealth/ROADMAP.md](plugins/zimbraHealth/ROADMAP.md) |
| zimbraLdap | scaffold | [plugins/zimbraLdap/ROADMAP.md](plugins/zimbraLdap/ROADMAP.md) |
| pmgHealth | scaffold | [plugins/pmgHealth/ROADMAP.md](plugins/pmgHealth/ROADMAP.md) |
| postalHealth | scaffold | [plugins/postalHealth/ROADMAP.md](plugins/postalHealth/ROADMAP.md) |
| traefikHealth | scaffold | [plugins/traefikHealth/ROADMAP.md](plugins/traefikHealth/ROADMAP.md) |
| pritunlHealth | scaffold | [plugins/pritunlHealth/ROADMAP.md](plugins/pritunlHealth/ROADMAP.md) |
| wppconnectHealth | scaffold | [plugins/wppconnectHealth/ROADMAP.md](plugins/wppconnectHealth/ROADMAP.md) |
| upCheck | scaffold | [plugins/upCheck/ROADMAP.md](plugins/upCheck/ROADMAP.md) |
| lbPolicy | scaffold | [plugins/lbPolicy/ROADMAP.md](plugins/lbPolicy/ROADMAP.md) |
| sshNotifier | scaffold | [plugins/sshNotifier/ROADMAP.md](plugins/sshNotifier/ROADMAP.md) |

monokit1 components absorbed into base monokit2 instead of plugins: `daemon`,
`shutdownNotifier`, `uninstall` (reset), `logs`, `db` CLI, alarm/Redmine core,
updater, client/server API. `versionCheck` lives in osHealth's vlib.
