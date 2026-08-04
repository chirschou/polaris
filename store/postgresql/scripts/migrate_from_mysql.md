# 从 MySQL 迁移到 PostgreSQL

记录一次真实迁移中踩到的坑。数据搬运本身不难，容易出问题的是**下面三件事**。

## 1. 不要整表 TRUNCATE 后只导业务数据

`polaris_server.sql` 除建表外还带初始化数据，其中鉴权部分是**新版本正常工作的前提**：

| 表 | 初始化内容 |
|---|---|
| `auth_strategy` | 默认策略、全局只读策略、全局读写策略（共 3 条） |
| `auth_strategy_function` | 各策略允许调用的方法，默认策略为 `*` |
| `auth_strategy_resource` | 各策略可操作的资源类型 |
| `auth_principal` | 策略与用户的关联 |

迁移时若为规避主键冲突而 `TRUNCATE` 全部表，这些种子数据会一并被清掉。
而老版本 MySQL 库里**没有**对应记录可供补回（`auth_strategy_function` 等表在老 schema 中根本不存在），
结果是鉴权检查永远匹配不上，控制台所有操作被拒，日志表现为：

```
error auth policy/auth_checker.go  access resource match policy fail
  principal: PrincipalUser/xxx  policy-id: xxx
```

**正确做法**：只清空确实要被业务数据覆盖的表，鉴权初始化数据保留；
或先迁移业务数据，再用 `ON CONFLICT DO NOTHING` 重新执行一遍 schema 中的鉴权 INSERT。

## 2. 手工改鉴权数据后必须更新 auth_strategy.mtime

策略缓存按 `auth_strategy.mtime` 增量拉取。直接往 `auth_strategy_resource` /
`auth_strategy_function` 插行**不会**改变 `auth_strategy.mtime`，缓存因此永远不会重新加载该策略——
数据库里改对了，运行时行为却毫无变化。

polaris 自身的写入路径就带着这一步（`store/*/strategy.go`）：

```go
// 主要是为了能够触发 StrategyCache 的刷新逻辑
updateStrategySql := "UPDATE auth_strategy SET mtime = now() WHERE id = ?"
```

所以任何手工补数据之后都要跟上：

```sql
UPDATE auth_strategy SET mtime = now();
```

注意：缓存的 `lastMtime` 会随轮询持续推进，若更新时刻早于缓存已推进到的位置，
这次 `UPDATE` 同样会被跳过。此时最省事的办法是滚动重启，强制全量重载：

```bash
kubectl rollout restart deploy <polaris-server> -n <ns>
```

重启后日志中应出现 `[Cache][AuthStrategy] get more auth strategy {"add": 3, ...}`。

## 3. 老库 schema 可能落后于代码，缺列要补默认值

生产库未必执行过全部 delta 脚本。本次迁移中源库就缺 `service_contract.type` 等列，
且缺少 `auth_role`、`lane_group` 等整张表。迁移脚本需要能处理：

- MySQL 有、PG 没有的列 → 跳过
- PG 有 NOT NULL 且无默认值、MySQL 没有的列 → 补类型零值
- 列名大小写不一致（如 L5 的 `Fip` 与 `fip`）→ 按小写匹配

`auth_role`、`auth_role_principal`、`auth_strategy_label`、`lane_group`、`lane_rule`
迁移后为空是**正常的**，schema 本就未给它们提供初始化数据。

## 4. 自增序列必须重置

`BIGSERIAL` 列导入既有数据后若不 `setval`，下一次插入立刻主键冲突：

```sql
SELECT setval(pg_get_serial_sequence('config_file','id'),
              COALESCE((SELECT MAX(id) FROM config_file), 0) + 1, false);
```

涉及 6 张表：`config_file`、`config_file_group`、`config_file_release`、
`config_file_release_history`、`config_file_tag`、`config_file_template`。

## 迁移后自检

```sql
-- 鉴权种子数据齐全（缺任一项控制台都会被拒绝访问）
SELECT count(*) FROM auth_strategy;           -- 期望 >= 3
SELECT count(*) FROM auth_strategy_function;  -- 期望 >= 5
SELECT count(*) FROM auth_strategy_resource;  -- 期望 >= 41

-- 序列已重置
SELECT last_value FROM config_file_id_seq;

-- 两侧行数一致（逐表比对）
SELECT count(*) FROM instance;
```

最后更新：2026-08-04
