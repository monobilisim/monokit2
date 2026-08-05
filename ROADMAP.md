
# Feature Parity with <https://github.com/monobilisim/monokit>

## plugins

- [ ] __Elastic Search__
  - [ ] Status color check
  - [ ] Shard assignment check
  - [ ] Problematic allocation alarm

- [ ] __Kubernetes__
  - [ ] kube-vip health check
  - [ ] Node health check
  - [ ] Pod health check
  - [ ] RKE2 Ingress Nginx health check
  - [ ] Cert Manager health check
  - [ ] Cluster Api Cert health check
  - [ ] RKE2 health check
  - [ ] Kubernetes end of life check
  - [ ] ETCD backup and status check
  - [ ] Namespace Compliance check
  - [ ] Master Taint Compliance check
  - [ ] continuous mode alongside regular one shot mode with cron ?

- [-] __MySql__
  - [X] Up check
  - [X] Process count check
  - [ ] Certification waiting check
  - [ ] Cluster check (available in limeted form on MariaDB)
  - [X] Auto repair with timing
  - [X] PMM check
  - [~] MariaDB support (it is its own module in monokitv2 )

- [-] __MariaDB__
  - [X] Up check
  - [X] Process count check
  - [ ] Certification waiting check
  - [-] Cluster check
    - [X] Inaccessible clusters check
    - [ ] Cluster status check
    - [ ] Node status check
    - [X] Cluster sync check
    - [ ] Receive queue check
    - [ ] Flow Control check
    - Added Cluster certification check
  - [X] Auto repair with timing
  - [X] PMM check

- [- ] __OS Health__

- [-] __PostgreSql__ (currently for #34)
  - [X] Up check
  - [X] Process check
  - [x] uptime monitoring
  - [X] Check connections
  - [X] Version Check (Moved to OS health)
  - [X] Check running queries (missing alerts on long running queries)
  - [ ] Wall-g support
  - [ ] Patroni cluster monitoring
  - [X] PMM check

- [ ] __Proxmox Mail Gateway__

- [ ] __Postal__

- [ ] __Pritunl__

- [ ] __Redis__

- [ ] __RabbitMq__

- [ ] __Vault Service__

- [ ] __Windows OS__

- [ ] __WPPConnect__

- [ ] __Zimbra__

- [ ] __Zimbra Ldap__
