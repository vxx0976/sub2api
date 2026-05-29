# sub2api 高可用运维手册 (HA-RUNBOOK)

> 最后更新: 2026-05-29 (阶段0-4 + 监控 + CF LB + etcd/sentinel 5节点 + solid三副本 + 商户域名on-demand/API端点展现 + 服务商故障域 + §13失效矩阵)
> 适用: 半夜出状况时照着操作。先看「§2 一键体检」定位故障,再到「§4 故障 SOP」处理。
> 失效自愈能力速查见「§13 一/两节点失效矩阵」。

---

## §1 架构速览

> 本图聚焦 sub2api。inkmirage 共用同一批节点(及 etcd/sentinel/HAProxy/CF),另行讨论。

```
入口(DNS):  mayi.one ── CF LB(DNS-only灰云)──┬─► main 主源站(CN2)
                                              └─► hostdzire 备源站(别家)
            dsrrr.com + 商户域名 ──(无CF LB,单点)─► merchant   [on-demand TLS: merchant+main+hostdzire]
              │
应用入口Caddy: 各节点 reverse_proxy  →  primary=hostdzire:8080  backup=admin:8080  (health /health, 切 5-10s)
              ▼
sub2api 应用: hostdzire 生产主(16G)  ⇄ 兜底 ⇄  admin 备(3G, 兼测试/预发)
              │ 各连本机 HAProxy
HAProxy:     :5433 → 当前 PG 主      :6380 → 当前 Redis 主   (查 patroni role, 永远指当前主)
              ▼
数据(流复制): admin(PG主★+Redis主★) ──┬─► hostdzire 从(patroni托管, 同机房, lag0, 会被提主)
                                       └─► solid 从(standalone, 异地灾备 + 每日备份机, 不参与选主)
              │  patroni 选主(admin⇄hostdzire) + sentinel 切 Redis + HAProxy 透明跟随 (~30s)
协调(投票):  etcd ×5 + sentinel ×5 = {main, merchant, admin, hostdzire, bwh2}  quorum=3 容忍挂2
              │  ⚠️ 2026-05-29 第5票从 solid 迁到 bwh2(搬瓦工,独立服务商); solid 退出仲裁仅作数据从
监控:        node_exporter(main/merchant/admin/hostdzire/bwh2) → Prometheus
              告警: main check.sh(主) + bwh2哨兵(独立监 main) + solid 复制自检 → TG+邮件

故障域: {main,merchant,admin}=DMIT(main+merchant 同机房) | hostdzire=别家 | solid=异地 | bwh2=搬瓦工
节点纯度: bwh2 只投票+监控(无 sub2api/PG/Redis) ; solid 只数据从+备份(已无投票)
```

| 节点 | 配置/线路 | 入口Caddy | sub2api | inkmirage | PG | Redis | etcd | sentinel | 其他 |
|---|---|---|---|---|---|---|---|---|---|
| dmit-main | 2C/2G·CN2 | mayi.one源站 | — | — | — | — | ✅ | ✅ | 监控脚本 |
| dmit-merchant | 2C/2G·CN2 | dsrrr/商户 | — | — | — | — | ✅ | ✅ | — |
| dmit-admin | 4C/8G·T1 | api.dsrrr(反代hostdzire) | **backup兜底**(3G) | 备 | **主★** | **主★** | ✅ | ✅ | HAProxy |
| hostdzire | 6C/16G·别家 | mayi备/inkmir | **生产主★** + canary(dev) | **主★** | 从 | 从 | ✅ | ✅ | HAProxy、Prometheus、**测试栈(canary+独立test库,/opt/sub2api-test)** |
| solid | 8C/16G·异地 | — | — | — | 从(灾备) | 从(灾备) | ❌退出 | ❌退出 | 每日备份(2026-05仅数据从,待释放) |
| **bwh2** | 1C/0.5G·搬瓦工 | — | — | — | — | — | ✅ | ✅ | **第5票** + main独立哨兵 + node_exporter |

> mesh: main=10.88.0.1 merchant=.2 admin=.3 hostdzire=.4 **solid=.5(数据从) bwh2=.6(仲裁)**

**关键端口**: sub2api 8080 / PG 5432 / Redis 6379 / HAProxy PG 5433 / HAProxy Redis 6380 / patroni REST 8008 / etcd 2379-2380 / sentinel 26379 / HAProxy stats 127.0.0.1:8404

**集群名**: PG patroni scope=`17-main` namespace=`/postgresql-common/` ; Redis sentinel master=`mymaster`

**三层自动切换**: 入口(CF LB ~60s) / 应用(Caddy 5-10s) / 数据(patroni+sentinel ~30s)
**数据三副本**: admin主 →流复制→ hostdzire(同机房) + solid(异地) ; +solid每日pg_dump(30天)

**服务商故障域**: main / merchant / admin 同属 **DMIT**(其中 main+merchant 还同机房 CN2) ; hostdzire = **别家** ; solid = **异地另一家**。
→ 跨服务商「同时」宕(任一 DMIT 节点 + hostdzire/solid)几率极低,**忽略不计**。故 §13 只认真对待: ①单台任意挂 ②DMIT 内部双挂(main+merchant / main+admin / merchant+admin)。
⚠️ 反过来的相关性尾部风险: DMIT 三台(main+merchant+admin)同时挂 = etcd 仅剩 2 < quorum 3,patroni 无法选主、写入中断需人工(超出"一两台"范围,但同服务商需留意)。

**已接受的取舍**(站长拍板,非缺口,详见 §13): ①merchant/dsrrr.com/apex 商户域名入口单点 —— on-demand 多端点已尽力,不再加成本 ; ②mayi.one/CN2 GIA 入口 ~60s 切换 + 国内线路抖动 —— 为合规走 CN2,尽力而为、不追 1 分钟。
~~main+hostdzire / admin+hostdzire~~ 属跨服务商双挂,按上忽略。
**结论**: 排除上面两项后,应用层+数据层的任意一/两节点故障均 ~30s 内自愈,无其余开放缺口。

---

## §2 一键体检 (出问题先跑这个)

```bash
# 业务是否活
for u in "https://mayi.one|64.186.225.230" "https://dsrrr.com|64.186.226.91"; do
  d=${u%|*}; ip=${u#*|}; echo -n "$d: "; curl -s -o /dev/null -w 'HTTP %{http_code}\n' "$d" --resolve "${d#https://}:443:$ip" --max-time 8
done

# PG 集群 (谁是 Leader)
ssh dmit-admin "patronictl -c /etc/patroni/config.yml list"

# Redis 主 + sentinel
ssh dmit-main "redis-cli -p 26379 sentinel get-master-addr-by-name mymaster"

# HAProxy 后端 (admin 视角, UP 的是当前主)
ssh dmit-admin "curl -s 'http://127.0.0.1:8404/;csv' | awk -F, '/postgres|redis/{print \$1\"/\"\$2\"=\"\$18}'"

# etcd 健康
ssh hostdzire "docker exec etcd etcdctl --endpoints=http://10.88.0.1:2379,http://10.88.0.3:2379,http://10.88.0.4:2379 endpoint health"

# sub2api 两实例
ssh hostdzire "docker ps --filter name=sub2api --format '{{.Status}}'"
ssh dmit-admin "docker ps --filter name=^sub2api$ --format '{{.Status}}'"
```

---

## §3 凭据位置 (不要明文记在这)

| 凭据 | 位置 |
|---|---|
| PG superuser(postgres) | `/etc/patroni/config.yml` (admin/hostdzire) 的 authentication.superuser |
| PG replicator | 同上 replication 段 (值 `Fan@931025`) |
| sub2api DB/Redis 密码 | `/opt/sub2api/.env` (hostdzire/admin) |
| Redis requirepass | `/etc/redis/redis.conf` (admin) |
| 本地临时存的 patroni 密码 | 本机 `/tmp/patroni_pgpw.txt` (重启丢失,以 config.yml 为准) |

---

## §4 故障处理 SOP

### 4.1 hostdzire 挂 (生产 sub2api 宕)
**现象**: mayi.one/dsrrr.com 短暂抖动后恢复 (Caddy 5-10s 切 admin standby)
**自愈**: ✅ 自动。Caddy 健康检查切 backup=admin:8080
**人工**: 修 hostdzire。生产兜底现是 admin 的 **sub2api-backup**(限 3G;2026-05-29 拆双容器后,见 §12 发布流程)。hostdzire 挂且撞 4G+ 峰值时可临时给 backup 调大(牺牲 inkmirage standby):
```bash
ssh dmit-admin "cd /opt/sub2api && sed -i 's/mem_limit: 3g/mem_limit: 5g/;s/memswap_limit: 3g/memswap_limit: 5g/' docker-compose.yml && docker compose up -d sub2api-backup"
```
> 注: sub2api 后台「系统监控」的内存数字来自 ops_system_metrics 表,多实例(hostdzire 16G + admin 的 backup/canary 各报 admin 的 8G)混写且不记 hostname,所以可能显示任一实例的值。判断生产真实状态看 memory_total_mb=15988 那条(hostdzire)。(canary 连生产库,也会写这表,属正常噪声)
hostdzire 修好后 PG/Redis 会自动 rejoin (patroni/sentinel)。

### 4.2 admin 挂 (PG主+Redis主 宕)
**现象**: 几秒写失败后恢复 (patroni promote hostdzire ~30s)
**自愈**: ✅ 自动。patroni 选 hostdzire 为新 PG 主 + sentinel 切 Redis 主 + HAProxy 自动跟随
**确认**:
```bash
ssh hostdzire "patronictl -c /etc/patroni/config.yml list"   # hostdzire 应为 Leader
ssh dmit-main "redis-cli -p 26379 sentinel get-master-addr-by-name mymaster"  # 应为 10.88.0.4
```
**admin 修好后**: patroni 自动用 pg_rewind 把 admin 变回从 (无需人工)。Redis admin 自动变从。
**切回主到 admin** (可选,低峰做):
```bash
ssh hostdzire "patronictl -c /etc/patroni/config.yml switchover --leader pg-hostdzire --candidate pg-admin --force"
```

### 4.3 dmit-main 挂 (mayi.one 入口宕)
**现象**: mayi.one 不可访问; dsrrr.com 正常
**自愈**: ❌ 当前无 (CF LB 未部署)。
**人工**: 改 mayi.one DNS 临时指向 hostdzire(23.80.82.115) 或 merchant; 或修 main。
> 注: 阶段4 (CF LB / 多A记录) 做完后此项变自动。

### 4.4 dmit-merchant 挂 (dsrrr.com 入口宕)
同 4.3,改 dsrrr.com DNS 指 hostdzire 或 main。mayi.one 不受影响。

### 4.5 PG 手动切换 (patroni 没自动切时)
```bash
ssh dmit-admin "patronictl -c /etc/patroni/config.yml list"                          # 看状态
ssh dmit-admin "patronictl -c /etc/patroni/config.yml switchover --force"            # 优雅切(主从都活)
ssh dmit-admin "patronictl -c /etc/patroni/config.yml failover --force"              # 强制切(主已挂)
ssh dmit-admin "patronictl -c /etc/patroni/config.yml reinit 17-main pg-hostdzire"   # 从库坏了重建
```

### 4.6 Redis 手动切换
```bash
ssh dmit-main "redis-cli -p 26379 sentinel failover mymaster"   # 手动触发 failover
```

### 4.7 sub2api 滚动重启/升级 (零中断)
```bash
# 先 admin standby (backup,无流量)
ssh dmit-admin "cd /opt/sub2api && docker compose pull && docker compose up -d"
# 再 hostdzire 生产 (重启时 Caddy 切 admin standby 兜底)
ssh hostdzire "cd /opt/sub2api && docker compose pull && docker compose up -d"
```

---

## §5 关键服务重启 (谨慎)

```bash
# patroni (会重启该节点 PG,主节点重启=切主)
ssh <node> "systemctl restart patroni"
# HAProxy (重启不影响已建连接太久,应用会重连)
ssh <node> "systemctl restart haproxy"
# Caddy (reload 不断连)
ssh <node> "systemctl reload caddy"
# Redis sentinel
ssh <node> "systemctl restart redis-sentinel"
# etcd (docker)
ssh <node> "docker restart etcd"
```

⚠️ **不要** `systemctl start postgresql@17-main` —— 它已被 mask,PG 由 patroni 管理。

---

## §6 回滚 (HA 出大问题时退回直连)

如果 patroni/HAProxy 整体故障,想让 sub2api 直连 admin PG 应急:
```bash
# hostdzire sub2api 直连 admin PG (绕过本地 HAProxy)
ssh hostdzire "cd /opt/sub2api && sed -i 's/DATABASE_HOST=10.88.0.4/DATABASE_HOST=10.88.0.3/;s/DATABASE_PORT=5433/DATABASE_PORT=5432/;s/REDIS_HOST=10.88.0.4/REDIS_HOST=10.88.0.3/;s/REDIS_PORT=6380/REDIS_PORT=6379/' .env && docker compose up -d"
```
(admin PG 必须是当时的主; 用 patronictl list 确认)

---

## §7 待优化项 (非紧急)

1. **入口冗余** (阶段4): CF LB 或多 A 记录,解决 main/merchant 挂时入口自愈
2. ~~PG max_connections 200→500~~ ✅ 已完成 + 全面 PG/Redis 调优 (2026-05-28,见 §9)
3. ~~admin standby mem_limit 1.5G→3G~~ ✅ 已完成 (2026-05-28)
4. **api.dsrrr.com / sub2api.dsrrr.com**: 确认用途,前者打 admin standby(1.5G),后者打老容器:9000
5. **关 admin 公网 5432/6379** (阶段7): `iptables` DB-ACCESS chain 链尾 RETURN 改 DROP
6. **老容器 sub-sub2api-1(:9000)**: 站长确认是"另一套 sub2api 临时栈"(自带 sub-postgres-1/sub-redis-1,不碰生产库),~104MB 仍有外部流量。**可随时停,销毁等站长通知**。
7. ~~异地备份(解决"备份无异地")~~ ✅ 已完成 (2026-05-29,solid 每日 dump 异地推 jp1/sg1/alice,见 §14)。**剩 WAL 归档→R2 PITR**:把 RPO 从 24h 降到任意时间点,非紧急
8. **监控告警** (阶段6): inkmirage Grafana 加 PG lag/Redis/Caddy upstream 面板
9. ~~etcd/sentinel 第5票迁出 solid~~ ✅ 已完成 (2026-05-29,迁到 bwh2;见 BWH2-ONBOARD.md)。**solid 退费时只剩**: WAL→R2(#7)+ 摘 solid 的 PG/Redis 数据从 + 删 mesh peer(.5)+ check.sh/Prometheus 去掉 solid(见 BWH2-ONBOARD.md §8)
10. ~~解 admin 测试/兜底耦合~~ ✅ 已完成 (2026-05-29,测试环境整体迁到 hostdzire `/opt/sub2api-test`,canary 连独立 test 库;admin 只剩生产兜底,见 §12)
11. **config.yaml 残留公网IP**: hostdzire/admin 的 `/opt/sub2api/config.yaml` 里 database/redis host 写的是 `45.59.186.84`(admin 公网,绕 HAProxy),现被 `.env`(10.88.0.4:5433 本机HAProxy)覆盖无害,但属潜在雷 —— 改成与 .env 一致或删掉 config.yaml 的 DB/Redis 段
12. **ops_system_logs 表膨胀**: 单表 4.3GB = sub2api 库 76%,撑大 dump(463M)与库(5.7G)。给它(及 usage_logs)定保留期可大幅瘦身,加快备份/省异地盘

---

## §8 当前故障自愈能力 (实测验证)

| 故障 | 自愈 | RTO |
|---|---|---|
| hostdzire 挂 | ✅ Caddy 切 standby | 5-10s |
| admin 挂 | ✅ patroni+sentinel+HAProxy | ~30s |
| solid 挂 | ✅ 无感 | 0 |
| main 挂 | ❌ mayi.one 入口失效 (待阶段4) | 人工 |
| merchant 挂 | ❌ dsrrr.com 入口失效 (待阶段4) | 人工 |
| main+merchant 同挂 | ❌ 两入口全失 (CN2同机房,待阶段4) | 人工 |
| admin+hostdzire 同挂 | ❌ 需 solid PITR | 小时级 |

---

## §9 数据库调优记录 (2026-05-28)

> 通过 patroni 全局 dcs 配置下发(主从一致 + 持久化 + 开机自启)。改法见下,改后用 `patronictl restart` 滚动重启(先 replica,leader 用 switchover→restart→switchback 避免中断)。

### PostgreSQL (基准 admin 8G,failover 到 hostdzire 16G 时偏保守但安全)
| 参数 | 调优前 | 调优后 | 目的 |
|---|---|---|---|
| shared_buffers | 128MB | **1536MB** | 缓存大幅提升(需预热) |
| max_connections | 200 | **500** | 突发连接承载 |
| effective_cache_size | 4GB | 4GB | 规划器提示 |
| work_mem | 4MB | 6MB | 排序/聚合(×连接数,勿过大) |
| maintenance_work_mem | 64MB | 256MB | vacuum/建索引加速 |
| wal_buffers | 4MB | 16MB | 写入吞吐 |
| max_wal_size / min_wal_size | 1GB/80MB | 4GB/1GB | 减少 checkpoint 频率 |
| random_page_cost | 4 | 1.1 | SSD |
| effective_io_concurrency | 1 | 200 | SSD 并发 IO |

改参数(非交互): `curl -s -XPATCH http://10.88.0.3:8008/config -d '{"postgresql":{"parameters":{"参数":"值"}}}'`
重启类参数(shared_buffers/max_connections/wal_buffers)改后需 `patronictl restart`。

### Redis (admin 主 + hostdzire 从)
| 参数 | 调优前 | 调优后 | 目的 |
|---|---|---|---|
| maxmemory | 0(无限) | **2GB** | 防爆(数据实占仅 ~10MB,纯保险) |
| maxmemory-policy | noeviction | **volatile-lru** | 满了只淘汰有TTL的(token/缓存),保护无TTL的 sched 调度数据 |
| activedefrag | off | yes | 碎片整理(原碎片率 2.59) |

改法(零中断): `redis-cli -a <pw> config set <参数> <值>; redis-cli -a <pw> config rewrite` (主从都做)

⚠️ **不可用 allkeys-lru**: redis 里有 `sched:ready:*` 等 TTL=-1 的调度持久数据,会被误删。

### 未来可继续 (非必须)
- hostdzire 做主时 shared_buffers 可单独提到 4GB(它 16G),但 patroni 全局一致,需用节点级 tags 覆盖,暂保持统一 1.5GB
- 若突发仍紧,考虑 PgBouncer 连接池(transaction 模式)替代直连,大幅降低 PG 连接数

---

## §10 监控告警 (2026-05-29)

### 数据采集
- node_exporter (main/merchant/admin/hostdzire/**bwh2** :9100,bwh2 原生绑 10.88.0.6) + redis_exporter (admin/hostdzire :9121) + patroni 自带 metrics (:8008/metrics)
- → inkmirage Prometheus (hostdzire 容器,`/opt/inkmirage/docker/prometheus/prometheus.yml` 的 ha-* jobs;ha-node 含 bwh2)
- 改 prometheus 配置后: `docker exec inkmirage-prometheus-1 promtool check config /etc/prometheus/prometheus.yml` 校验,再 `docker restart inkmirage-prometheus-1` (没开 hot reload)

### 告警 (TG + 邮件,状态变化才发)
- **主监控**: dmit-main `/opt/ha-monitor/check.sh` (cron 每分钟)。14 项: PG有主/patroni从数/mesh可达(含bwh2)/mayi+dsrrr业务/redis主/磁盘(含bwh2)
- **bwh2 哨兵** (2026-05-29 新增): bwh2 `/opt/ha-monitor/check.sh` (cron 每分钟,TG/邮件带 `[bwh2哨兵]` 前缀)。独立服务商,专盯 main存活(公网+mesh)/main源站mayi/mayi端到端/PG有主/Redis主 —— **补上"main 挂则主监控随之失效"的盲区**
- **solid 灾备自检**: solid `/opt/ha-monitor/solid-check.sh` (cron 每分钟,查复制 streaming+lag;solid 释放时随机器下线)
- 凭据: `/opt/ha-monitor/secrets` (TG_TOKEN/TG_CHAT/ALERT_EMAIL/REDIS_PW,chmod 600;bwh2 同款 + `/etc/msmtprc`)
- 状态文件: `/opt/ha-monitor/state/*` (每项一个,去重防刷屏)

### 常用操作
- 手动跑一次: `bash /opt/ha-monitor/check.sh`
- 看当前状态: `cat /opt/ha-monitor/state/*`
- 静默某项(改阈值/暂停): 编辑 check.sh 对应行
- 测试告警链路: `echo FAIL > /opt/ha-monitor/state/biz_mayi` 然后跑(会发恢复告警)
- ✅ 主监控在 main 的盲区已补: bwh2 哨兵(独立服务商)每分钟独立监 main,main 挂时由 bwh2 报警(见上)

### 待优化
- Grafana 可视化面板(导入 node-exporter dashboard 1860 + patroni/redis 面板),数据已在 Prometheus,只差画图
- ~~告警冗余: solid 跑 check.sh~~ ✅ 已用 bwh2 哨兵实现(独立服务商,比 solid 更合适;solid 反正要释放)

---

## §11 入口冗余 CF Load Balancer (2026-05-29, mayi.one)

### 架构
- **DNS-only LB**(灰云,proxied=false): 流量直连源站 IP 保持 CN2,CF 只在 DNS 层健康检查+切换
- mayi.one: default pool=main(CN2 GIA) / fallback pool=hostdzire(T1)
- CF 资源 id: LB=`d4ddaadc7bfa42265a795c3f42bf9ad7` monitor=`15941481e1675895834795364f60a972` pool-main=`ee706bf77b39c6863161aeb74a90f761` pool-hostdzire=`2ce1734ecb4f6ce4c2d713e4468cefe8`
- account=`012dc037...` zone(mayi.one)=`bd2bdd0aea70da99797bca369c5fe2c9`

### 源站
- main: Caddy mayi.one vhost(原有) → 反代 hostdzire sub2api(mesh)
- hostdzire: Caddy mayi.one vhost → 本地 sub2api;证书 /etc/caddy/certs/mayi.one/(从 main 同步)
- 证书续期: hostdzire `/opt/sync-mayi-cert.sh` cron 每天4点拉 main 证书(用 hostdzire→main mesh ssh key,不依赖 CF token)

### ⚠️ 已知局限
- CF 健康检查节点在**海外**: 只能感知"机房整个断网"(切fallback),**感知不到 CN2 GIA 国内线路抖动**(机房活着国内卡时不切)
- 国内线路抖动仍靠 §10 国内监控(待) + 手动应急

### 故障切换/回滚
- 自动: main origin 健康检查失败(60s) → 解析切 hostdzire
- 手动演练: `curl -XPATCH .../pools/{pool-main} -d '{"origins":[{"name":"main","address":"64.186.225.230","enabled":false}]}'` (enable=true 恢复)
- 整体回滚(弃用LB): 删 LB + 重建 A 记录 `mayi.one A 64.186.225.230 proxied=false`
- CF 操作需 API token(权限: Account LB Monitors&Pools + Zone LB + Zone DNS)

### dsrrr.com (未做)
- 要做需: merchant 作 primary origin(第3个endpoint,+$5/月) + hostdzire 作 fallback + hostdzire Caddy 加 dsrrr.com vhost+证书

---

## §12 商户自定义域名接入 (2026-05-29)

### 标准接入强制项 (两条铁律,影响 HA)
1. **绑定域名 = 强制子域 CNAME → mayi.one**: 商户验证/绑定域名时,引导其将 `api.<商户域名>` 子域 **CNAME 到 mayi.one**(而非 A 记录指 merchant)。这样商户流量跟随 mayi.one 的 CF LB(main 主 / hostdzire 备),**自动继承入口冗余**,merchant 挂时商户子域不受影响。
2. **api_base_url 强制写入子用户 API Key 管理页**: 子用户访问商户域名时,`mergeResellerBranding()`([backend/internal/web/embed_on.go:247](backend/internal/web/embed_on.go:247))按 Host 注入 `api_base_url="https://<商户域名>"`,**强制**覆盖到该商户子用户的 API Key / 密钥查询 / 接入文档页端点 base URL —— 子用户看到的始终是商户自己的域名,主站端点不外泄。
> 残留单点: `dsrrr.com` 本站 + **apex 根域**(不能 CNAME,只能 A 记录指 merchant) + 旧 A 记录商户,仍落在 merchant 单点。

### 接入方式
- **推荐**(继承入口冗余): 商户验证域名时 **子域 CNAME → mayi.one** (跟随 CF LB: main主/hostdzire备) + `_domain-verify.<域名>` TXT 验证所有权
  - 子域(api.商户域名)可 CNAME; 根域(apex)不能 CNAME,只能 A 记录指 merchant(单点无冗余)
  - 旧方式(A记录指 merchant)仍兼容,只是 merchant 单点
- **签证书**: 商户流量到达的源站 Caddy on-demand TLS 动态签
  - ask 端点 `http://10.88.0.4:8080/api/v1/caddy/ask-tls` → sub2api 查 reseller_domains.verified, 只有 TXT 验证过的放行(防滥用)
  - on-demand 已配源站: **merchant(原有) + main + hostdzire**(2026-05-29加全局 on_demand_tls + :443 catch-all)
  - ⚠️ 真实签证书要求域名解析到该源站: 商户 CNAME mayi.one 后 main/hostdzire 才能签; 现状指 merchant 则只 merchant 签

### API 端点展现 (子用户看到商户域名)
- **代码**: `backend/internal/web/embed_on.go` 的 `mergeResellerBranding()` 注入 `m["api_base_url"]="https://"+商户域名`
- **效果**: 商户子用户访问商户域名 → API Key/密钥查询/接入文档页的端点 base URL = 商户自己域名(非主站)
- **隔离**: DomainDetect 按 Host 注入 → 主站访问不经过(主站端点不变) + 商户之间按各自 Host 互不可见 + 数据隔离(api_key.user_id / user.parent_id)未动
- 对存量商户(A记录指merchant)也生效(只要访问的是商户域名)
- 前端无需改(它本就读 `__APP_CONFIG__.api_base_url`)

### 发布流程 (站长工作流, 2026-05-29 测试环境隔离到 hostdzire 后)
> **三处 sub2api**:① 测试栈 = hostdzire `/opt/sub2api-test/`(canary dev + **独立 test 库** sub2api-testdb + testredis,绑 `10.88.0.4:8081`,服务 api.dsrrr.com)② 生产主 = hostdzire `/opt/sub2api/`(:8080)③ 生产兜底 = admin `/opt/sub2api/` 的 `sub2api-backup`(`10.88.0.3:8080`)。
> ✅ **canary 连的是隔离 test 库,migration 只动 test、绝不碰生产**(已验证 0 连接到生产)。admin 不再跑测试码。

1. `push dev` → CI(dev-build.yml) build `vxx0976/sub2api:dev`
2. **测试**: `ssh hostdzire "cd /opt/sub2api-test && docker compose pull && docker compose up -d sub2api-canary"` → 在 **api.dsrrr.com** 验证(对着 test 库,migration 安全)
3. 验证 OK → **部署生产主**: `ssh hostdzire "cd /opt/sub2api && docker compose pull && docker compose up -d"`(此时才在生产库跑 migration)
4. **部署生产兜底**: `ssh dmit-admin "cd /opt/sub2api && docker compose pull && docker compose up -d"`(admin 现单 service,直接 up 即可)

🔄 **重置 test 库数据**(可选,想用更新的真实数据测时): 重跑 §14 恢复演练(solid 最新 dump → restore 进 sub2api-testdb)。test 库是某次快照,会随时间与生产漂移。
🔧 回滚 admin 兜底拆分: `ssh dmit-admin "cd /opt/sub2api && cp docker-compose.yml.bak.split docker-compose.yml && docker rm -f sub2api-backup && docker compose up -d"`(退回单容器)。

---

## §13 一/两节点失效矩阵 (实际服务商拓扑, 2026-05-29)

> **前置假设**(站长确认,见 §1 服务商故障域):
> 1. 跨服务商「同时」宕忽略不计 → 只看: 单台任意挂 + DMIT 内部双挂(main+merchant / main+admin / merchant+admin)。`main+hostdzire`、`admin+hostdzire` 跨服务商,忽略。
> 2. 商户按标准接入(子域 CNAME mayi.one,见 §12)→ merchant 挂时商户子域随 CF LB 自愈,仅 `dsrrr.com` 本站 + apex + 旧A记录商户 留 merchant 单点。

### 单台宕机
| 宕机 | 影响 & 自愈 | RTO | <1min |
|---|---|---|---|
| solid | 异地灾备从丢失,无感 | 0s | ✅ |
| hostdzire | Caddy 切 admin standby;PG/Redis 主在 admin 不动;mayi 默认源站 main 仍在 | 5–10s | ✅ |
| admin | patroni 提 hostdzire 为主 + sentinel 切 Redis + HAProxy 跟随 | ~30s | ✅ |
| main | mayi.one + 挂它的商户子域靠 CF LB 切 hostdzire;**监控 check.sh 随 main 失效**(solid 自检兜底) | ~60s+TTL | ⚠️ 卡线/偏超 |
| merchant | 标准接入商户子域随 CF LB 不受影响 ✅;**`dsrrr.com`本站 + apex + 旧A记录商户 入口失效,手动改DNS** | dsrrr 人工 | ❌(仅 dsrrr/apex) |

### DMIT 内部双挂 (同服务商,realistic)
| 组合 | 影响 & 自愈 | <1min |
|---|---|---|
| main + admin | 数据 hostdzire 提主 ~30s ✅;mayi 借 CF LB ~60s;dsrrr(merchant在)正常;etcd 剩3 ✅ | ⚠️ 数据✅ / mayi 卡线 |
| main + merchant | mayi+商户子域 借 CF LB→hostdzire ~60s;**dsrrr本站+apex 失效**;数据/应用不动;etcd 剩3 ✅ | ❌(dsrrr/apex)+ mayi卡线 |
| merchant + admin | 数据 hostdzire 提主 ~30s ✅;商户子域 CF LB→main ✅;**dsrrr本站+apex 失效**;etcd 剩3 ✅ | ❌(仅 dsrrr/apex) |

### 已接受的设计取舍 (站长拍板,不再当缺口处理)
- **merchant / dsrrr.com / apex 商户域名入口单点**: 不再投入(不增加 dsrrr.com 的 CF LB 成本)。on-demand 多端点(merchant+main+hostdzire 源站)已尽力提高商户域名可用性,到此为止。**接受**: merchant 挂时 dsrrr.com 本站 + apex 根域商户需手动改 DNS。
- **mayi.one / CN2 GIA 线路**: 为国内合规走 CN2,定位"能用就用、不能用拉倒"的尽力而为线路。**接受**: CF LB ~60s+TTL 的入口切换、以及海外健康检查感知不到国内线路抖动,均不再追求 1 分钟内。

### 已澄清,无开放待办
- **`10.88.0.4` 是 WireGuard mesh 分配地址,非脆弱硬编码** (mesh 网段 `10.88.0.x`,见 §6: .4=hostdzire / .3=admin)。ask-tls 指向哪台是稳定事实;商户域名既定为尽力而为,此项不再追。
- **check.sh 是告警发送方本身(必要,非冗余)**: §10 cron 每分钟跑 14 项 → TG+邮件,发告警的就是它。唯一盲区是「main 自己挂 → 告警器随之挂」,而 §10 已记录该场景由 CF LB/DNS 切换暴露 + solid 自检独立兜底,站长已接受。无新待办。

### 一句话结论 (按站长决定收敛后,最终)
**应用层 + 数据层: 所有现实的一/两节点故障都在 ~30s 内自愈** —— 单台 admin/hostdzire/solid,及 DMIT 内部双挂(main+merchant / main+admin / merchant+admin),app 始终有 hostdzire 主或 admin standby 兜底、DB 始终能 patroni 提 hostdzire 主。**唯二不在 1 分钟内自愈的就是两条已接受的入口取舍**(merchant 入口单点 + mayi CN2 尽力而为)。**无其余开放缺口或待办。**

---

## §14 备份与异地容灾 (2026-05-29)

> HA(节点挂)≠ 备份(数据坏)。流复制会把误删/逻辑损坏复制到所有从库,只有备份能回滚。本节是数据持久性的底线。

### 备份生成 (solid)
- `0 3 * * * /opt/sub2api/backup_and_mail.sh` (solid):pg_dump -Fc **sub2api(~463MB/天)** + inkmirage(~0.3MB) + Redis RDB(~4MB) → `/root/backups/sub2api/`,本地留 **30 天**,完成发报告邮件。dump 从 admin(10.88.0.3)经 mesh 拉,不依赖 solid 是从库 → 可搬。

### 异地多副本 (2026-05-29 新增,补"备份无异地"缺口)
- `20 3 * * * /opt/sub2api/offsite_push.sh` (solid):dump 后 rsync 推到 3 台异地、不同地理:

| 目标 | 位置/服务商 | IP | 保留 | 盘 |
|---|---|---|---|---|
| jp1 | 日本千叶 / Oracle | 155.248.176.74 | 30天 | 35G |
| sg1 | 新加坡 / Oracle | 168.138.172.122 | 30天 | 34G |
| alice | 香港 / Alice(异厂) | 5.102.125.51 | 10天 | 6.7G(盘小) |

- 路径 `/opt/offsite-backup/sub2api/`;solid 用专用密钥 `/root/.ssh/offsite_bkup`(目标 authorized_keys 限 `from=154.3.224.215`)。
- 推送任一失败 → TG 告警;日志 `/var/log/offsite_push.log`。
- **服务商多样性**: jp1+sg1 同属 Oracle(免费实例有被回收风险,会一起没),alice 是唯一非 Oracle,专门破这个相关性。

### 手动操作
```bash
ssh solid /opt/sub2api/offsite_push.sh          # 手动补推一次
ssh solid 'tail /var/log/offsite_push.log'      # 看推送结果
for h in jp1 sg1 alice; do echo "== $h =="; ssh $h 'ls -lh /opt/offsite-backup/sub2api/ | tail -3'; done
```

### 恢复演练 ✅ 已做 (2026-05-29)
- **结果 PASS**: 最新 dump(pg_sub2api_20260529_030001,463M)→ hostdzire 上隔离 test PG(`sub2api-testdb` 容器,postgres:17,127.0.0.1:5434)`pg_restore` 成功:exit 0、1m45s、**81 表 / 5.7GB**、usage_logs 102万行、最新记录 `2026-05-28 21:30 UTC`=备份时刻(完整无截断)。恢复期间 hostdzire 生产 sub2api + PG 复制 lag 0,无影响。
- 复跑演练:
```bash
# solid 推最新 dump 到 hostdzire(mesh, 用 offsite_bkup key)→ 容器内 pg_restore
ssh solid 'rsync -az -e "ssh -i /root/.ssh/offsite_bkup" /root/backups/sub2api/$(ls -t /root/backups/sub2api/pg_sub2api_*.dump|head -1|xargs basename) root@10.88.0.4:/tmp/'
ssh hostdzire 'docker cp /tmp/<dump> sub2api-testdb:/tmp/d.dump && docker exec sub2api-testdb pg_restore -U sub2api -d sub2api --no-owner --no-privileges -j2 /tmp/d.dump'
```
- ⚠️ **`ops_system_logs` 单表 4.3GB = 整库 76%**:撑大了 dump(463M)和库(5.7GB)。加保留期可大幅瘦身(备份更快、异地机更省)。**待办**: 给 ops_system_logs(及 usage_logs)定保留策略。
- 这个 `sub2api-testdb`(hostdzire,带真实数据)同时作为 §1 测试环境的隔离库(见下,待接入 canary)。

### 仍待优化 (RPO)
- 当前 RPO = **最长 24h**(每日快照)。要更小需 **WAL 归档→R2 PITR**(§7 #7),可恢复到任意时间点。daily 异地快照已大幅改善,但 PITR 仍是下一步。
- solid 仍是异地**流复制从**(RPO≈0 的实时副本),直到将来整机释放;释放前需把 dump+push 任务迁到 hostdzire(见 BWH2-ONBOARD §8 思路)。
