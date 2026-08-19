# sub2api 高可用运维手册 (HA-RUNBOOK)

> 最后更新: **2026-08-19** (jp/sg 被甲骨文回收 → 异地备份改推 spartan + 修「不可达即静默」的备份巡检漏报)
> 上一次: 2026-05-29 (bwh2接入第5票 + 备份app→R2 + 测试环境隔离hostdzire + admin公网DB收口 + 残留清理 + **solid 已释放**)
> 当前节点(5): main / merchant / admin / hostdzire / bwh2(solid 异地从已退役)
> 集群外备份收端: **spartan**(172.83.153.98,不入 mesh、不属 HA 集群,只收 pg_dump)
> 适用: 半夜出状况时照着操作。先看「§2 一键体检」定位故障,再到「§4 故障 SOP」处理。
> 失效自愈能力速查见「§13 一/两节点失效矩阵」。

---

## §1 架构速览

> 本图聚焦 sub2api。**inkmirage + 监控栈(Prometheus/Grafana/Loki)已于 2026-06-22 整体迁到 spartan(SpartanHost 西雅图,公网 172.83.153.98 / mesh 10.88.0.5,自带本地 PG),不再共用 sub2api 节点;spartan 仅为抓取指标入 mesh,非 PG/etcd/sentinel 成员。** sub2api 集群已 DROP 掉 inkmirage 库。
> 📊 可视化架构图: [docs/architecture-cn.png](docs/architecture-cn.png)(中文) · [docs/architecture-en.png](docs/architecture-en.png)(英文)。下方 ASCII 图为**权威源**,改架构时以文字为准、图后补。

```
入口(DNS):  mayi.one ── CF LB(DNS-only灰云)──┬─► main 主源站(CN2)
                                              └─► hostdzire 备源站(别家)
            dsrrr.com + 商户域名 ──(无CF LB,单点)─► merchant   [on-demand TLS: merchant+main+hostdzire]
            api.dsrrr.com ───────────────────────► admin Caddy → 兜底(**生产库**,真实管理/预发)
            canary.dsrrr.com ────────────────────► hostdzire Caddy → canary(**隔离 test 库**,测新版本)
              │
应用入口Caddy: 各节点 reverse_proxy  →  primary=hostdzire:8080  backup=admin:8080  (health /health, 切 5-10s; 2026-06-11 main/merchant 加 dial_timeout 2s + lb_retries 3 + lb_try_duration 5s,秒级mesh抖动自动重试admin,黑洞注入实测0失败,改前备份 Caddyfile.bak-20260611)
              ▼
sub2api 应用: hostdzire 生产主(16G)  ⇄ 兜底 ⇄  admin 备(3G, 兼测试/预发)
              │ 各连本机 HAProxy
HAProxy:     :5433 → 当前 PG 主      :6380 → 当前 Redis 主   (查 patroni role, 永远指当前主)
              ▼
数据(流复制): admin(PG主★+Redis主★) ──► hostdzire 从(patroni托管, lag0, 会被提主)
              │  patroni 选主(admin⇄hostdzire) + sentinel 切 Redis + HAProxy 透明跟随 (~30s)
协调(投票):  etcd ×5 + sentinel ×5 = {main, merchant, admin, hostdzire, bwh2}  quorum=3 容忍挂2
监控:        node_exporter(main/merchant/admin/hostdzire/bwh2) → Prometheus
              告警: main check.sh(主) + bwh2哨兵(独立监 main) → TG+邮件 ; 备份: app自带→R2(每日,UI恢复)

故障域: {main,merchant,admin}=DMIT不同机房(main/merchant=三网优化, admin=T1) | hostdzire=别家 | bwh2=搬瓦工
节点纯度: bwh2 只投票+监控 ; admin=数据主(PG/Redis)+sub2api兜底 ; hostdzire=sub2api生产主+PG/Redis从+测试栈
(solid 异地从已于 2026-05-29 释放,退出集群)
```

| 节点 | 配置/线路 | 入口Caddy | sub2api | inkmirage | PG | Redis | etcd | sentinel | 其他 |
|---|---|---|---|---|---|---|---|---|---|
| dmit-main | 2C/2G·CN2 | mayi.one源站 | — | — | — | — | ✅ | ✅ | 监控脚本 |
| dmit-merchant | 2C/2G·CN2 | dsrrr/商户 | — | — | — | — | ✅ | ✅ | — |
| dmit-admin | 4C/8G·T1 | **api.dsrrr**(→兜底/生产库) | **backup兜底**(3G) | — | **主★** | **主★** | ✅ | ✅ | HAProxy |
| hostdzire | 6C/16G·别家 | mayi备/**canary.dsrrr** | **生产主★** + canary(dev) | — | 从 | 从 | ✅ | ✅ | HAProxy、**测试栈(canary+独立test库,/opt/sub2api-test)**(Prometheus 已迁 spartan) |
| ~~solid~~ | 8C/16G·异地 | — | — | — | — | — | — | — | **已释放 2026-05-29**(退集群,数据从+第二备份均停) |
| **bwh2** | 1C/0.5G·搬瓦工 | — | — | — | — | — | ✅ | ✅ | **第5票** + main独立哨兵 + node_exporter |

> mesh: main=10.88.0.1 merchant=.2 admin=.3 hostdzire=.4 bwh2=.6(仲裁) ; .5 原 solid 已释放(空闲)

**关键端口**: sub2api 8080 / PG 5432 / Redis 6379 / HAProxy PG 5433 / HAProxy Redis 6380 / patroni REST 8008 / etcd 2379-2380 / sentinel 26379 / HAProxy stats 127.0.0.1:8404

**集群名**: PG patroni scope=`17-main` namespace=`/postgresql-common/` ; Redis sentinel master=`mymaster`

**三层自动切换**: 入口(CF LB ~60s) / 应用(Caddy 5-10s) / 数据(patroni+sentinel ~30s)
**数据双副本**(流复制): admin主 → hostdzire(别家从,patroni 会提主)。~~solid 异地从~~ 已释放(2026-05-29)
**备份**: sub2api 自带 →**Cloudflare R2**(每日,UI 一键恢复)。~~solid→alice 第二副本~~ 随 solid 释放已停(站长定,只留 app→R2);详见 §14

**服务商/机房故障域**: main / merchant / admin 同属 **DMIT 但不同机房**(main/merchant = **三网优化(CN2)**机房 ; admin = **T1** 机房) ; hostdzire = **别家**。(solid 异地从已 2026-05-29 释放)
→ 因为不同机房,**单个 DMIT 机房挂只带走其中的票,剩余仍 ≥ quorum 3**(如 三网优化机房挂带走 main+merchant 2票 → admin+hostdzire+bwh2 = 3 ✓)。跨服务商「同时」宕几率极低,忽略。§13 认真对待: ①单台任意挂 ②机房内/跨机房双挂。
⚠️ 仅剩的相关性尾部风险: **DMIT 账号级 / 全局网络级**事件(同时打掉三网优化+T1 两机房三票)= etcd 仅剩 hostdzire+bwh2=2 < quorum 3 → 写入中断需人工。比"单机房挂"罕见得多;要彻底免疫需 ≥3 票在非 DMIT(即再加独立服务商投票),复杂度不值,留意即可。

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
| PG replicator | `/etc/patroni/config.yml` (admin/hostdzire) 的 replication 段 (**口令不在此明文记录;以节点 config.yml 为准**) |
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
> ⚠️ **glibc 异构提醒**: hostdzire=glibc **2.35**、admin=**2.36**,库为 `en_US.UTF-8`。短时切换无碍(2.35↔2.36 排序实际未变)。**若 hostdzire 长期当主、期间有大量文本写入**,切回 admin 后在**主库**跑一次:`REINDEX DATABASE sub2api;` 然后 `ALTER DATABASE sub2api REFRESH COLLATION VERSION;`,消除潜在排序偏差。背景见 §7.14。

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
> **发新版本**走完整发布流程(§12:canary 测 → 生产主 → 兜底)。下面仅用于**已验证镜像的纯滚动重启**。
```bash
# 先 admin 兜底(backup,无流量),再 hostdzire 生产(重启秒级窗口 Caddy 切 admin 兜底)
ssh dmit-admin "cd /opt/sub2api && ./update.sh"
ssh hostdzire "cd /opt/sub2api && ./update.sh"
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
5. ~~关 admin 公网 5432/6379~~ ✅ 已完成 (2026-05-29): 先在 DB-ACCESS 链顶加 `-i lo ACCEPT`(保回环),再把链尾 RETURN 改 DROP;`netfilter-persistent save` 持久化。放行=回环+mesh(10.88.0.0/24@wg0)。legacy main/merchant 公网白名单已删(2026-05-29,确认无公网连接)。前置: 已先把 solid redis 从公网改走 mesh(见 #13)
6. ~~**老容器 sub-sub2api-1(:9000)** 另一套 sub2api 临时栈~~ ✅ **已停 2026-06-19**(`/opt/sub` 整套 `docker compose stop`,容器+数据卷保留,`docker compose start` 可恢复;释放 ~364M,unless-stopped 重启宿主不复活)。停前仅 bwh1 每 ~30-60s 轮询 :9000(dashboard/探活,非真实用户——找出来停掉)。**销毁(`down -v` + 删 /opt/sub)待站长通知**。
7. ~~异地备份 / WAL→R2 PITR~~ ✅ 已解决,且发现**重复**: sub2api **自带 S3/R2 备份**(后台数据备份页,每日 `0 2`→R2,留14天/3份,UI可下载恢复)早已配好。本次手搭的 solid→jp1/sg1/alice 管道是冗余,**待退役**(见 §14)。PITR 决定**跳过**(每日 R2 + canary 测 migration 已足够,不为分钟级 RPO 加复杂度)
8. **监控告警** (阶段6): inkmirage Grafana 加 PG lag/Redis/Caddy upstream 面板
9. ~~etcd/sentinel 第5票迁出 solid~~ ✅ 已完成 (2026-05-29,迁到 bwh2)。**solid 退费时只剩**(备份已不依赖 solid,app→R2 接管): 摘 solid PG/Redis 数据从 + 删 mesh peer(.5)+ check.sh/Prometheus 去掉 solid + 退役 offsite 管道(见 §14)。✅ **2026-06-19 实测确认 solid 已彻底出集群**: Patroni 仅 `admin`+`hostdzire` 两成员;`154.3.224.215` 不再跑 patroni/PG/redis(现为另一台 `Umasou`)。ssh 别名 `solid` + 本条 todo 可清理 → ✅ **2026-08-19 已清**(本机 `~/.ssh/config` 删掉 solid/jp/sg/alice 四个失效别名;注意 `154.3.224.215` 现已被服务商再分配给无关主机 `Umasou`)
10. ~~解 admin 测试/兜底耦合~~ ✅ 已完成 (2026-05-29,测试环境整体迁到 hostdzire `/opt/sub2api-test`,canary 连独立 test 库;admin 只剩生产兜底,见 §12)
11. ~~config.yaml 残留公网IP~~ ✅ 已修 (2026-05-29,admin/hostdzire 的 config.yaml db/redis host 改为本机 HAProxy 与 .env 一致)
12. **ops_system_logs 表膨胀**: ✅ 清理已开启并生效(2026-06-19,后台「系统监控→数据保留」,retention 30天)——autovacuum 健康(死元组 4–10%),活行已降到 ~30天量。但库仍 **8.2G**:30天日志量本身大,且 DELETE 不自动还盘。想再瘦(让备份/恢复/复制更快):retention 降 **7–14 天** + 一次性 `VACUUM FULL ops_system_logs`(+ `ops_error_logs`),**择 CN 夜里**做(会给 flaky hostdzire 灌 WAL,白天别做)。磁盘 158G 用 17%,不急。
13. ~~一轮残留清理~~ ✅ (2026-05-29): 5节点清悬空镜像共 ~10.7GB;admin 删孤儿卷+停用空闲 Caddy;main/merchant 删孤儿卷 `sub2api_sub2api_data`+收敛 .bak;solid 清 etcd 残留(unit/datadir/bin)。**solid redis 从公网(45.59.186.84)改走 mesh(10.88.0.3)** —— 消除公网复制通道,也是关 #5 公网DB端口的前置。残留待办仅剩 main/merchant 的 legacy DB-ACCESS 白名单(可删,无害)
14. **HA 对 glibc 异构 (2026-05-31 审计发现)**: admin=Debian glibc **2.36**、hostdzire=Ubuntu glibc **2.35**,所有库 `en_US.UTF-8`(locale 感知排序)。日常 App 经 HAProxy 只连主库 admin(2.36,stored=actual 匹配)→ **零影响**;隐患仅在**故障切换到 hostdzire 后**(用 2.35 排序导航 2.36 建的索引)。但 2.35→2.36 之间 en_US.UTF-8 排序数据实际未变(真正大改在 2.27→2.28),PG 仅按版本号字符串保守报警 → **实际风险低,建议接受 + 记录**。切换补救已写入 §4.2。长期最干净是两机 glibc 对齐(重装 hostdzire OS,代价大,暂不做)。复制层无碍:物理流复制按字节拷页,与排序无关。
15. **2026-06-19 评审结论(成本/可靠性重审)**: ①**单机整合方案否决**——计费不许退化到小时、HA 形状本就对,就地演进不重建(别再探索"砍 HA 换单机")。②**per-IP 限流否决**——API 服务单 IP 高频正常(站长:5K rpm 都正常),限流会误伤;分布式洪水靠 CF 不靠 app。③**trusted_proxies 未设**: gin `c.ClientIP()` 取到前置 Caddy mesh IP → **API Key 的 IP 白/黑名单当前按 Caddy IP 判断 = 失效**;要用该 ACL 才修(`server.trusted_proxies: ["10.88.0.0/24"]`,会改 ACL 判定)。④**CF 逃生**: 被打时把 mayi.one 的 CF LB「已代理」拨开变橙云(手动一键,丢 CN2 换 L3/4 抗D;免费版 + 源站 IP 已泄 → 仅挡走 CF 的流量)。⑤**✅ Redis 主已 pin 回 admin**(2026-06-19,`sentinel failover mymaster`,恢复"PG主+Redis主同机 admin"不变量;此前漂到 hostdzire 是一次 admin 故障后未回切的残留。check.sh 已加 `redis_drift`:master≠admin 即告警)。⑥**✅ 第二份独立备份已做(2026-06-19)**: admin `pg_dump`(**独立于 R2**——R2 凭据在 DB 加密,故走自带 dump 非 rclone)→ gzip → ssh 推 **jp `/opt/backup/pg/`**,cron `/opt/sub2api/pg_backup_jp.sh` 每日 **20:00 UTC(4am 北京,+2h 错开 app 的 R2 备份)**,留 14 天。实测 ~340MB/63s。恢复: `scp jp:/opt/backup/pg/sub2api-*.sql.gz . && zcat 文件 | psql`(PG17,dump 带 `--clean --if-exists`)。消除"R2/CF 账号 = 备份+逃生+入口LB"三合一单点。✅ **新鲜度告警已加**(main `check.sh` 每小时 :17 经 `/root/.ssh/jp_check` 查 jp 最新备份,>26h 或缺失即 `notify backup_jp FAIL`;jp 不可达则跳过避免误报)。
16. **2026-06-19 节点清理 + :9100 加固**: ①**入口机清理**: main 删失效 inkmirage cron(`/opt/inkmirage` 已不存在)、废弃 `/opt/sub2apipay`、遗留 `/opt/sub2api`,`docker system prune -af`,磁盘 **40%→18%**;merchant 同样 prune + 删 `/opt/sub2api`。删除物已 tar 到各机 `/root/cleanup-20260619/`(含旧 `.env`,确认后可删)。main 另删 `~/.local/share/claude` 894M(/root 908M→14M)。②**🔒 node_exporter :9100 公网暴露已封**: 5 个 HA 节点 node_exporter(`--net=host`)裸绑 `*:9100`,公网可拉系统指标(泄露 mesh 拓扑),实测 HTTP 200。iptables 顶插 `mesh(10.88.0.0/24)+lo ACCEPT / --dport 9100 DROP` + 持久化(netfilter-persistent 或 rules.v4);实测公网超时、mesh 抓取仍 200(Prometheus 不受影响)。③**SSH key 已厘清(非僵尸,勿删)**: main `authorized_keys` 那 2 把"来历不明"实为**节点间访问**(journald 近30天:`+6pbs4Ac`=bwh1/bwh2→main 6次、`7slxDw0H`=hostdzire→main 23次)。已把注释重命名为 `bwh1-bwh2-root` / `hostdzire-root`(只改注释、指纹不变),**未删**——删会掐断节点间 ssh/自动化。

17. **2026-06-19 审 admin/bwh2 → 🔴 IPv6 防火墙洞(全队)+ 持久化漏洞 已修**:
    - **真相**: #16② 的 :9100「已封」**只是 v4**。全队 `ip6tables -P INPUT ACCEPT` + **0 规则**,且 main/merchant/admin 都有公网 v6、服务多为 dual-stack(`*:9100`/`[::]:5432`)。等于 v4 那套 DB-ACCESS + :9100 DROP **被 IPv6 整条绕过**。**实测**(main 走公网 v6→admin):`:9100`/`:9121` 吐指标、**`:5432` PG 端口 OPEN**。另:**redis_exporter `:9121` 连 v4 都没封**(admin/hostdzire)。
    - **修复**: ①v4 补封 `:9121`(admin/hostdzire,照 9100 式样 lo+mesh ACCEPT / DROP);②v6 每台加 `-A INPUT -i lo ACCEPT` + 敏感端口 DROP——admin=`5432,6379,9100,9121`、main/merchant=`9100`、hostdzire=`9100,9121`(22/80/443 保持公网;mesh 是 v4 无需放行);③持久化:main/merchant/admin 走 netfilter-persistent(已 enabled,rules.v6 落盘),**hostdzire 原来根本没有开机加载 iptables 的机制**(rules.v4/v6 在但没人 restore,重启全丢)→ 新建并 enable `iptables-restore.service`(oneshot,`--test` 校验通过)。
    - **验证**: 跨节点公网 v6 探 4 个目标**全 blocked**;mesh v4 抓取仍正常;**PG 复制 leader+replica streaming lag=0**(v6 5432 DROP 不影响走 mesh 的流复制)。
    - **清理**: admin `authorized_keys` 有**重复 RSA**(`s0I`,= bwh1/bwh2→admin 活跃)→去重保一份 + 标 `bwh1-bwh2-root`(备份在 `~/.ssh/authorized_keys.bak-20260619`);admin 删 `~/.local/share/claude` 670M(/root→8.5M)+ `docker image prune` 163M;bwh2 删 `~/.claude/downloads` 225M(/root→616K)。
    - **站长已决定(2026-06-19,均保留不动)**: ⓐ bwh2 `*:42422` = **xray**(站长自有公网代理,跑在 quorum 投票节点上;注意 0.5G 内存与 etcd/sentinel 共存,被刷可能 OOM 影响仲裁——**保留不动**)。`/root/microsocks` 是没在跑的死目录(留着)。ⓑ admin `/opt/sub` 474M(已停 sub-* 临时栈)**先留着**(磁盘 17% 不缺),可随时 `docker compose start` 恢复或日后 down 销毁。

18. **2026-06-19 全队服务安全审计(多智能体)+ 内部卫生加固**:
    - **方法/教训**: 先多智能体并行审 7 节点 + 从**真·集群外跳板(jp v4 / main 真公网 v6 出口)**实测公网可达性 + 对抗复验。**第一轮 5 个节点 recon 因 API 瞬断失败,导致误判 "clean";补审后翻盘**——结论:外部探测成立要看覆盖,recon 失败必须重跑,别拿残缺结果下结论。
    - **✅ 确认干净**: 公网攻击面 v4+v6 全闭合(16-17 敏感端口从集群外全 blocked,仅 22/80/443、bwh2+42422 开放);PG/Redis **数据面**认证完好(scram/requirepass、绑 mesh/lo);Grafana 关匿名+默认口令被拒、Prometheus/Loki 无宿主端口;无挖矿/反弹 shell/可疑容器;authorized_keys 授权侧全干净无陌生 key;bwh2 xray=VLESS+REALITY 非开放中继。**误报澄清**: canary 的 `SUB2APIPAY_DATABASE_URL` 是未使用的遗留变量,并未直连生产计费库。
    - **🔴 发现并已修(低风险项,本轮全做)**: ①**SSH 口令面**: hostdzire(**critical**,公网 root+口令可爆破、无 fail2ban)/jp/sg/bwh2/admin 此前 `PasswordAuthentication`/`KbdInteractive` 未真正关 → 全队改 `PasswordAuthentication no` + `KbdInteractiveAuthentication no` + `ChallengeResponseAuthentication no` + `PermitRootLogin prohibit-password` + `X11Forwarding no`(统一用 sort 最前的 `00-hardening.conf` 压过 cloud-init 的 `50-*.conf`;`sshd -t`+`sshd -T` 双校验、reload、新连接验证均通,纯 key 登录不受影响)。②**密钥/备份 world-readable**: admin `/opt/sub2api/.env`+`/opt/inkmirage/.env`、hostdzire 三份 `.env`+`haproxy.cfg`+预部署 PG dump、jp 整库 dump → 全 `chmod 600`(haproxy 640;jp `/opt/backup/pg` 700);`pg_backup_jp.sh` 落盘加 `umask 077`+`chmod 600` 防回归。③**主机防火墙**: jp/sg 此前主机层全裸(仅靠云安全组)→ 装最小 nft(default drop input,放行 lo/established/icmp/22)+ `systemctl enable nftables`;jp/sg 的 fail2ban 原 iptables 后端在本机无效 → 改 `nftables-multiport`。④**控制面 mesh 收口**: 5 节点 etcd/sentinel `2379/2380/26379` 加 iptables(v4+v6)`lo+mesh ACCEPT / DROP`(原仅靠 bind 隔离);**bwh2 原无开机加载机制** → 新建 enable `iptables-restore.service`。验证:Patroni leader+replica lag=0、etcd health=true、sentinel PONG、备份链路 admin→jp/main→jp 均通。
    - **⏳ 已记待办(需规划/会动 app,未在本轮做)**: ①**密钥轮换**(站长定:先 chmod 止血、稍后安排)——admin/hostdzire 长期 world-readable 的 DB/Redis/JWT/TOTP/ADMIN 口令、R2/BACKUP_R2、MASTER_KEY_V1、各第三方 API key 应轮换。②**etcd RBAC + client/peer TLS + 各 sentinel `requirepass`**(集群级,mesh 内目前免认证可读写 DCS/下发 failover;公网不可达)。③**jp 备份/巡检 key 加 `command=`/`from=`/`restrict`**(现为公网可达的全 shell root;需改造成受限 forced-command,避免破坏现有备份链路才动)。④**admin pg_hba 收敛到 mesh**(现含公网 IP 的 scram 行,被防火墙遮蔽)。⑤main/merchant `/root/cleanup-20260619/` 旧密钥 tar(/root 0700 实际仅 root 可达,确认后 rm)。

19. **2026-06-19 异地备份扩到 3 份 + 备份 key 收敛(关 §7#18待办③)** —— ⚠️ **本条已被 §7#21 取代**(jp/sg 于 2026-08-18 被甲骨文回收,现为 R2+spartan 两份;下文保留作历史):
    - **第三份独立备份(sg)**: 站长决定"jp 既然备份了,sg 也来一套"。现 **R2(每日,app→对象存储)+ jp + sg(各每日 pg_dump)** 三份独立;jp 日本、sg 新加坡,两台都是甲骨文可回收免费机 → 任一被回收也不丢。两台 197/199G 盘充裕,保留 14 天。
    - **🔒 备份/巡检 key 全部 forced-command 化(closed §7#18待办③)**: jp + sg 的收端 key 不再是"公网可达全 shell root",改为 `command="/usr/local/bin/pg-backup-recv.sh",restrict,from="45.59.186.84"`(收备份,只读 stdin 落盘+清理)与 `command=".../pg-backup-stat.sh",restrict,from="64.186.225.230"`(巡检,只回最新份的 mtime/size)。**实测传 `rm -rf`/`cat /etc/shadow` 当参数均被忽略、只跑固定脚本**。站长本人的 apple key 仍全权(兜底)。
    - **脚本/巡检**: admin `pg_backup_jp.sh` → **`pg_backup_offsite.sh`**(dump 一次到 `/opt/backup/spool` 600 → 推 jp+sg → trap 清 spool;cron 每日 20:00 UTC),实测各推 360MB、gzip 完整、落盘 600/目录 700。main `check.sh` 改为解析 forced-command 的 stat 输出、**同时监 jp+sg 新鲜度**(>26h 各自告警 `backup_jp`/`backup_sg`,不可达跳过)。
    - **仍待办**: §7#18 的 ①密钥轮换 ②etcd RBAC+TLS+sentinel requirepass ④pg_hba 收敛 ⑤cleanup tar 清理 — 未动。

20. **2026-06-19 低峰窗口:pg_hba 收敛(关 §7#18待办④)+ ②评估后判定不做 + cleanup tar 清理(⑤)**:
    - **④ pg_hba 收敛(done)**: admin Patroni DCS 的 pg_hba 去掉 6 条公网 IP 授权行(`45.62.114.56`/`64.186.225.230`/`64.186.226.91`/`23.106.157.112` 的 `host all all`,及 main/merchant 公网的两条**弱 md5** `host all sub2api`),保留 localhost / docker 桥(172.18/172.21,app 用)/ **mesh 10.88.0.0/24** / 复制。手法 `patronictl edit-config --apply - --force`(回滚基线 `admin:/root/patroni-config.bak-20260619.yml`)。验证:leader+replica streaming lag=0、pg_hba.conf 公网行清零、mesh `psql select` ok、app healthy;**PG 未重启**(postmaster 自 5-28,journal 仅 `Reloading PostgreSQL configuration` = reload)。这些公网行本就被 DB-ACCESS 防火墙遮蔽,收敛是"防火墙失效时的兜底 + 去弱 md5"。
    - **② etcd/sentinel 鉴权:评估后【决定不做】**(非偷懒,拓扑使然): ① **TLS 冗余**——mesh 是 WireGuard,传输已加密;② **etcd RBAC 价值≈0**——全 5 节点 main/merchant/admin/hostdzire/bwh2 **本身都是 etcd 成员**,一台被攻陷=一个 etcd 成员被攻陷(本地有 DCS 数据 + 参与 raft),RBAC 只挡"非成员"客户端,挡不住成员;③ **sentinel requirepass 价值低**——app/HAProxy 都走 `HAProxy:6380`(`tcp-check AUTH` 找主)**不连 sentinel**,且被攻陷节点多半本身就是 sentinel 宿主。结论:在"WG 加密 + 全节点皆成员 + 已全面加固 + 端口 mesh 收口"的前提下,这三项是安全表演且要在活集群冒险,**不划算 → 接受当前姿态**。威胁模型变化(如往 mesh 加入不可信节点)再议。
    - **⑤ cleanup tar**: merchant `/root/cleanup-20260619` 已删;main 删配置 tar + .bak,**保留 `sub2apipay.tgz`**(含 4 月订单导出 xlsx,业务数据,待站长定夺删否)。
    - **唯一仍待办**: ①**密钥轮换**(站长定稍后;因 admin/hostdzire 一批密钥曾长期 world-readable)。

21. **2026-08-19 jp/sg 被甲骨文回收 → 异地备份改推 spartan + 修掉「不可达即静默」的漏报**:
    - **事件**: jp(131.186.35.70)/sg(138.2.85.162) 两台甲骨文免费机被回收。最后一次成功推送 `2026-08-18 20:01/20:03 UTC`;之后从 admin 侧实测两台 **22 端口 timeout**。**机器上的两份 dump 随机器一起没了** → 异地独立备份从 3 份掉到 **1 份(只剩 R2)**。当时"随时可能被回收"是已知风险(节点清单里早写着"50Mbps + 随时被回收,只能当冷备"),这次兑现了。
    - **🔴 关键教训 —— 备份没了却一条告警都没有**: main `check.sh` 的新鲜度巡检写成 `mt=$(...); if [ -n "$mt" ]; then notify ...; fi`,**主机不可达 → mt 空 → 整段跳过,不发任何通知**,state 文件永远停在最后一次的 `OK`。当初 §7#19 特意设计"不可达则跳过(避免误报)",在"机器永久消失"这个场景下正好是反的。**通用结论: 监控里「取不到数据」必须能告警,不能等同于「没问题」;抑制瞬时抖动要用重试/连续N次,不能用沉默。**
    - **✅ 已修(check.sh)**: 巡检目标换成 spartan;取不到 mtime 时走 `notify backup_spartan FAIL "巡检失联..."`;瞬时抖动改用**重试一次(间隔 20s)**过滤。备份文件、主机、收端脚本三种失败都会报。旧 `state/backup_jp`、`backup_sg` 已删。备份前基线 `main:/opt/ha-monitor/check.sh.bak-20260819`。
    - **✅ 新收端(spartan)**: `spartanhost`(172.83.153.98,SpartanHost,与 DMIT 不同服务商 → 故障域仍隔离;代价是同在美西,地理上不如日/新分散)。落盘 `/mnt/backup/sub2api-pg`(700,独立 232G `pgbackup` 分区,与 inkmirage 自己的 `/mnt/backup/postgres` 分开),留 14 天。**spartan 不入 mesh、不属 HA 集群**,只作收端。
    - **收端 key 仍是 forced-command**(沿用 §7#19 手法): `command="/usr/local/bin/pg-backup-recv.sh",restrict,from="45.59.186.84"`(admin 推)+ `command="/usr/local/bin/pg-backup-stat.sh",restrict,from="64.186.225.230"`(main 巡检)。**实测带 `cat /etc/shadow` 当参数被忽略、只跑固定脚本**。
    - **收端比 jp/sg 多一道校验**: 落盘后跑 `gzip -t`,大小 <1000B 或 gzip 损坏则**丢弃并回 RECV_FAIL**——堵住"传输截断也照样存下、巡检还报 OK"这个盲区。
    - **✅ 端到端实测**: admin 手跑 `pg_backup_offsite.sh` → `RECV_OK … size=857023689`、`rc=0`(全程约 2.5min,含 dump+gzip+传输+校验);main 巡检读回 `mtime/size/name` 一致;失败路径用不可达 IP 演练确认会 `notify FAIL`。
    - **⏳ 遗留**: ① **备份仍只有 2 份**,想回到 3 份需再找一台异地机(地理分散优先,别再用"随时回收"的免费机当唯一异地)。② **恢复演练仍未做**(bk-4 老待办,现在更值得做:全靠 R2+spartan 两条没验证过的恢复路径)。③ admin 上 `jp_backup`/`sg_backup`、main 上 `jp_check`/`sg_check` 四把废 key 未删(无害,目标已不存在)。

---

## §8 当前故障自愈能力 (实测验证)

| 故障 | 自愈 | RTO |
|---|---|---|
| hostdzire 挂 | ✅ Caddy 切 standby（⚠️**容量降级**: 兜底 admin 仅 3G vs 生产 16G,峰值需手工扩容,见 §4.1） | 5-10s |
| admin 挂 | ✅ patroni+sentinel+HAProxy | ~30s |
| bwh2 (仲裁) 挂 | ✅ 无感 (etcd/sentinel 仍 4/5 过半) | 0 |
| main 挂 | ❌ mayi.one 入口失效 (待阶段4) | 人工 |
| merchant 挂 | ❌ dsrrr.com 入口失效 (待阶段4) | 人工 |
| main+merchant 同挂 | ❌ 两入口全失 (CN2同机房,待阶段4) | 人工 |
| admin+hostdzire 同挂 | ❌ 需 app→R2 恢复 | 小时级 |

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
- → Prometheus (**2026-06-22 起在 spartan**,ssh `spartan` / mesh 10.88.0.5,`/opt/inkmirage/docker/prometheus/prometheus.yml` 的 ha-* jobs;ha-node 含 bwh2。原 hostdzire 的 inkmirage 容器已随 inkmirage 整体迁走;spartan 在 mesh 内抓取,实测 10/10 目标 up,且修复了旧机抓自身 :9100 的 self-loop)
- 改 prometheus 配置后(**在 spartan**): `docker exec inkmirage-prometheus-1 promtool check config /etc/prometheus/prometheus.yml` 校验,再 `docker restart inkmirage-prometheus-1` (没开 hot reload)

### 告警 (TG + 邮件,状态变化才发)
- **主监控**: dmit-main `/opt/ha-monitor/check.sh` (cron 每分钟)。14 项: PG有主/patroni从数/mesh可达(含bwh2)/mayi+dsrrr业务/redis主/磁盘(含bwh2)
- **bwh2 哨兵** (2026-05-29 新增): bwh2 `/opt/ha-monitor/check.sh` (cron 每分钟,TG/邮件带 `[bwh2哨兵]` 前缀)。独立服务商,专盯 main存活(公网+mesh)/main源站mayi/mayi端到端/PG有主/Redis主 —— **补上"main 挂则主监控随之失效"的盲区**
- ~~solid 灾备自检~~ 已随 solid 释放下线(2026-05-29)
- 凭据: `/opt/ha-monitor/secrets` (TG_TOKEN/TG_CHAT/ALERT_EMAIL/REDIS_PW,chmod 600;bwh2 同款 + `/etc/msmtprc`)
- 状态文件: `/opt/ha-monitor/state/*` (每项一个,去重防刷屏)
- **时区 (2026-06-09)**: 告警时间戳统一**北京时间**(两处 check.sh 第6行 `NOW` 用 `TZ='Asia/Shanghai' date`)。系统时区: main/admin/hostdzire = **UTC**(跑 PG/etcd + `0 2` 定时,刻意不动,避免平移备份/清理);bwh2 = **Asia/Shanghai**(纯投票+监控,已 `timedatectl set-timezone` 切,原为洛杉矶 PDT)。→ 排障看 `docker inspect`/`journalctl` 日志时记得 main/admin/hostdzire 是 **UTC(+8=北京)**

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
> **三处 sub2api**:① 测试栈 = hostdzire `/opt/sub2api-test/`(canary dev + **独立 test 库** sub2api-testdb + testredis,绑 `10.88.0.4:8081`)② 生产主 = hostdzire `/opt/sub2api/`(:8080)③ 生产兜底 = admin `/opt/sub2api/` 的 `sub2api-backup`(`10.88.0.3:8080`)。
> **域名分工(2026-05-30)**:`api.dsrrr.com` DNS→**admin(45.59.186.84)**,admin Caddy 反代本机兜底 `10.88.0.3:8080`→**生产库**(真实管理/预发,改东西生效);`canary.dsrrr.com` DNS→**hostdzire(23.80.82.115)**,hostdzire Caddy 反代 canary `10.88.0.4:8081`→**隔离 test 库**(测新版本,migration 安全)。两者都 Let's Encrypt 自动签。
> ✅ **canary 连的是隔离 test 库,migration 只动 test、绝不碰生产**(已验证 0 连接到生产)。admin 不再跑测试码。

> **锁定不可变 tag 发布**(根治浮动 `:dev` 顺序部署的版本错位)。CI 每次构建同时推 `:dev` + `:dev-<12位sha>`(dev-build.yml);compose 用 `image: ...:${SUB2API_TAG:-dev}`,`update.sh <tag>` 会把 tag 写进 `.env` 锁定。三处发**同一个 tag** → 镜像必然一致(已验证)。
1. `git push origin dev` → CI build `:dev` + `:dev-<sha12>`。**等 CI 完成**。
2. 取本次构建 tag:`TAG=dev-$(git rev-parse origin/dev | cut -c1-12)`(与 CI 同算法,可复现)
3. **测试**:`ssh hostdzire "cd /opt/sub2api-test && ./update.sh $TAG"` → 在 **canary.dsrrr.com** 验证(对着隔离 test 库)
4. **生产主**:`ssh hostdzire "cd /opt/sub2api && ./update.sh $TAG"`(此时才在生产库跑 migration)
5. **生产兜底**:`ssh dmit-admin "cd /opt/sub2api && ./update.sh $TAG"`
> 🔙 **回滚** = 三处再发旧 tag:`./update.sh dev-<旧sha12>`(registry 里历史 tag 都在;2026-05-29 前的老 tag 是 7 位)。
> 不传 tag 的 `./update.sh` 仍可用(用 `.env` 现值,默认浮动 `:dev`),但**正式发布务必带 tag**避免错位。

🔄 **重置 test 库数据**(可选,想用更新的真实数据测时): 重跑 §14 恢复演练(从 **R2 下载**或 **spartan `/mnt/backup/sub2api-pg` 取最新 dump** → restore 进 sub2api-testdb;原文的 solid 早已释放)。test 库是某次快照,会随时间与生产漂移。
🔧 回滚 admin 兜底拆分: `ssh dmit-admin "cd /opt/sub2api && cp docker-compose.yml.bak.split docker-compose.yml && docker rm -f sub2api-backup && docker compose up -d"`(退回单容器)。

---

## §13 一/两节点失效矩阵 (实际服务商拓扑, 2026-05-29)

> **前置假设**(站长确认,见 §1 服务商故障域):
> 1. 跨服务商「同时」宕忽略不计 → 只看: 单台任意挂 + DMIT 内部双挂(main+merchant / main+admin / merchant+admin)。`main+hostdzire`、`admin+hostdzire` 跨服务商,忽略。
> 2. 商户按标准接入(子域 CNAME mayi.one,见 §12)→ merchant 挂时商户子域随 CF LB 自愈,仅 `dsrrr.com` 本站 + apex + 旧A记录商户 留 merchant 单点。

### 单台宕机
| 宕机 | 影响 & 自愈 | RTO | <1min |
|---|---|---|---|
| bwh2(仲裁) | 第5票丢失,etcd/sentinel 仍 4/5 过半,无感 | 0s | ✅ |
| hostdzire | Caddy 切 admin standby（⚠️容量降级 16G→3G,见 §4.1/§8）;PG/Redis 主在 admin 不动;mayi 默认源站 main 仍在 | 5–10s | ✅ |
| admin | patroni 提 hostdzire 为主 + sentinel 切 Redis + HAProxy 跟随 | ~30s | ✅ |
| main | mayi.one + 挂它的商户子域靠 CF LB 切 hostdzire;**监控 check.sh 随 main 失效**(bwh2 哨兵独立兜底,见 §10) | ~60s+TTL | ⚠️ 卡线/偏超 |
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
- **check.sh 是告警发送方本身(必要,非冗余)**: §10 cron 每分钟跑 14 项 → TG+邮件,发告警的就是它。唯一盲区是「main 自己挂 → 告警器随之挂」,而 §10 已记录该场景由 CF LB/DNS 切换暴露 + **bwh2 哨兵**(独立服务商,见 §10)独立兜底,站长已接受。无新待办。

### 一句话结论 (按站长决定收敛后,最终)
**应用层 + 数据层: 所有现实的一/两节点故障都在 ~30s 内自愈** —— 单台 admin/hostdzire,及 DMIT 内部双挂(main+merchant / main+admin / merchant+admin),app 始终有 hostdzire 主或 admin standby 兜底、DB 始终能 patroni 提 hostdzire 主。**唯二不在 1 分钟内自愈的就是两条已接受的入口取舍**(merchant 入口单点 + mayi CN2 尽力而为)。
> ⚠️ **结论范围限定**: 上述"自愈"仅指【一/两节点故障】这一类。以下**已登记的尾部/系统性残余风险不在此结论内**,勿据本句认为"零缺口": ①DMIT 账号级/全局网络级事件可一次打掉 3/5 票→quorum 丢失停写(§1);②CF 账号级事件可同时打掉 mayi 权威 DNS+LB+R2 备份+橙云逃生;③hostdzire failover 后 admin 3G 兜底容量降级(§4.1/§8);④Prometheus 单点、Grafana 面板未做(§10);⑤密钥轮换待办(§7#20)。这些是【已知并(多数)已接受】,非"无缺口"。

---

## §14 备份与异地容灾 (2026-05-29)

> HA(节点挂)≠ 备份(数据坏)。流复制会把误删/逻辑损坏复制到所有从库,只有备份能回滚。本节是数据持久性的底线。
> 📌 **现状(2026-08-19 更新,以 §7#21 为准)**: **两份独立** = ① app→**R2**(每日 `0 2`,UI 一键恢复)② admin pg_dump→**spartan**(每日 20:00 UTC,`/mnt/backup/sub2api-pg`,留 14 天)。
> ⚠️ **jp/sg 两份已于 2026-08-18 随机器被甲骨文回收而消失**(不是停更,是副本没了),§7#19/#20 里"三份独立"的说法**已作废**,见 §7#21。
> spartan 恢复路径 = `scp spartan:/mnt/backup/sub2api-pg/sub2api-*.sql.gz . && zcat 文件 | psql`(PG17,dump 带 `--clean --if-exists`;⚠️尚未端到端演练,见审查 bk-4)。

### 主备份: sub2api 自带 → Cloudflare R2 (✅ 站长早已配好;2026-05-29 确认)
- 后台「系统设置 → 数据备份」: S3/R2(`…r2.cloudflarestorage.com`, bucket `sub2api-backups`, 前缀 `backups/`),**定时 `0 2 * * *`**,保留 **14天 / 3份**。产物 `.sql.gz`(~440MB/天),UI 可**下载 + 一键恢复**(应用 BackupService)。R2 durable —— **这是数据持久性的正主**。RPO=24h。
- **不另上 WAL/PITR**: 坏迁移风险已由「canary 对隔离 test 库测 migration」前置拦截(§12),再加 PITR 属过度工程、与"降复杂度"冲突,站长决定跳过。

### ~~第二份独立副本: solid → alice~~ 已停(solid 于 2026-05-29 释放)
> 站长定: solid 释放时**直接弃**第二副本,只留主备份 **app→R2**(durable + UI 一键恢复)。alice 旧副本目录已删、solid 密钥已撤;jp1/sg1 早先已退役。
> **备份现状 = 仅 app→R2 一条。** 将来若想要独立于 R2/CF 的第二份:给某台 VPS 配 rclone 每天拉 R2(需 R2 只读密钥)即可恢复此能力。

### 恢复(唯一路径 = app→R2)
- 后台「系统设置 → 数据备份」→ 选一条备份 → 一键「恢复」(从 R2)。
- 验证可恢复性: 把某条备份恢复到 hostdzire 的 `sub2api-testdb`(test 库)核对行数/时间(参考下方恢复演练)。

### 恢复演练 ✅ 已做 (2026-05-29)
- **结果 PASS**: 最新 dump(pg_sub2api_20260529_030001,463M)→ hostdzire 上隔离 test PG(`sub2api-testdb` 容器,postgres:17,127.0.0.1:5434)`pg_restore` 成功:exit 0、1m45s、**81 表 / 5.7GB**、usage_logs 102万行、最新记录 `2026-05-28 21:30 UTC`=备份时刻(完整无截断)。恢复期间 hostdzire 生产 sub2api + PG 复制 lag 0,无影响。
- 复跑演练(solid 已释放,改从 R2):后台数据备份页「下载」一条 `.sql.gz` → `gunzip` → 恢复到 hostdzire `sub2api-testdb` 核对行数/最新时间。
- ⚠️ **`ops_system_logs` 单表 4.3GB = 整库 76%**:撑大了 dump(463M)和库(5.7GB)。加保留期可大幅瘦身(备份更快、异地机更省)。**待办**: 给 ops_system_logs(及 usage_logs)定保留策略。
- 这个 `sub2api-testdb`(hostdzire,带真实数据)同时作为 §1 测试环境的隔离库(见下,待接入 canary)。

### RPO 现状(已定,不再追)
- RPO = **24h**(app→R2 每日快照)。PITR **决定跳过**(见上)。坏迁移由 canary 前置拦截;节点/机房挂由 HA ~30s 自愈;异地不追(站长定)。
- ~~solid 异地流复制从~~ 已于 2026-05-29 释放。数据现为 **admin + hostdzire 两副本**(patroni 管理);app→R2 备份不受影响。
