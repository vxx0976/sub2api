# bwh2 接入步骤 — 第5票 etcd/sentinel 投票器 + main 独立哨兵

> ✅ **已执行完成 2026-05-29**(步骤1-7 全部完成并验证)。bwh2=10.88.0.6 公网45.62.114.56;etcd 5成员 {.1,.2,.3,.4,.6} RAFT 一致、leader=main;sentinel num-other-sentinels=4/quorum=3/master=.3;check.sh 哨兵 cron 每分钟、TG+邮件双通道实测通;node_exporter 1.11.1 已纳入 Prometheus(health=up);main check.sh 已含 bwh2。solid 已退出 etcd/sentinel 仲裁,仍作 PG/Redis 数据从(.5)。
> 步骤8(solid 释放)✅ 已于 2026-05-29 执行(集群侧 decommission 完成,见 §8)。

> 目的: 现在就把 solid 的 **etcd + sentinel 仲裁** 角色迁到 bwh2(搬瓦工,独立服务商),保住"任意挂2台仍有 quorum";顺带在 bwh2 跑一份 check.sh 哨兵,补 §10「main 挂则主监控随之失效」的盲区。
> 方案(站长定): bwh2 取 **10.88.0.6** 作第5票;**solid 不动 IP、不动数据角色**(继续在 .5 跑 PG/Redis 从 + 每日备份),只摘掉它的 etcd/sentinel 投票。半年后(约 2026-11)solid 整机释放时再处理它的数据从(步骤8)。
>   为何不让 bwh2 占 .5:那要把 solid 的 patroni(PG从 connect_address)+ redis(announce-ip)+ wg 全改到 .6 并重启——在一台即将释放的机器上重配+重启 PG/Redis 复制,风险和工作量都更大;bwh2 直接放 .6 效果完全一样,solid 一个字节不用动。IP 号只是标签,最终 .5 空着无妨。
> 原则: **先把 bwh2 加成投票成员(learner→promote),再摘 solid 投票**,全程容忍挂2不降级。
> ⚠️ 本文是步骤稿,执行前站长过目。所有 `<...>` 占位需替换;密码/Token 一律从 main `/opt/ha-monitor/secrets`、各节点 `/etc/redis/sentinel.conf` 取,**不要写进本文**。

## 0. 关键参数(实采)

| 节点 | mesh IP | 公网 endpoint | etcd pubkey(WireGuard) |
|---|---|---|---|
| main | 10.88.0.1 | 64.186.225.230 | RN6ohxvwjeTQrTyxmkUNTfw3+th4k34Kqiuki+JqAGk= |
| merchant | 10.88.0.2 | 64.186.226.91 | Vd0aAio+7mDK9soIG6GKLKQEDvUtC7qDlElcbZMIm2M= |
| admin | 10.88.0.3 | 45.59.186.84 | +9uf6r4a94+chvQqk96HkMd8qCEAEWyK70GrE6++nWw= |
| hostdzire | 10.88.0.4 | 23.80.82.115 | Eo7f4ixlVENQUVISsMw7t3Y1hcrpYPPF2PS02C3adXQ= |
| solid | 10.88.0.5 | 154.3.224.215 | RVSms9gp9LuREi1vRE4Uv8dnMzTtWqo/A4dcSDZXigo= |
| **bwh2** | **10.88.0.6** | **45.62.114.56** | (步骤1生成) |

- mesh: wg0, 网段 10.88.0.0/24, 端口 51820/udp
- etcd: v3.5.16, token=`sub2api-ha`, 全 http(无TLS)走 mesh, 端口 2379(client)/2380(peer)
- sentinel: master `mymaster` = 10.88.0.3:6379, quorum=3, down-after=5000, failover-timeout=15000, 端口 26379

**Quorum 推演(etcd/sentinel 投票成员)**: 现 5 票 {.1,.2,.3,.4,solid.5}(q=3,容忍2)→ bwh2 加成 learner 不计票(仍 q=3)→ promote 后 6 票(q=4,容忍2)→ 摘 solid 投票后 5 票 {.1,.2,.3,.4,bwh2.6}(q=3,容忍2)。**全程容忍2,不降级**。
> 注: solid 退出投票后仍在 .5 跑 PG/Redis 数据从(数据仍 3 副本),直到半年后整机释放(步骤8)。

---

## 1. WireGuard 入网

### 1a. bwh2 上
```bash
ssh bwh2
apt-get update && apt-get install -y wireguard-tools
umask 077; wg genkey | tee /etc/wireguard/privkey | wg pubkey | tee /etc/wireguard/pubkey
cat /etc/wireguard/pubkey   # 记下 bwh2 公钥, 步骤1b要用
```
写 `/etc/wireguard/wg0.conf`(PrivateKey 用上面的 privkey):
```ini
[Interface]
Address = 10.88.0.6/24
ListenPort = 51820
PrivateKey = <bwh2 privkey>

[Peer]   # main
PublicKey = RN6ohxvwjeTQrTyxmkUNTfw3+th4k34Kqiuki+JqAGk=
Endpoint = 64.186.225.230:51820
AllowedIPs = 10.88.0.1/32
PersistentKeepalive = 25
[Peer]   # merchant
PublicKey = Vd0aAio+7mDK9soIG6GKLKQEDvUtC7qDlElcbZMIm2M=
Endpoint = 64.186.226.91:51820
AllowedIPs = 10.88.0.2/32
PersistentKeepalive = 25
[Peer]   # admin
PublicKey = +9uf6r4a94+chvQqk96HkMd8qCEAEWyK70GrE6++nWw=
Endpoint = 45.59.186.84:51820
AllowedIPs = 10.88.0.3/32
PersistentKeepalive = 25
[Peer]   # hostdzire
PublicKey = Eo7f4ixlVENQUVISsMw7t3Y1hcrpYPPF2PS02C3adXQ=
Endpoint = 23.80.82.115:51820
AllowedIPs = 10.88.0.4/32
PersistentKeepalive = 25
[Peer]   # solid (释放时删本段)
PublicKey = RVSms9gp9LuREi1vRE4Uv8dnMzTtWqo/A4dcSDZXigo=
Endpoint = 154.3.224.215:51820
AllowedIPs = 10.88.0.5/32
PersistentKeepalive = 25
```
```bash
systemctl enable --now wg-quick@wg0
```
> ⚠️ 确认搬瓦工 KiwiVM 面板/上游没挡 UDP 51820(本机 iptables 默认 ACCEPT,无需改)。

### 1b. 5 台现有节点上(各加 bwh2 为 peer)
在 main / merchant / admin / hostdzire / solid 各自执行(立即生效 + 落盘):
```bash
BWH2_PUB="<bwh2 公钥>"
# 立即生效
wg set wg0 peer "$BWH2_PUB" endpoint 45.62.114.56:51820 allowed-ips 10.88.0.6/32 persistent-keepalive 25
# 落盘持久化:把下面追加到 /etc/wireguard/wg0.conf
cat >> /etc/wireguard/wg0.conf <<EOF

[Peer]   # bwh2
PublicKey = $BWH2_PUB
Endpoint = 45.62.114.56:51820
AllowedIPs = 10.88.0.6/32
PersistentKeepalive = 25
EOF
```

### 1c. 验证
```bash
ssh bwh2 'for i in 1 2 3 4 5; do ping -c1 -W2 10.88.0.$i >/dev/null && echo ".$i ok" || echo ".$i FAIL"; done; wg show'
```

---

## 2. etcd(原生二进制,加为第6个 member,先 learner 后 promote)

### 2a. bwh2 装二进制(对齐 v3.5.16)
```bash
ssh bwh2
cd /tmp
curl -L -o etcd.tgz https://github.com/etcd-io/etcd/releases/download/v3.5.16/etcd-v3.5.16-linux-amd64.tar.gz
tar xzf etcd.tgz
install -m755 etcd-v3.5.16-linux-amd64/etcd etcd-v3.5.16-linux-amd64/etcdctl /usr/local/bin/
mkdir -p /var/lib/etcd && chmod 700 /var/lib/etcd
etcd --version   # 应为 3.5.16
```

### 2b. 把 bwh2 作为 learner 加入(从 bwh2 经 mesh 调 admin 的 etcd)
```bash
# learner 不计入 quorum,加入零风险
etcdctl --endpoints=http://10.88.0.3:2379 member add etcd-bwh2 --learner \
  --peer-urls=http://10.88.0.6:2380
etcdctl --endpoints=http://10.88.0.3:2379 member list   # 记下 etcd-bwh2 的 member ID
```

### 2c. bwh2 起 etcd —— `/etc/systemd/system/etcd.service`
(镜像 solid 的 unit;额外给 512MB 小机加了 compaction + quota 护栏)
```ini
[Unit]
Description=etcd
After=network.target wg-quick@wg0.service
Wants=wg-quick@wg0.service

[Service]
Type=notify
ExecStart=/usr/local/bin/etcd --name etcd-bwh2 --data-dir /var/lib/etcd \
  --listen-client-urls http://10.88.0.6:2379,http://127.0.0.1:2379 \
  --advertise-client-urls http://10.88.0.6:2379 \
  --listen-peer-urls http://10.88.0.6:2380 \
  --initial-advertise-peer-urls http://10.88.0.6:2380 \
  --initial-cluster etcd-main=http://10.88.0.1:2380,etcd-merchant=http://10.88.0.2:2380,etcd-admin=http://10.88.0.3:2380,etcd-hostdzire=http://10.88.0.4:2380,etcd-solid=http://10.88.0.5:2380,etcd-bwh2=http://10.88.0.6:2380 \
  --initial-cluster-state existing --initial-cluster-token sub2api-ha \
  --auto-compaction-mode=periodic --auto-compaction-retention=1h \
  --quota-backend-bytes=536870912
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
```bash
systemctl daemon-reload && systemctl enable --now etcd
journalctl -u etcd -n30 --no-pager
etcdctl --endpoints=http://10.88.0.6:2379 endpoint health   # healthy
```

### 2d. learner 追平后 promote 为正式投票成员
```bash
etcdctl --endpoints=http://10.88.0.3:2379 endpoint status -w table --cluster   # 看 bwh2 RAFT INDEX 追平
etcdctl --endpoints=http://10.88.0.3:2379 member promote <etcd-bwh2 member ID>
etcdctl --endpoints=http://10.88.0.3:2379 member list -w table   # etcd-bwh2 IS LEARNER=false
```
> 现在 6 个投票成员,q=4,容忍2。patroni 无感(它只读 etcd)。

---

## 3. sentinel(仅哨兵,不跑数据 Redis)

```bash
ssh bwh2
apt-get install -y redis-sentinel
systemctl disable --now redis-server 2>/dev/null || true   # 不需要本地数据 Redis,省内存
```
写 `/etc/redis/sentinel.conf`(镜像 solid,IP 改 .6;auth-pass 取自任一节点 sentinel.conf 的 `sentinel auth-pass mymaster`,即 REDIS_PW):
```
port 26379
bind 127.0.0.1 10.88.0.6
protected-mode no
dir "/var/lib/redis"
sentinel announce-ip "10.88.0.6"
sentinel announce-port 26379
sentinel monitor mymaster 10.88.0.3 6379 3
sentinel auth-pass mymaster <REDIS_PW>
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 15000
```
```bash
systemctl enable --now redis-sentinel
# 验证:应返回 10.88.0.3 6379,且 num-other-sentinels 随发现增长
redis-cli -p 26379 sentinel get-master-addr-by-name mymaster
redis-cli -p 26379 sentinel master mymaster | grep -A1 num-other-sentinels
```
> sentinel 会自动把 .6 通知给其余哨兵(它们 known-sentinel 里会出现 10.88.0.6)。

---

## 4. node_exporter(可选但推荐:供 check.sh 磁盘项 + Prometheus 抓取)

> 现有节点 :9100 在跑 node_exporter,但 unit 名未确认。**安装前先在任一节点 `ss -tlnp | grep 9100` / `systemctl list-units | grep -i node` 确认版本与装法,bwh2 对齐即可。** 标准做法:
```bash
# 下载与现网同版本的 node_exporter, 装到 /usr/local/bin, systemd 起在 mesh IP
# ExecStart=/usr/local/bin/node_exporter --web.listen-address=10.88.0.6:9100
```
装好后在 main 的 check.sh 节点列表里加 bwh2(见步骤6),即可纳入磁盘监控。

---

## 5. check.sh 哨兵(bwh2 独立监控 main / 入口 / 集群核心)

### 5a. 依赖 + 密钥
```bash
ssh bwh2
apt-get install -y msmtp msmtp-mta   # 站长已确认要邮件告警,必装
mkdir -p /opt/ha-monitor/state
# 从 main 拷 secrets(含 TG_TOKEN/TG_CHAT/ALERT_EMAIL/REDIS_PW),chmod 600
#   scp 不便就手动 cat main 的 /opt/ha-monitor/secrets 后粘贴
chmod 600 /opt/ha-monitor/secrets
# 邮件必做:把 main 的 msmtp 配置(/etc/msmtprc,含 SMTP 账号/密码)一并拷到 bwh2 /etc/msmtprc, chmod 600
#   自测: echo "test" | msmtp <ALERT_EMAIL>   收到即 OK
```

### 5b. `/opt/ha-monitor/check.sh`(chmod 700)
聚焦 main 挂时主监控覆盖不到的盲区;TG/邮件带 `[bwh2哨兵]` 前缀以区分来源。
```bash
#!/bin/bash
# sub2api HA 哨兵 - bwh2 (独立服务商,补 main 挂时主监控失效的盲区). cron 每分钟, 状态变化才告警.
set -uo pipefail
. /opt/ha-monitor/secrets
STATE=/opt/ha-monitor/state; mkdir -p "$STATE"
NOW=$(date '+%m-%d %H:%M'); TAG="bwh2哨兵"
notify() {
  local name="$1" status="$2" msg="$3" sf="$STATE/$1" prev
  prev=$(cat "$sf" 2>/dev/null || echo INIT); echo "$status" > "$sf"
  [ "$status" = "$prev" ] && return
  [ "$prev" = INIT ] && [ "$status" = OK ] && return
  local icon; [ "$status" = OK ] && icon="✅恢复" || icon="🔴故障"
  local text="[$TAG] $icon | $name | $msg | $NOW"
  curl -s "https://api.telegram.org/bot$TG_TOKEN/sendMessage" -d chat_id="$TG_CHAT" --data-urlencode text="$text" >/dev/null 2>&1
  command -v msmtp >/dev/null && printf "Subject: [sub2api HA %s] %s %s\nFrom: noreply@mayi.one\nTo: %s\n\n%s\n" "$TAG" "$icon" "$name" "$ALERT_EMAIL" "$text" | msmtp "$ALERT_EMAIL" >/dev/null 2>&1
}
# A. main 存活(核心:main 挂→其本机 check.sh 失效,这里独立报)
ping -c3 -W3 64.186.225.230 >/dev/null 2>&1 && notify main_pub OK "main公网可达"   || notify main_pub FAIL "main公网不可达!"
ping -c3 -W3 10.88.0.1      >/dev/null 2>&1 && notify main_mesh OK "main mesh可达" || notify main_mesh FAIL "main mesh不可达"
# B. main 源站在 serving(mayi.one 经 main)
mh=$(curl -s -o /dev/null -w '%{http_code}' https://mayi.one/healthz --resolve mayi.one:443:64.186.225.230 --max-time 8 2>/dev/null)
[ "$mh" = 200 ] && notify main_caddy OK "main源站mayi OK" || notify main_caddy FAIL "main源站mayi HTTP$mh"
# C. mayi.one 端到端(CF LB 正常解析,整体入口是否活)
me=$(curl -s -o /dev/null -w '%{http_code}' https://mayi.one/healthz --max-time 10 2>/dev/null)
[ "$me" = 200 ] && notify biz_mayi_e2e OK "mayi.one端到端 OK" || notify biz_mayi_e2e FAIL "mayi.one端到端 HTTP$me"
# D. PG 有主(独立确认)
cl=$(curl -s --max-time 5 http://10.88.0.3:8008/cluster 2>/dev/null); [ -z "$cl" ] && cl=$(curl -s --max-time 5 http://10.88.0.4:8008/cluster 2>/dev/null)
echo "$cl" | grep -q '"role": *"leader"' && notify pg_leader_b OK "PG有主" || notify pg_leader_b FAIL "PG集群无主!"
# E. Redis 主存活
rmaster=$(redis-cli -p 26379 sentinel get-master-addr-by-name mymaster 2>/dev/null | head -1)
if [ -n "$rmaster" ] && redis-cli -h "$rmaster" -p 6379 -a "$REDIS_PW" --no-auth-warning ping 2>/dev/null | grep -q PONG; then
  notify redis_master_b OK "Redis主$rmaster"; else notify redis_master_b FAIL "Redis主不可达($rmaster)"; fi
```
### 5c. cron + 自测
```bash
chmod 700 /opt/ha-monitor/check.sh
( crontab -l 2>/dev/null; echo '* * * * * /opt/ha-monitor/check.sh >/dev/null 2>&1' ) | crontab -
bash /opt/ha-monitor/check.sh        # 首跑只建状态不发(INIT→OK 静默)
echo FAIL > /opt/ha-monitor/state/main_pub && bash /opt/ha-monitor/check.sh  # 应收到 TG「✅恢复 main_pub」验证链路
```

---

## 6. 把 bwh2 纳入 main 主监控 + Prometheus(让别人也看着 bwh2)

- **main 的 `/opt/ha-monitor/check.sh`**:两处 `for n in main:... solid:10.88.0.5` 列表里追加 `bwh2:10.88.0.6`(可达性 + 磁盘项)。
- **hostdzire inkmirage Prometheus**(`/opt/inkmirage/docker/prometheus/prometheus.yml` ha-* job):加 target `10.88.0.6:9100`,然后 `docker restart inkmirage-prometheus-1`。(需先完成步骤4)

---

## 7. 切换:摘除 solid 的 etcd/sentinel 仲裁(现在做)

> ⚠️ 前提: bwh2 的 etcd 已 **promote 成功**(步骤2d)+ sentinel 已上线被发现(步骤3)。确认后再做本步。
> solid 仅退出"投票",**不动它的 PG/Redis 数据从**——solid 继续在 .5 复制 + 备份,直到步骤8 整机释放。

```bash
# 1) etcd:移除 solid 投票成员(6投票 → 5投票 {.1,.2,.3,.4,.6}, q=3, 容忍2)
etcdctl --endpoints=http://10.88.0.3:2379 member list          # 取 etcd-solid 的 ID
etcdctl --endpoints=http://10.88.0.3:2379 member remove <etcd-solid member ID>
ssh solid 'systemctl disable --now etcd'                       # 停 solid 的 etcd
# 2) sentinel:停 solid 哨兵,其余 4 台忘记它
ssh solid 'systemctl disable --now redis-sentinel'
for n in 1 2 3 4 6; do redis-cli -h 10.88.0.$n -p 26379 sentinel reset mymaster; done
```
验证: `etcdctl --endpoints=http://10.88.0.6:2379 member list` 应为 5 成员(无 solid);任一节点 `sentinel master mymaster` 的 num-other-sentinels=4。
> solid 仍在 mesh(.5)、仍是 PG/Redis 从、main 的 check.sh 仍监控它——**不变**。

---

## 8. solid 整机释放 —— ✅ 已执行 (2026-05-29)

> 提前到 2026-05-29 释放(原计划 11月)。备份不再依赖 solid(app→R2 主备份;站长定弃 solid→alice 第二副本)。数据从 3→2(admin+hostdzire patroni 对,已接受)。
> 集群侧 decommission 已全部完成并验证(下面均 ✅),站长可随时在服务商后台终止该 VPS。

实际执行(集群侧):
- ✅ solid: 停 redis-server + postgresql@17-main + 备份 cron(backup_and_mail/offsite_push/solid-check)
- ✅ **复制槽 `solid_standby` 清理(隐蔽坑)**: 它是 patroni 的**永久槽**(DCS config `slots: solid_standby:`),且因历史 failover **admin 和 hostdzire 上都有**。光 drop 会被 patroni 重建。正确做法: ①`curl -XPATCH http://10.88.0.3:8008/config -d '{"slots":{"solid_standby":null}}'` 从配置移除 → ②两节点各 `pg_drop_replication_slot('solid_standby')`。否则 inactive 槽囤 WAL(有 `max_slot_wal_keep_size:50GB` 兜底但仍浪费)。现 admin 只剩 `pg_hostdzire`、hostdzire 只剩 `pg_admin`
- ✅ main check.sh: 两处节点列表删 solid + 清 state(node_solid/disk_solid)
- ✅ 5 节点 sentinel reset(num-slaves 1=hostdzire)
- ✅ 5 节点 WireGuard 删 solid peer(.5)live + wg0.conf(各剩 4 peer)
- ✅ alice: 删第二副本目录 + 撤 solid 密钥;jp1/sg1 早先已退役
- 注: solid 的 etcd/sentinel 投票早在接入 bwh2 时已摘(步骤7),无需再动
- 验证: etcd 5成员 / sentinel quorum3 num-slaves1 / patroni admin-Leader+hostdzire-Replica lag0 / mayi.one 200

---

## 校验清单
- [ ] 步骤6后(摘 solid 前): `etcdctl member list` 6 成员、bwh2 非 learner、`endpoint health` 全 healthy
- [ ] 步骤7后(摘 solid 仲裁后): `etcdctl member list` **5 成员且无 solid**;任一节点 `sentinel master mymaster` 的 num-other-sentinels=**4**
- [ ] solid 仍是 PG/Redis 从: `patronictl list` 含 pg-solid(replica, lag 小)
- [ ] bwh2 `free -h` 余量充足(etcd+sentinel 后 available 仍 >150MB;journald 已限 SystemMaxUse=50M)
- [ ] 邮件链路: bwh2 `echo test | msmtp <ALERT_EMAIL>` 收到
- [ ] 制造一次 main 不可达(防火墙临时挡 / 关 main caddy)→ 收到 bwh2哨兵 TG + 邮件告警
- [ ] HA-RUNBOOK §1 节点表 + §13 假设里 etcd/sentinel 投票成员已更新为 {main,merchant,admin,hostdzire,bwh2}(solid 退出投票)
