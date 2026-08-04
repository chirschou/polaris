/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package postgresql

// SQL 完整性测试：直接调用 store 层方法，验证各参数分支拼出的 SQL 能在真实 PostgreSQL 上执行。
//
// 关注点是 SQL 本身（语法、标识符、占位符编号、类型推断），而非业务结果——
// 调用方逻辑在三种 store 实现间共享，已由 MySQL 实现验证。
//
// 运行方式（未设置 POLARIS_PG_ADDR 时整体跳过，不影响 go test ./...）：
//
//	docker run -d --name pg -e POSTGRES_PASSWORD=verify -e POSTGRES_DB=polaris_server \
//	    -p 55432:5432 postgres:16-alpine
//	docker exec -i -e PGPASSWORD=verify pg psql -U postgres -d polaris_server \
//	    < store/postgresql/scripts/polaris_server.sql
//	POLARIS_PG_ADDR=127.0.0.1:55432 go test ./store/postgresql/ -run TestSQL -v

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/polarismesh/polaris/common/model"
	authcommon "github.com/polarismesh/polaris/common/model/auth"
	"github.com/polarismesh/polaris/store"
)

// newTestDB 连接测试用 PostgreSQL，未配置地址时跳过。
func newTestDB(t *testing.T) *BaseDB {
	t.Helper()
	addr := os.Getenv("POLARIS_PG_ADDR")
	if addr == "" {
		t.Skip("未设置 POLARIS_PG_ADDR，跳过 SQL 完整性测试")
	}
	user, pwd, name := envOr("POLARIS_PG_USER", "postgres"),
		envOr("POLARIS_PG_PWD", "verify"), envOr("POLARIS_PG_DB", "polaris_server")

	db, err := NewBaseDB(&dbConfig{
		dbType: "postgres", dbUser: user, dbPwd: pwd, dbAddr: addr,
		dbName: name, sslMode: "disable",
	}, nil)
	require.NoError(t, err, "连接 PostgreSQL 失败")
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// runCase 执行一个查询分支，只要求 SQL 能被 PostgreSQL 接受并执行。
// 空结果集是正常的——这里验证的是 SQL 可执行性，不是数据内容。
func runCase(t *testing.T, name string, fn func() error) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		if err := fn(); err != nil {
			// store.Error 会把驱动错误包装为状态码，原始信息仍在 Error() 中
			t.Fatalf("SQL 执行失败: %v", err)
		}
	})
}

var (
	zeroTime = time.Unix(0, 1)
	someTime = time.Now().Add(-time.Hour)
)

func TestSQLNamespace(t *testing.T) {
	db := newTestDB(t)
	s := &namespaceStore{master: db, slave: db}

	runCase(t, "GetNamespaces/空filter", func() error {
		_, _, err := s.GetNamespaces(map[string][]string{}, 0, 10)
		return err
	})
	runCase(t, "GetNamespaces/name精确", func() error {
		_, _, err := s.GetNamespaces(map[string][]string{"name": {"default"}}, 0, 10)
		return err
	})
	runCase(t, "GetNamespaces/name通配", func() error {
		_, _, err := s.GetNamespaces(map[string][]string{"name": {"def*"}}, 0, 10)
		return err
	})
	runCase(t, "GetNamespaces/owner模糊", func() error {
		_, _, err := s.GetNamespaces(map[string][]string{"owner": {"polaris"}}, 0, 10)
		return err
	})
	runCase(t, "GetNamespaces/多key多值", func() error {
		_, _, err := s.GetNamespaces(map[string][]string{
			"name": {"default", "Polaris"}, "owner": {"polaris"},
		}, 0, 10)
		return err
	})
	runCase(t, "GetMoreNamespaces", func() error {
		_, err := s.GetMoreNamespaces(zeroTime)
		return err
	})
	runCase(t, "GetNamespace", func() error {
		_, err := s.GetNamespace("default")
		return err
	})
}

func TestSQLInstance(t *testing.T) {
	db := newTestDB(t)
	s := &instanceStore{master: db, slave: db}

	// filter 的每个分支在 genFilterSQL 中走不同的 where 拼装。
	// key 取自 service.InstanceFilterAttributes 白名单（经 InsFilter2toreAttr 映射后的存储层属性），
	// owner/business/managed 虽然 genFilterSQL 有对应分支，但那是 service 等其它调用方在用，
	// 上层不会传给实例查询——instance 表没有这些列，MySQL 实现同样会报错。
	filters := map[string]map[string]string{
		"空filter":       {},
		"host":          {"host": "1.1.1.1"},
		"host多值":        {"host": "1.1.1.1,2.2.2.2"},
		"id精确":          {"id": "abc"},
		"id通配":          {"id": "ab*"},
		"name通配":        {"name": "svc*"},
		"namespace通配":   {"namespace": "def*"},
		"health_status": {"health_status": "1"},
		"isolate":       {"isolate": "0"},
		"port":          {"port": "8080"},
		"protocol":      {"protocol": "grpc"},
		"version":       {"version": "v1"},
		"weight":        {"weight": "100"},
		"logic_set":     {"logic_set": "set1"},
		"cmdb_region":   {"cmdb_region": "sh"},
		"priority":      {"priority": "0"},
		"组合":            {"host": "1.1.1.1", "name": "svc*", "protocol": "grpc"},
	}
	for name, f := range filters {
		f := f
		runCase(t, "GetExpandInstances/"+name, func() error {
			_, _, err := s.GetExpandInstances(f, nil, 0, 10)
			return err
		})
	}
	runCase(t, "GetExpandInstances/带metaFilter", func() error {
		_, _, err := s.GetExpandInstances(map[string]string{}, map[string]string{"env": "prod"}, 0, 10)
		return err
	})
	runCase(t, "GetExpandInstances/limit0", func() error {
		_, _, err := s.GetExpandInstances(map[string]string{}, nil, 0, 0)
		return err
	})

	// firstUpdate 决定走全量还是增量两条完全不同的 SQL。
	// GetMoreInstances 会对 tx 调用 GetDelegateTx，调用方始终传真实事务，这里同样构造一个。
	withTx := func(fn func(store.Tx) error) error {
		baseTx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() { _ = baseTx.Rollback() }()
		return fn(NewSqlDBTx(baseTx))
	}
	for _, first := range []bool{true, false} {
		for _, needMeta := range []bool{true, false} {
			first, needMeta := first, needMeta
			runCase(t, name2("GetMoreInstances", first, needMeta), func() error {
				return withTx(func(tx store.Tx) error {
					_, err := s.GetMoreInstances(tx, zeroTime, first, needMeta, nil)
					return err
				})
			})
			runCase(t, name2("GetMoreInstances/带serviceID", first, needMeta), func() error {
				return withTx(func(tx store.Tx) error {
					_, err := s.GetMoreInstances(tx, zeroTime, first, needMeta, []string{"sid1", "sid2"})
					return err
				})
			})
		}
	}
	runCase(t, "GetInstancesCount", func() error { _, err := s.GetInstancesCount(); return err })
	runCase(t, "GetInstanceMeta", func() error { _, err := s.GetInstanceMeta("no-such-id"); return err })
	runCase(t, "GetInstance", func() error { _, err := s.GetInstance("no-such-id"); return err })
	runCase(t, "GetInstancesBrief/多id", func() error {
		_, err := s.GetInstancesBrief(map[string]bool{"a": true, "b": true, "c": true})
		return err
	})
	runCase(t, "BatchGetInstanceIsolate", func() error {
		_, err := s.BatchGetInstanceIsolate(map[string]bool{"a": true, "b": true})
		return err
	})
	runCase(t, "GetInstancesMainByService", func() error {
		_, err := s.GetInstancesMainByService("sid", "1.1.1.1")
		return err
	})
}

func name2(prefix string, first, needMeta bool) string {
	return prefix + "/firstUpdate=" + b2s(first) + ",needMeta=" + b2s(needMeta)
}

func b2s(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestSQLService(t *testing.T) {
	db := newTestDB(t)
	s := &serviceStore{master: db, slave: db}

	runCase(t, "GetServices/空", func() error {
		_, _, err := s.GetServices(map[string]string{}, nil, nil, 0, 10)
		return err
	})
	runCase(t, "GetServices/name通配", func() error {
		_, _, err := s.GetServices(map[string]string{"name": "svc*"}, nil, nil, 0, 10)
		return err
	})
	runCase(t, "GetServices/owner", func() error {
		_, _, err := s.GetServices(map[string]string{"owner": "polaris"}, nil, nil, 0, 10)
		return err
	})
	runCase(t, "GetServices/带serviceMeta", func() error {
		_, _, err := s.GetServices(map[string]string{}, map[string]string{"k": "v"}, nil, 0, 10)
		return err
	})
	runCase(t, "GetServices/带instanceFilter", func() error {
		_, _, err := s.GetServices(map[string]string{}, nil, &store.InstanceArgs{Hosts: []string{"1.1.1.1"}}, 0, 10)
		return err
	})
	runCase(t, "GetServiceAliases/空", func() error {
		_, _, err := s.GetServiceAliases(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "GetServiceAliases/alias.owner", func() error {
		_, _, err := s.GetServiceAliases(map[string]string{"alias.owner": "polaris"}, 0, 10)
		return err
	})
	for _, first := range []bool{true, false} {
		for _, needMeta := range []bool{true, false} {
			first, needMeta := first, needMeta
			runCase(t, name2("GetMoreServices", first, needMeta), func() error {
				_, err := s.GetMoreServices(zeroTime, first, false, needMeta)
				return err
			})
		}
	}
	runCase(t, "GetServicesCount", func() error { _, err := s.GetServicesCount(); return err })
	runCase(t, "GetService", func() error { _, err := s.GetService("no-svc", "default"); return err })
	runCase(t, "GetSourceServiceToken", func() error {
		_, err := s.GetSourceServiceToken("no-svc", "default")
		return err
	})
}

func TestSQLAuth(t *testing.T) {
	db := newTestDB(t)
	us := &userStore{master: db, slave: db}
	gs := &groupStore{master: db, slave: db}
	ss := &strategyStore{master: db, slave: db}
	rs := &roleStore{master: db, slave: db}

	runCase(t, "GetUsers/空", func() error { _, _, err := us.GetUsers(map[string]string{}, 0, 10); return err })
	runCase(t, "GetUsers/name模糊", func() error {
		_, _, err := us.GetUsers(map[string]string{"name": "pol*"}, 0, 10)
		return err
	})
	runCase(t, "GetUsers/group_id", func() error {
		_, _, err := us.GetUsers(map[string]string{"group_id": "g1"}, 0, 10)
		return err
	})
	runCase(t, "GetUsers/owner+source", func() error {
		_, _, err := us.GetUsers(map[string]string{"owner": "polaris", "source": "Polaris"}, 0, 10)
		return err
	})
	for _, first := range []bool{true, false} {
		first := first
		runCase(t, "GetUsersForCache/firstUpdate="+b2s(first), func() error {
			_, err := us.GetUsersForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetGroupsForCache/firstUpdate="+b2s(first), func() error {
			_, err := gs.GetGroupsForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetMoreStrategies/firstUpdate="+b2s(first), func() error {
			_, err := ss.GetMoreStrategies(zeroTime, first)
			return err
		})
		runCase(t, "GetMoreRoles/firstUpdate="+b2s(first), func() error {
			_, err := rs.GetMoreRoles(first, zeroTime)
			return err
		})
	}
	runCase(t, "GetGroups/空", func() error { _, _, err := gs.GetGroups(map[string]string{}, 0, 10); return err })
	runCase(t, "GetGroups/name+user_id", func() error {
		_, _, err := gs.GetGroups(map[string]string{"name": "g*", "user_id": "u1"}, 0, 10)
		return err
	})
	runCase(t, "GetUser", func() error { _, err := us.GetUser("no-user"); return err })
	runCase(t, "GetUserByName", func() error { _, err := us.GetUserByName("no-user", "polaris"); return err })
	runCase(t, "GetSubCount", func() error {
		_, err := us.GetSubCount(&authcommon.User{ID: "no-user"})
		return err
	})
	runCase(t, "GetGroup", func() error { _, err := gs.GetGroup("no-group"); return err })
	runCase(t, "GetStrategyDetail", func() error { _, err := ss.GetStrategyDetail("no-strategy"); return err })
}

func TestSQLConfig(t *testing.T) {
	db := newTestDB(t)
	cf := &configFileStore{master: db, slave: db}
	cg := &configFileGroupStore{master: db, slave: db}
	cr := &configFileReleaseStore{master: db, slave: db}
	ch := &configFileReleaseHistoryStore{master: db, slave: db}
	ct := &configFileTemplateStore{master: db, slave: db}

	runCase(t, "QueryConfigFiles/空", func() error {
		_, _, err := cf.QueryConfigFiles(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "QueryConfigFiles/namespace+group", func() error {
		_, _, err := cf.QueryConfigFiles(map[string]string{"namespace": "default", "group": "g1"}, 0, 10)
		return err
	})
	runCase(t, "QueryConfigFiles/name模糊", func() error {
		_, _, err := cf.QueryConfigFiles(map[string]string{"name": "app*"}, 0, 10)
		return err
	})
	runCase(t, "CountConfigFileEachGroup", func() error { _, err := cf.CountConfigFileEachGroup(); return err })
	runCase(t, "GetConfigFile", func() error { _, err := cf.GetConfigFile("default", "g1", "no.yaml"); return err })
	runCase(t, "CountConfigFiles", func() error { _, err := cf.CountConfigFiles("default", "g1"); return err })
	runCase(t, "GetConfigFileGroup", func() error { _, err := cg.GetConfigFileGroup("default", "g1"); return err })
	runCase(t, "GetConfigFileRelease", func() error {
		_, err := cr.GetConfigFileRelease(&model.ConfigFileReleaseKey{
			Namespace: "default", Group: "g1", FileName: "app.yaml", Name: "rel1",
		})
		return err
	})
	runCase(t, "GetConfigFileActiveRelease", func() error {
		_, err := cr.GetConfigFileActiveRelease(&model.ConfigFileKey{
			Namespace: "default", Group: "g1", Name: "app.yaml",
		})
		return err
	})
	runCase(t, "QueryConfigFileReleaseHistories/空", func() error {
		_, _, err := ch.QueryConfigFileReleaseHistories(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "QueryConfigFileReleaseHistories/带条件", func() error {
		_, _, err := ch.QueryConfigFileReleaseHistories(map[string]string{"namespace": "default", "group": "g1", "name": "app.yaml"}, 0, 10)
		return err
	})
	runCase(t, "QueryAllConfigFileTemplates", func() error {
		_, err := ct.QueryAllConfigFileTemplates()
		return err
	})
	for _, first := range []bool{true, false} {
		first := first
		runCase(t, "GetMoreReleaseFile/firstUpdate="+b2s(first), func() error {
			_, err := cr.GetMoreReleaseFile(first, zeroTime)
			return err
		})
		runCase(t, "GetMoreConfigGroup/firstUpdate="+b2s(first), func() error {
			_, err := cg.GetMoreConfigGroup(first, zeroTime)
			return err
		})
	}
}

func TestSQLRules(t *testing.T) {
	db := newTestDB(t)
	rl := &rateLimitStore{master: db, slave: db}
	cb := &circuitBreakerStore{master: db, slave: db}
	fd := &faultDetectRuleStore{master: db, slave: db}
	rt := &routingConfigStore{master: db, slave: db}
	rt2 := &routingConfigStoreV2{master: db, slave: db}
	ln := &laneStore{master: db, slave: db}
	sc := &serviceContractStore{master: db, slave: db}
	gr := &grayStore{master: db, slave: db}

	runCase(t, "GetExtendRateLimits/空", func() error {
		_, _, err := rl.GetExtendRateLimits(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "GetExtendRateLimits/带条件", func() error {
		_, _, err := rl.GetExtendRateLimits(map[string]string{"name": "rl*", "namespace": "default"}, 0, 10)
		return err
	})
	runCase(t, "GetCircuitBreakerRules/空", func() error {
		_, _, err := cb.GetCircuitBreakerRules(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "GetCircuitBreakerRules/带条件", func() error {
		_, _, err := cb.GetCircuitBreakerRules(map[string]string{"name": "cb*", "level": "1"}, 0, 10)
		return err
	})
	runCase(t, "GetFaultDetectRules/空", func() error {
		_, _, err := fd.GetFaultDetectRules(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "GetRoutingConfigs/空", func() error {
		_, _, err := rt.GetRoutingConfigs(map[string]string{}, 0, 10)
		return err
	})
	runCase(t, "GetRoutingConfigV2WithID", func() error {
		_, err := rt2.GetRoutingConfigV2WithID("no-such-rule")
		return err
	})
	// GetLaneGroups 直接把 filter["order_field"] 拼进 ORDER BY 且无默认值兜底，
	// 缺失时会生成 "ORDER BY  offset ..." 语法错误——上层负责填充，这里按真实调用传入
	runCase(t, "GetLaneGroups/默认排序", func() error {
		_, _, err := ln.GetLaneGroups(map[string]string{
			"order_field": "mtime", "order_type": "desc",
		}, 0, 10)
		return err
	})
	runCase(t, "GetLaneGroups/带name条件", func() error {
		_, _, err := ln.GetLaneGroups(map[string]string{
			"name": "lane*", "order_field": "name", "order_type": "asc",
		}, 0, 10)
		return err
	})
	runCase(t, "GetServiceContracts/空", func() error {
		_, _, err := sc.GetServiceContracts(context.Background(), map[string]string{}, 0, 10)
		return err
	})
	for _, first := range []bool{true, false} {
		first := first
		runCase(t, "GetRateLimitsForCache/firstUpdate="+b2s(first), func() error {
			_, err := rl.GetRateLimitsForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetCircuitBreakerRulesForCache/firstUpdate="+b2s(first), func() error {
			_, err := cb.GetCircuitBreakerRulesForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetFaultDetectRulesForCache/firstUpdate="+b2s(first), func() error {
			_, err := fd.GetFaultDetectRulesForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetRoutingConfigsForCache/firstUpdate="+b2s(first), func() error {
			_, err := rt.GetRoutingConfigsForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetRoutingConfigsV2ForCache/firstUpdate="+b2s(first), func() error {
			_, err := rt2.GetRoutingConfigsV2ForCache(zeroTime, first)
			return err
		})
		runCase(t, "GetMoreLaneGroups/firstUpdate="+b2s(first), func() error {
			_, err := ln.GetMoreLaneGroups(zeroTime, first)
			return err
		})
		runCase(t, "GetMoreServiceContracts/firstUpdate="+b2s(first), func() error {
			_, err := sc.GetMoreServiceContracts(first, zeroTime)
			return err
		})
		runCase(t, "GetMoreGrayResouces/firstUpdate="+b2s(first), func() error {
			_, err := gr.GetMoreGrayResouces(first, zeroTime)
			return err
		})
	}
}

func TestSQLClientAndL5(t *testing.T) {
	db := newTestDB(t)
	cs := &clientStore{master: db, slave: db}
	l5 := &l5Store{master: db, slave: db}
	ts := &toolStore{db: db}

	for _, first := range []bool{true, false} {
		first := first
		runCase(t, "GetMoreClients/firstUpdate="+b2s(first), func() error {
			_, err := cs.GetMoreClients(someTime, first)
			return err
		})
	}
	runCase(t, "GetClientStat", func() error { _, err := cs.GetClientStat("no-client"); return err })
	// L5：读路径在默认配置下缓存未启用，这里直接覆盖其 SQL
	runCase(t, "GetMoreL5Routes", func() error { _, err := l5.GetMoreL5Routes(0); return err })
	runCase(t, "GetMoreL5Policies", func() error { _, err := l5.GetMoreL5Policies(0); return err })
	runCase(t, "GetMoreL5Sections", func() error { _, err := l5.GetMoreL5Sections(0); return err })
	runCase(t, "GetMoreL5IPConfigs", func() error { _, err := l5.GetMoreL5IPConfigs(0); return err })
	runCase(t, "GetMoreL5Extend", func() error { _, err := l5.GetMoreL5Extend(zeroTime); return err })
	runCase(t, "GetUnixSecond", func() error { _, err := ts.GetUnixSecond(0); return err })
}

func TestSQLAdmin(t *testing.T) {
	db := newTestDB(t)
	le := &leaderElectionStore{master: db}

	runCase(t, "CreateLeaderElection", func() error { return le.CreateLeaderElection("sql-cov-key") })
	runCase(t, "CreateLeaderElection/重复(ON CONFLICT DO NOTHING)", func() error {
		return le.CreateLeaderElection("sql-cov-key")
	})
	runCase(t, "GetVersion", func() error { _, err := le.GetVersion("sql-cov-key"); return err })
	runCase(t, "CompareAndSwapVersion", func() error {
		_, err := le.CompareAndSwapVersion("sql-cov-key", 0, 1, "host-1")
		return err
	})
	runCase(t, "CheckMtimeExpired", func() error {
		_, _, err := le.CheckMtimeExpired("sql-cov-key", 10)
		return err
	})
	runCase(t, "ListLeaderElections", func() error { _, err := le.ListLeaderElections(); return err })
}

// BatchOperation 会按批次切分 SQL，需覆盖单条、多条与跨批次三种规模
func TestSQLBatchSizes(t *testing.T) {
	db := newTestDB(t)
	s := &instanceStore{master: db, slave: db}

	for _, n := range []int{1, 2, 50, 101, 300} {
		n := n
		ids := make(map[string]bool, n)
		for i := 0; i < n; i++ {
			ids[randID(i)] = true
		}
		runCase(t, "GetInstancesBrief/规模="+itoa(n), func() error {
			_, err := s.GetInstancesBrief(ids)
			return err
		})
	}
}

func randID(i int) string { return "id-" + itoa(i) }

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

var _ = store.Error
