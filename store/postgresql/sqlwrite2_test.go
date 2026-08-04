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

// 写路径 SQL 测试（续）：治理规则、泳道、契约、灰度、角色、模板、事务锁与 L5。
// 与 sqlwrite_test.go 共用 sqlw- 前缀与清理逻辑。

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/polarismesh/polaris/common/model"
	authcommon "github.com/polarismesh/polaris/common/model/auth"
	"github.com/polarismesh/polaris/store"
)

// txRunner 提供一个提交语义的事务包装，供需要 store.Tx 的写方法使用
func txRunner(db *BaseDB) func(func(store.Tx) error) error {
	return func(fn func(store.Tx) error) error {
		baseTx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() { _ = baseTx.Rollback() }()
		if err := fn(NewSqlDBTx(baseTx)); err != nil {
			return err
		}
		return baseTx.Commit()
	}
}

func TestSQLWriteRules(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	ss := &serviceStore{master: db, slave: db}
	rl := &rateLimitStore{master: db, slave: db}
	cb := &circuitBreakerStore{master: db, slave: db}
	fd := &faultDetectRuleStore{master: db, slave: db}
	rt := &routingConfigStore{master: db, slave: db}
	rt2 := &routingConfigStoreV2{master: db, slave: db}

	nsName := wPrefix + "ns"
	svcID := wid("svc", 1)
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))
	require.NoError(t, ss.AddService(&model.Service{
		ID: svcID, Name: wPrefix + "svc", Namespace: nsName, Token: "tk", Owner: "polaris", Revision: "r",
	}))

	// 限流规则
	rlID := wid("ratelimit", 1)
	runCase(t, "CreateRateLimit", func() error {
		return rl.CreateRateLimit(newRateLimit(rlID, svcID, "rev-1"))
	})
	runCase(t, "UpdateRateLimit", func() error {
		l := newRateLimit(rlID, svcID, "rev-2")
		l.Priority = 5
		l.Disable = true
		return rl.UpdateRateLimit(l)
	})
	runCase(t, "DeleteRateLimit", func() error {
		return rl.DeleteRateLimit(newRateLimit(rlID, svcID, "rev-3"))
	})

	// 熔断规则
	cbID := wid("cbrule", 1)
	runCase(t, "CreateCircuitBreakerRule", func() error {
		return cb.CreateCircuitBreakerRule(newCBRule(cbID, nsName, "rev-1"))
	})
	runCase(t, "UpdateCircuitBreakerRule", func() error {
		r := newCBRule(cbID, nsName, "rev-2")
		r.Description = "updated"
		r.Enable = true
		return cb.UpdateCircuitBreakerRule(r)
	})
	runCase(t, "DeleteCircuitBreakerRule", func() error { return cb.DeleteCircuitBreakerRule(cbID) })

	// 探测规则
	fdID := wid("fdrule", 1)
	runCase(t, "CreateFaultDetectRule", func() error {
		return fd.CreateFaultDetectRule(newFDRule(fdID, nsName, "rev-1"))
	})
	runCase(t, "UpdateFaultDetectRule", func() error {
		return fd.UpdateFaultDetectRule(newFDRule(fdID, nsName, "rev-2"))
	})
	runCase(t, "DeleteFaultDetectRule", func() error { return fd.DeleteFaultDetectRule(fdID) })

	// 路由规则 v1：以 serviceID 为主键
	runCase(t, "CreateRoutingConfig", func() error {
		return rt.CreateRoutingConfig(&model.RoutingConfig{
			ID: svcID, InBounds: "[]", OutBounds: "[]", Revision: "rrev-1",
		})
	})
	runCase(t, "UpdateRoutingConfig", func() error {
		return rt.UpdateRoutingConfig(&model.RoutingConfig{
			ID: svcID, InBounds: `[{"k":"v"}]`, OutBounds: "[]", Revision: "rrev-2",
		})
	})
	runCase(t, "DeleteRoutingConfig", func() error { return rt.DeleteRoutingConfig(svcID) })

	// 路由规则 v2
	rt2ID := wid("routev2", 1)
	runCase(t, "CreateRoutingConfigV2", func() error {
		return rt2.CreateRoutingConfigV2(newRouterConfig(rt2ID, nsName, "v2rev-1"))
	})
	runCase(t, "UpdateRoutingConfigV2", func() error {
		c := newRouterConfig(rt2ID, nsName, "v2rev-2")
		c.Enable = false
		c.Priority = 3
		return rt2.UpdateRoutingConfigV2(c)
	})
	runCase(t, "CreateRoutingConfigV2Tx", func() error {
		return txRunner(db)(func(tx store.Tx) error {
			return rt2.CreateRoutingConfigV2Tx(tx, newRouterConfig(wid("routev2", 2), nsName, "v2rev-3"))
		})
	})
	runCase(t, "UpdateRoutingConfigV2Tx", func() error {
		return txRunner(db)(func(tx store.Tx) error {
			return rt2.UpdateRoutingConfigV2Tx(tx, newRouterConfig(wid("routev2", 2), nsName, "v2rev-4"))
		})
	})
	runCase(t, "DeleteRoutingConfigV2", func() error { return rt2.DeleteRoutingConfigV2(rt2ID) })
}

func newRateLimit(id, svcID, revision string) *model.RateLimit {
	return &model.RateLimit{
		ID: id, Name: wPrefix + "rl", ServiceID: svcID, Method: "/api",
		Labels: "{}", Priority: 0, Rule: "{}", Revision: revision, Disable: false,
	}
}

func newCBRule(id, ns, revision string) *model.CircuitBreakerRule {
	return &model.CircuitBreakerRule{
		ID: id, Name: wPrefix + "cb", Namespace: ns, Description: "d", Level: 1,
		SrcService: "*", SrcNamespace: "*", DstService: "*", DstNamespace: "*", DstMethod: "*",
		Rule: "{}", Revision: revision, Enable: false,
	}
}

func newFDRule(id, ns, revision string) *model.FaultDetectRule {
	return &model.FaultDetectRule{
		ID: id, Name: wPrefix + "fd", Namespace: ns, Description: "d",
		DstService: "*", DstNamespace: "*", DstMethod: "*",
		Rule: "{}", Revision: revision,
	}
}

func newRouterConfig(id, ns, revision string) *model.RouterConfig {
	return &model.RouterConfig{
		ID: id, Namespace: ns, Name: wPrefix + "route", Policy: "RulePolicy",
		Config: "{}", Enable: true, Priority: 0, Revision: revision,
		Description: "d", Valid: true,
	}
}

func TestSQLWriteLaneAndContract(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	ln := &laneStore{master: db, slave: db}
	sc := &serviceContractStore{master: db, slave: db}
	withTx := txRunner(db)

	nsName := wPrefix + "ns"
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))

	// 泳道
	laneID := wid("lane", 1)
	runCase(t, "AddLaneGroup", func() error {
		return withTx(func(tx store.Tx) error {
			return ln.AddLaneGroup(tx, newLaneGroup(laneID, "lrev-1"))
		})
	})
	runCase(t, "UpdateLaneGroup", func() error {
		return withTx(func(tx store.Tx) error {
			g := newLaneGroup(laneID, "lrev-2")
			g.Description = "updated"
			return ln.UpdateLaneGroup(tx, g)
		})
	})
	runCase(t, "DeleteLaneGroup", func() error { return ln.DeleteLaneGroup(laneID) })

	// 服务契约
	ctID := wid("contract", 1)
	runCase(t, "CreateServiceContract", func() error {
		return sc.CreateServiceContract(newContract(ctID, nsName, "crev-1"))
	})
	runCase(t, "UpdateServiceContract", func() error {
		return sc.UpdateServiceContract(newContract(ctID, nsName, "crev-2"))
	})
	runCase(t, "AddServiceContractInterfaces", func() error {
		return sc.AddServiceContractInterfaces(newEnrichContract(ctID, nsName, "crev-3"))
	})
	runCase(t, "AddServiceContractInterfaces/重复(ON CONFLICT 覆盖)", func() error {
		return sc.AddServiceContractInterfaces(newEnrichContract(ctID, nsName, "crev-4"))
	})
	runCase(t, "AppendServiceContractInterfaces", func() error {
		return sc.AppendServiceContractInterfaces(newEnrichContract(ctID, nsName, "crev-5"))
	})
	runCase(t, "DeleteServiceContractInterfaces", func() error {
		return sc.DeleteServiceContractInterfaces(newEnrichContract(ctID, nsName, "crev-6"))
	})
	runCase(t, "DeleteServiceContract", func() error {
		return sc.DeleteServiceContract(newContract(ctID, nsName, "crev-7"))
	})
}

func newLaneGroup(id, revision string) *model.LaneGroup {
	return &model.LaneGroup{
		ID: id, Name: wPrefix + "lane", Rule: "{}", Revision: revision,
		Description: "d", Valid: true, Labels: map[string]string{"lk": "lv"},
	}
}

func newContract(id, ns, revision string) *model.ServiceContract {
	return &model.ServiceContract{
		ID: id, Namespace: ns, Service: wPrefix + "svc", Type: wPrefix + "ct",
		Protocol: "http", Version: "1.0.0", Revision: revision, Content: "c",
	}
}

func newEnrichContract(id, ns, revision string) *model.EnrichServiceContract {
	return &model.EnrichServiceContract{
		ServiceContract: newContract(id, ns, revision),
		Interfaces: []*model.InterfaceDescriptor{{
			ID: wid("iface", 1), ContractID: id, Namespace: ns, Service: wPrefix + "svc",
			Protocol: "http", Version: "1.0.0", Type: "GET", Path: "/api/a",
			Method: "GET", Content: "ic", Revision: revision,
		}},
	}
}

func TestSQLWriteGrayAndRole(t *testing.T) {
	db := newWriteTestDB(t)
	gr := &grayStore{master: db, slave: db}
	rs := &roleStore{master: db, slave: db}
	ss := &strategyStore{master: db, slave: db}
	withTx := txRunner(db)

	// 灰度资源
	grayName := wPrefix + "gray"
	runCase(t, "CreateGrayResourceTx", func() error {
		return withTx(func(tx store.Tx) error {
			return gr.CreateGrayResourceTx(tx, &model.GrayResource{
				Name: grayName, MatchRule: `[{"k":"v1"}]`, CreateBy: "polaris", ModifyBy: "polaris",
			})
		})
	})
	runCase(t, "CreateGrayResourceTx/重复(ON CONFLICT 覆盖)", func() error {
		return withTx(func(tx store.Tx) error {
			return gr.CreateGrayResourceTx(tx, &model.GrayResource{
				Name: grayName, MatchRule: `[{"k":"v2"}]`, CreateBy: "u2", ModifyBy: "u2",
			})
		})
	})
	runCase(t, "CleanGrayResource", func() error {
		return withTx(func(tx store.Tx) error {
			return gr.CleanGrayResource(tx, &model.GrayResource{Name: grayName})
		})
	})
	runCase(t, "DeleteGrayResource", func() error {
		return withTx(func(tx store.Tx) error {
			return gr.DeleteGrayResource(tx, &model.GrayResource{Name: grayName})
		})
	})

	// 角色
	roleID := wid("role", 1)
	runCase(t, "AddRole", func() error { return rs.AddRole(newRole(roleID)) })
	runCase(t, "UpdateRole", func() error {
		r := newRole(roleID)
		r.Comment = "updated"
		return rs.UpdateRole(r)
	})
	runCase(t, "CleanPrincipalRoles", func() error {
		return withTx(func(tx store.Tx) error {
			return rs.CleanPrincipalRoles(tx, &authcommon.Principal{
				PrincipalID: wid("user", 1), PrincipalType: authcommon.PrincipalUser,
			})
		})
	})
	runCase(t, "CleanPrincipalPolicies", func() error {
		return withTx(func(tx store.Tx) error {
			return ss.CleanPrincipalPolicies(tx, authcommon.Principal{
				PrincipalID: wid("user", 1), PrincipalType: authcommon.PrincipalUser,
			})
		})
	})
	runCase(t, "DeleteRole", func() error {
		return withTx(func(tx store.Tx) error {
			return rs.DeleteRole(tx, &authcommon.Role{ID: roleID})
		})
	})
}

func newRole(id string) *authcommon.Role {
	return &authcommon.Role{
		ID: id, Name: wPrefix + "role", Owner: "polaris", Source: "Polaris", Type: "20",
		Comment: "c", Valid: true, Metadata: map[string]string{"rk": "rv"},
		Users:      []authcommon.Principal{},
		UserGroups: []authcommon.Principal{},
	}
}

func TestSQLWriteTemplateAndCleanup(t *testing.T) {
	db := newWriteTestDB(t)
	ct := &configFileTemplateStore{master: db, slave: db}
	cr := &configFileReleaseStore{master: db, slave: db}
	ch := &configFileReleaseHistoryStore{master: db, slave: db}
	withTx := txRunner(db)

	runCase(t, "CreateConfigFileTemplate", func() error {
		_, err := ct.CreateConfigFileTemplate(&model.ConfigFileTemplate{
			Name: wPrefix + "tpl", Content: "k: v", Comment: "c", Format: "yaml",
			CreateBy: "polaris", ModifyBy: "polaris",
		})
		return err
	})

	// 清理类 SQL：PostgreSQL 的 DELETE 不支持 LIMIT，这几处已改写为主键子查询
	runCase(t, "CleanDeletedConfigFileRelease", func() error {
		return cr.CleanDeletedConfigFileRelease(time.Now(), 10)
	})
	runCase(t, "CleanConfigFileReleaseHistory", func() error {
		return ch.CleanConfigFileReleaseHistory(time.Now(), 10)
	})
	runCase(t, "CleanConfigFileReleasesTx", func() error {
		return withTx(func(tx store.Tx) error {
			return cr.CleanConfigFileReleasesTx(tx, wPrefix+"ns", wPrefix+"grp", wPrefix+"file.yaml")
		})
	})

	// 清理测试产生的模板数据
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM config_file_template WHERE name LIKE 'sqlw-%'`)
	})
}

// transaction 承载 SELECT ... FOR UPDATE / FOR SHARE 等锁语句，
// MySQL 的 LOCK IN SHARE MODE 在 PostgreSQL 下改写为 FOR SHARE，需实际执行确认
func TestSQLWriteTransaction(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	ss := &serviceStore{master: db, slave: db}

	nsName := wPrefix + "ns"
	svcID := wid("svc", 1)
	svcName := wPrefix + "svc"
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))
	require.NoError(t, ss.AddService(&model.Service{
		ID: svcID, Name: svcName, Namespace: nsName, Token: "tk", Owner: "polaris", Revision: "r",
	}))
	// 别名指向上面的服务，用于验证 DeleteAliasWithSourceID
	require.NoError(t, ss.AddService(&model.Service{
		ID: wid("alias", 1), Name: wPrefix + "alias", Namespace: nsName,
		Reference: svcID, Token: "tk", Owner: "polaris", Revision: "ra",
	}))

	// 每个用例独立开启并立即结束事务：这些方法持有行锁（FOR UPDATE / FOR SHARE），
	// 若把提交推迟到测试函数结束，多个事务会互相等待造成死锁
	inTrans := func(fn func(*transaction) error) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		trans := &transaction{tx: tx}
		defer func() { _ = trans.Commit() }()
		return fn(trans)
	}

	// 必须使用 start_lock 中已存在的 key：LockBootstrap 内部对查出的 count 取模，
	// key 不存在时 count=0 会 integer divide by zero（MySQL 实现同样如此）
	runCase(t, "LockBootstrap", func() error {
		return inTrans(func(tr *transaction) error { return tr.LockBootstrap("sz", "server-1") })
	})
	runCase(t, "LockNamespace/FOR UPDATE", func() error {
		return inTrans(func(tr *transaction) error { _, err := tr.LockNamespace(nsName); return err })
	})
	runCase(t, "RLockNamespace/FOR SHARE", func() error {
		return inTrans(func(tr *transaction) error { _, err := tr.RLockNamespace(nsName); return err })
	})
	runCase(t, "LockService/FOR UPDATE", func() error {
		return inTrans(func(tr *transaction) error { _, err := tr.LockService(svcName, nsName); return err })
	})
	runCase(t, "RLockService/FOR SHARE", func() error {
		return inTrans(func(tr *transaction) error { _, err := tr.RLockService(svcName, nsName); return err })
	})
	runCase(t, "BatchRLockServices/FOR SHARE", func() error {
		return inTrans(func(tr *transaction) error {
			_, err := tr.BatchRLockServices(map[string]bool{svcID: true, wid("svc", 2): true})
			return err
		})
	})
	runCase(t, "DeleteAliasWithSourceID", func() error {
		return inTrans(func(tr *transaction) error { return tr.DeleteAliasWithSourceID(svcID) })
	})
	runCase(t, "DeleteService", func() error {
		return inTrans(func(tr *transaction) error { return tr.DeleteService(svcName, nsName) })
	})
	runCase(t, "DeleteNamespace", func() error {
		return inTrans(func(tr *transaction) error { return tr.DeleteNamespace(nsName) })
	})
}

func TestSQLWriteL5(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	ss := &serviceStore{master: db, slave: db}
	l5 := &l5Store{master: db, slave: db}

	nsName := wPrefix + "ns"
	svcID := wid("svc", 1)
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))
	require.NoError(t, ss.AddService(&model.Service{
		ID: svcID, Name: wPrefix + "svc", Namespace: nsName, Token: "tk", Owner: "polaris", Revision: "r",
	}))

	runCase(t, "SetL5Extend", func() error {
		_, err := l5.SetL5Extend(svcID, map[string]interface{}{
			"cmdbMod1": "m1", "cmdbMod2": "m2", "cmdbMod3": "m3",
		})
		return err
	})
	runCase(t, "GetL5Extend", func() error {
		_, err := l5.GetL5Extend(svcID)
		return err
	})
	// GenNextL5Sid 会更新 cl5_module 表，用 offset 0 limit 1 for update 取行
	runCase(t, "GenNextL5Sid", func() error {
		_, err := l5.GenNextL5Sid(6)
		return err
	})
}
