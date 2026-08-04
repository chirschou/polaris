# PostgreSQL Store 插件

## 简介

Polaris 的 PostgreSQL 存储插件，实现 `store.Store` 接口，与 `store/mysql`、`store/boltdb` 并列，通过配置项 `store.name` 选择。启用后承载注册中心与配置中心的全部持久化数据。

本插件由 `store/mysql` 移植而来，业务逻辑与之保持一致，差异集中在 SQL 方言层。

## 版本要求

**PostgreSQL 11 及以上。**

下限由 schema 中的触发器语法决定：`CREATE TRIGGER ... EXECUTE FUNCTION` 是 PG 11 引入的，
PG 10 只认 `EXECUTE PROCEDURE`（会报 `syntax error at or near "FUNCTION"`）。
其余用到的特性门槛更低——`ON CONFLICT` 需 9.5+，`FOR SHARE` 需 9.3+。

若确有支持 PG 10 及以下的需要，把 `scripts/polaris_server.sql` 中 34 处
`EXECUTE FUNCTION` 改为 `EXECUTE PROCEDURE` 即可，该写法在 PG 11+ 上同样有效。

已在 PostgreSQL 11.22 与 16 上完整验证（schema 导入零错误、全部测试通过）。

## 数据所有权

拥有 `scripts/polaris_server.sql` 中定义的全部 45 张表，包括服务发现（`service`、`instance`、`instance_metadata`、`health_check`）、路由与限流规则、配置中心（`config_file*`）、鉴权（`user`、`auth_strategy*`、`auth_principal`）、泳道与灰度等。与 `store/mysql` 的数据模型完全对应，两者互斥使用。

## 核心功能

实现 `store` 包定义的全部接口，约 200 个方法：

| 接口文件 | 职责 |
|---|---|
| `store/discover_api.go` | 服务、实例、路由、限流、熔断、契约、泳道 |
| `store/config_file_api.go` | 配置文件、发布、历史、模板 |
| `store/auth_api.go` | 用户、用户组、鉴权策略、角色 |
| `store/admin_api.go` | Leader 选举、运维 |

## 上下游依赖

| 方向 | 对象 | 说明 |
|---|---|---|
| 上游调用方 | `service`、`config`、`auth`、`cache` 各层 | 通过 `store.GetStore()` 获取，不直接依赖本包 |
| 下游依赖 | PostgreSQL 实例 | 驱动 `github.com/lib/pq` |

## 配置项

在 `polaris-server.yaml` 中配置：

```yaml
store:
  name: postgresqlStore
  option:
    master:
      dbType: postgres
      dbName: polaris_server
      dbUser: ${POSTGRES_USER}
      dbPwd: ${POSTGRES_PWD}
      dbAddr: 127.0.0.1:5432
      sslMode: disable
      maxOpenConns: 300
      maxIdleConns: 50
      connMaxLifetime: 300
      txIsolationLevel: 2
    # slave 可选，配置后只读查询走从库
```

| 配置项 | 默认值 | 用途 |
|---|---|---|
| `dbType` | 无（必填） | 驱动名，固定为 `postgres` |
| `dbName` | 无（必填） | 数据库名 |
| `dbUser` / `dbPwd` | 无（必填） | 连接账号与密码 |
| `dbAddr` | 无（必填） | `host:port` |
| `sslMode` | `disable` | `disable`/`require`/`verify-ca`/`verify-full` |
| `maxOpenConns` | 不限制 | 最大连接数 |
| `maxIdleConns` | 驱动默认 | 最大空闲连接数 |
| `connMaxLifetime` | `1800` | 连接最大存活秒数 |
| `txIsolationLevel` | 驱动默认 | 对应 `sql.IsolationLevel` |

## 快速开始

```bash
# 1. 建库（需连接到 postgres 库执行）
psql -U postgres -c "CREATE DATABASE polaris_server ENCODING 'UTF8';"

# 2. 导入 schema（含表、索引、注释、触发器与初始化数据）
psql -U postgres -d polaris_server -f store/postgresql/scripts/polaris_server.sql

# 3. 修改 polaris-server.yaml 的 store 段为上述配置后启动
./polaris-server start
```

## 与 MySQL 实现的差异

业务逻辑一致，以下为方言与运行时行为的适配点：

| 项 | MySQL | PostgreSQL |
|---|---|---|
| 占位符 | `?` | `$N`，在 `dialect.go` 的 `convertPlaceholders` 中于执行入口统一改写 |
| bool 参数 | 驱动编码为 `0/1` | `lib/pq` 发送 `true/false`，由 `normalizeArgs` 统一转为 `0/1` |
| 分页 | `LIMIT offset, count` | `OFFSET ? LIMIT ?`，参数顺序不变 |
| `REPLACE INTO` | 删除后重插 | `INSERT ... ON CONFLICT (pk) DO UPDATE` |
| `INSERT IGNORE` | 忽略冲突 | `INSERT ... ON CONFLICT DO NOTHING` |
| `ON DUPLICATE KEY UPDATE` | — | `ON CONFLICT (pk) DO UPDATE SET` |
| `IFNULL` | — | `COALESCE` |
| `SYSDATE()` | 语句实时时间 | `now()`（事务开始时间） |
| `UNIX_TIMESTAMP(x)` | — | `extract(epoch from x)::bigint` |
| `FROM_UNIXTIME(x)` | — | `to_timestamp(x)` |
| `LOCK IN SHARE MODE` | — | `FOR SHARE` |
| `FORCE INDEX` | 索引提示 | 无等价语法，交由优化器；`needForceIndex` 参数不生效 |
| `ON UPDATE CURRENT_TIMESTAMP` | 列属性 | `BEFORE UPDATE` 触发器（`polaris_set_mtime` / `polaris_set_modify_time`） |
| 保留字 | `` `group` `` | `"group"`、`"user"`、`"default"` 用双引号 |
| `VALUE (...)` | 等价于 `VALUES` | 必须写 `VALUES` |
| `UPDATE t AS a SET a.col = ...` | 允许别名前缀 | `SET` 子句不能带别名前缀，须写 `SET col = ...` |
| `DELETE ... LIMIT n` | 支持 | 不支持，改写为 `DELETE ... WHERE id IN (SELECT id ... LIMIT n)` |
| 事务快照 | `START TRANSACTION WITH CONSISTENT SNAPSHOT` | `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ` |

需要注意的行为差异：

- **事务内语句失败**：MySQL 允许继续执行，PostgreSQL 会将整个事务标记为失败，后续语句一律报错，直到回滚。
- **`now()` 语义**：PostgreSQL 的 `now()` 返回事务开始时间，同一事务内多次写入得到相同时间戳；MySQL 的 `SYSDATE()` 为实时时间。若需实时值可改用 `clock_timestamp()`。
- **不提供 delta 升级脚本**：`store/mysql/scripts/delta` 用于历史 MySQL 部署的版本升级，PostgreSQL 为全新支持，仅提供全量 schema。

## 从 MySQL 迁移

见 [scripts/migrate_from_mysql.md](scripts/migrate_from_mysql.md)。

有三个坑值得先读：整表 `TRUNCATE` 会清掉鉴权初始化数据导致控制台全部拒绝访问；
手工改鉴权数据后必须 `UPDATE auth_strategy SET mtime = now()` 否则缓存不刷新；
`BIGSERIAL` 列导入后必须 `setval`，否则下一次插入即主键冲突。

## 测试

单元测试：`go test ./store/postgresql/...`。其中 `dialect_test.go` 覆盖占位符改写（字面量内的 `?`、`''` 转义、批量插入编号）与 bool 参数转换。

SQL 完整性测试（`sqlcoverage_test.go` 查询侧 + `sqlwrite_test.go` 写入侧）直接调用 store 层方法，
覆盖各参数分支拼出的最终 SQL，共 248 个用例，覆盖全部 92 个写方法与查询侧的参数分支。
需真实 PostgreSQL，未设 `POLARIS_PG_ADDR` 时自动跳过：

```bash
POLARIS_PG_ADDR=127.0.0.1:55432 go test ./store/postgresql/ -run TestSQL -v
```

写入侧测试数据统一以 `sqlw-` 前缀命名并在前后清理，可重复运行。涉及行锁的用例
（`FOR UPDATE` / `FOR SHARE`）在各自用例内完成事务生命周期，避免相互等待。

集成验证需真实 PostgreSQL：

```bash
docker run -d --name polaris-pg -e POSTGRES_PASSWORD=verify \
  -e POSTGRES_DB=polaris_server -p 55432:5432 postgres:16-alpine
docker exec -i -e PGPASSWORD=verify polaris-pg psql -U postgres -d polaris_server \
  < store/postgresql/scripts/polaris_server.sql
# 将 polaris-server.yaml 的 store 段指向 127.0.0.1:55432 后启动
```

已验证：namespace / service / instance（含 metadata、health_check）、配置中心（分组→文件→发布→历史）、鉴权（用户、用户组、策略）、限流、熔断、路由 v2、泳道、服务契约；全部 11 处 `ON CONFLICT` 的覆盖语义与幂等性。

最后更新：2026-08-03
