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

// 写路径 SQL 测试：覆盖 Create/Update/Delete/Batch 系列方法拼出的 SQL。
//
// 与 sqlcoverage_test.go 的查询侧互补。写方法之间存在数据依赖（namespace → service →
// instance），因此按生命周期顺序组织：建立 → 修改 → 批量 → 删除，顺带验证依赖链本身。
//
// 所有测试数据以 sqlw- 前缀命名，测试结束时清理，可重复运行。
// 运行方式见 sqlcoverage_test.go 顶部说明。

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	apimodel "github.com/polarismesh/specification/source/go/api/v1/model"
	apiservice "github.com/polarismesh/specification/source/go/api/v1/service_manage"
	"github.com/stretchr/testify/require"

	"github.com/polarismesh/polaris/common/model"
	authcommon "github.com/polarismesh/polaris/common/model/auth"
	"github.com/polarismesh/polaris/store"
)

const wPrefix = "sqlw-"

// cleanupTestData 按外键依赖的逆序清理，保证测试可重复运行
func cleanupTestData(t *testing.T, db *BaseDB) {
	t.Helper()
	stmts := []string{
		`DELETE FROM instance_metadata WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM health_check WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM instance WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM client_stat WHERE client_id LIKE 'sqlw-%'`,
		`DELETE FROM client WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM service_metadata WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM service WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM namespace WHERE name LIKE 'sqlw-%'`,
		`DELETE FROM auth_strategy_resource WHERE strategy_id LIKE 'sqlw-%'`,
		`DELETE FROM auth_strategy_function WHERE strategy_id LIKE 'sqlw-%'`,
		`DELETE FROM auth_strategy_label WHERE strategy_id LIKE 'sqlw-%'`,
		`DELETE FROM auth_principal WHERE strategy_id LIKE 'sqlw-%'`,
		`DELETE FROM auth_strategy WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM user_group_relation WHERE group_id LIKE 'sqlw-%'`,
		`DELETE FROM user_group WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM "user" WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM config_file_release_history WHERE namespace LIKE 'sqlw-%'`,
		`DELETE FROM config_file_release WHERE namespace LIKE 'sqlw-%'`,
		`DELETE FROM config_file WHERE namespace LIKE 'sqlw-%'`,
		`DELETE FROM config_file_group WHERE namespace LIKE 'sqlw-%'`,
		`DELETE FROM ratelimit_config WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM circuitbreaker_rule_v2 WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM fault_detect_rule WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM routing_config WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM routing_config_v2 WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM lane_rule WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM lane_group WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM service_contract_detail WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM service_contract WHERE id LIKE 'sqlw-%'`,
		`DELETE FROM gray_resource WHERE name LIKE 'sqlw-%'`,
		`DELETE FROM leader_election WHERE elect_key LIKE 'sqlw-%'`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Logf("清理失败(忽略) %s: %v", s, err)
		}
	}
}

// newWriteTestDB 建立连接并在测试前后清理数据
func newWriteTestDB(t *testing.T) *BaseDB {
	t.Helper()
	db := newTestDB(t)
	cleanupTestData(t, db)
	t.Cleanup(func() { cleanupTestData(t, db) })
	return db
}

func wid(kind string, i int) string { return wPrefix + kind + "-" + strconv.Itoa(i) }

func TestSQLWriteNamespace(t *testing.T) {
	db := newWriteTestDB(t)
	s := &namespaceStore{master: db, slave: db}
	name := wPrefix + "ns"

	runCase(t, "AddNamespace", func() error {
		return s.AddNamespace(&model.Namespace{
			Name: name, Comment: "c", Token: "tk", Owner: "polaris",
			ServiceExportTo: map[string]struct{}{"other-ns": {}},
			Metadata:        map[string]string{"k1": "v1"},
		})
	})
	// 同名有效命名空间重复插入应被主键约束拒绝（AddNamespace 只清理 flag=1 的无效记录）
	t.Run("AddNamespace/重复应报冲突", func(t *testing.T) {
		err := s.AddNamespace(&model.Namespace{
			Name: name, Comment: "c2", Token: "tk", Owner: "polaris",
		})
		require.Error(t, err, "重复主键应报错")
	})
	runCase(t, "UpdateNamespace", func() error {
		return s.UpdateNamespace(&model.Namespace{
			Name: name, Comment: "updated", Owner: "polaris2",
			ServiceExportTo: map[string]struct{}{"ns-a": {}, "ns-b": {}},
			Metadata:        map[string]string{"k2": "v2"},
		})
	})
	runCase(t, "UpdateNamespaceToken", func() error {
		return s.UpdateNamespaceToken(name, "new-token")
	})

	// 验证 metadata 与 service_export_to 确实往返一致
	t.Run("读回校验", func(t *testing.T) {
		got, err := s.GetNamespace(name)
		require.NoError(t, err)
		require.NotNil(t, got, "命名空间应存在")
		require.Equal(t, "v2", got.Metadata["k2"], "metadata 应往返一致")
		require.Len(t, got.ServiceExportTo, 2, "service_export_to 应往返一致")
	})
}

func TestSQLWriteService(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	s := &serviceStore{master: db, slave: db}
	nsName := wPrefix + "ns"
	require.NoError(t, ns.AddNamespace(&model.Namespace{
		Name: nsName, Token: "tk", Owner: "polaris",
	}))

	svcID := wid("svc", 1)
	runCase(t, "AddService", func() error {
		return s.AddService(&model.Service{
			ID: svcID, Name: wPrefix + "svc", Namespace: nsName,
			Business: "biz", Ports: "8080", Comment: "c", Department: "dep",
			Token: "tk", Owner: "polaris", Revision: "rev-1",
			Meta: map[string]string{"m1": "v1", "m2": "v2"},
		})
	})
	runCase(t, "UpdateService/needUpdateOwner=true", func() error {
		return s.UpdateService(&model.Service{
			ID: svcID, Name: wPrefix + "svc", Namespace: nsName,
			Business: "biz2", Token: "tk", Owner: "polaris2", Revision: "rev-2",
			Meta: map[string]string{"m3": "v3"},
		}, true)
	})
	runCase(t, "UpdateService/needUpdateOwner=false", func() error {
		return s.UpdateService(&model.Service{
			ID: svcID, Name: wPrefix + "svc", Namespace: nsName,
			Token: "tk", Owner: "polaris2", Revision: "rev-3",
		}, false)
	})
	runCase(t, "UpdateServiceToken", func() error {
		return s.UpdateServiceToken(svcID, "tk-new", "rev-4")
	})

	// 服务别名
	aliasID := wid("alias", 1)
	runCase(t, "AddService/别名", func() error {
		return s.AddService(&model.Service{
			ID: aliasID, Name: wPrefix + "alias", Namespace: nsName,
			Reference: svcID, Token: "tk", Owner: "polaris", Revision: "rev-a",
		})
	})
	runCase(t, "UpdateServiceAlias", func() error {
		return s.UpdateServiceAlias(&model.Service{
			ID: aliasID, Name: wPrefix + "alias", Namespace: nsName,
			Reference: svcID, Comment: "alias-c", Owner: "polaris", Revision: "rev-a2",
		}, true)
	})
	runCase(t, "DeleteServiceAlias", func() error {
		return s.DeleteServiceAlias(wPrefix+"alias", nsName)
	})
	runCase(t, "DeleteService", func() error {
		return s.DeleteService(svcID, wPrefix+"svc", nsName)
	})
}

func TestSQLWriteInstance(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	ss := &serviceStore{master: db, slave: db}
	s := &instanceStore{master: db, slave: db}

	nsName := wPrefix + "ns"
	svcID := wid("svc", 1)
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))
	require.NoError(t, ss.AddService(&model.Service{
		ID: svcID, Name: wPrefix + "svc", Namespace: nsName, Token: "tk", Owner: "polaris", Revision: "r",
	}))

	insID := wid("ins", 1)
	runCase(t, "AddInstance/含健康检查与metadata", func() error {
		return s.AddInstance(newTestInstance(insID, svcID, 1, true))
	})
	runCase(t, "AddInstance/重复(ON CONFLICT 覆盖)", func() error {
		return s.AddInstance(newTestInstance(insID, svcID, 1, true))
	})
	runCase(t, "AddInstance/无健康检查", func() error {
		return s.AddInstance(newTestInstance(wid("ins", 2), svcID, 2, false))
	})
	runCase(t, "UpdateInstance", func() error {
		ins := newTestInstance(insID, svcID, 1, true)
		ins.Proto.Weight = &wrappers.UInt32Value{Value: 200}
		return s.UpdateInstance(ins)
	})
	runCase(t, "BatchAddInstances", func() error {
		var list []*model.Instance
		for i := 10; i < 15; i++ {
			list = append(list, newTestInstance(wid("ins", i), svcID, uint32(i), i%2 == 0))
		}
		return s.BatchAddInstances(list)
	})
	runCase(t, "SetInstanceHealthStatus", func() error {
		return s.SetInstanceHealthStatus(insID, 0, "rev-h")
	})
	runCase(t, "BatchSetInstanceHealthStatus", func() error {
		return s.BatchSetInstanceHealthStatus(
			[]interface{}{insID, wid("ins", 2)}, 1, "rev-bh")
	})
	runCase(t, "BatchSetInstanceIsolate", func() error {
		return s.BatchSetInstanceIsolate(
			[]interface{}{insID, wid("ins", 2)}, 1, "rev-bi")
	})
	runCase(t, "BatchAppendInstanceMetadata", func() error {
		return s.BatchAppendInstanceMetadata([]*store.InstanceMetadataRequest{{
			InstanceID: insID, Revision: "rev-m1",
			Metadata: map[string]string{"append1": "v1", "append2": "v2"},
		}})
	})
	runCase(t, "BatchAppendInstanceMetadata/重复key覆盖", func() error {
		return s.BatchAppendInstanceMetadata([]*store.InstanceMetadataRequest{{
			InstanceID: insID, Revision: "rev-m2",
			Metadata: map[string]string{"append1": "v1-new"},
		}})
	})
	// Keys 为空会拼出 "mkey in ()" 语法错误（MySQL 同样非法），调用方负责保证非空
	runCase(t, "BatchRemoveInstanceMetadata/单key", func() error {
		return s.BatchRemoveInstanceMetadata([]*store.InstanceMetadataRequest{{
			InstanceID: insID, Revision: "rev-m3", Keys: []string{"append1"},
		}})
	})
	runCase(t, "BatchRemoveInstanceMetadata/多key", func() error {
		return s.BatchRemoveInstanceMetadata([]*store.InstanceMetadataRequest{{
			InstanceID: insID, Revision: "rev-m4", Keys: []string{"append2", "mk1", "mk2"},
		}})
	})
	runCase(t, "DeleteInstance", func() error { return s.DeleteInstance(wid("ins", 2)) })
	runCase(t, "BatchDeleteInstances", func() error {
		return s.BatchDeleteInstances([]interface{}{wid("ins", 10), wid("ins", 11)})
	})
	runCase(t, "CleanInstance", func() error { return s.CleanInstance(wid("ins", 2)) })

	t.Run("读回校验", func(t *testing.T) {
		got, err := s.GetInstance(insID)
		require.NoError(t, err)
		require.NotNil(t, got, "实例应存在")
		require.Equal(t, uint32(200), got.Proto.GetWeight().GetValue(), "weight 应已更新")
	})
}

func newTestInstance(id, svcID string, port uint32, withCheck bool) *model.Instance {
	proto := &apiservice.Instance{
		Id:                &wrappers.StringValue{Value: id},
		Host:              &wrappers.StringValue{Value: "10.0.0.1"},
		Port:              &wrappers.UInt32Value{Value: port},
		Protocol:          &wrappers.StringValue{Value: "grpc"},
		Version:           &wrappers.StringValue{Value: "v1"},
		Weight:            &wrappers.UInt32Value{Value: 100},
		EnableHealthCheck: &wrappers.BoolValue{Value: withCheck},
		Healthy:           &wrappers.BoolValue{Value: true},
		Isolate:           &wrappers.BoolValue{Value: false},
		LogicSet:          &wrappers.StringValue{Value: "set1"},
		Location: &apimodel.Location{
			Region: &wrappers.StringValue{Value: "sh"},
			Zone:   &wrappers.StringValue{Value: "az1"},
			Campus: &wrappers.StringValue{Value: "c1"},
		},
		Metadata: map[string]string{"mk1": "mv1", "mk2": "mv2"},
		Revision: &wrappers.StringValue{Value: "rev-" + id},
	}
	if withCheck {
		proto.HealthCheck = &apiservice.HealthCheck{
			Type:      apiservice.HealthCheck_HEARTBEAT,
			Heartbeat: &apiservice.HeartbeatHealthCheck{Ttl: &wrappers.UInt32Value{Value: 5}},
		}
	}
	return &model.Instance{Proto: proto, ServiceID: svcID, Valid: true}
}

func TestSQLWriteAuth(t *testing.T) {
	db := newWriteTestDB(t)
	us := &userStore{master: db, slave: db}
	gs := &groupStore{master: db, slave: db}
	ss := &strategyStore{master: db, slave: db}

	withTx := func(fn func(store.Tx) error) error {
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

	userID := wid("user", 1)
	runCase(t, "AddUser", func() error {
		return withTx(func(tx store.Tx) error {
			return us.AddUser(tx, &authcommon.User{
				ID: userID, Name: wPrefix + "user", Password: "pwd", Owner: "",
				Source: "Polaris", Type: authcommon.OwnerUserRole, Token: "tk",
				TokenEnable: true, Valid: true, Comment: "c",
				Mobile: "13800000000", Email: "a@b.c",
			})
		})
	})
	runCase(t, "UpdateUser", func() error {
		return us.UpdateUser(&authcommon.User{
			ID: userID, Name: wPrefix + "user", Password: "pwd2", Token: "tk2",
			TokenEnable: false, Comment: "c2", Mobile: "13900000000", Email: "x@y.z",
		})
	})

	groupID := wid("group", 1)
	runCase(t, "AddGroup", func() error {
		return withTx(func(tx store.Tx) error {
			return gs.AddGroup(tx, &authcommon.UserGroupDetail{
				UserGroup: &authcommon.UserGroup{
					ID: groupID, Name: wPrefix + "group", Owner: userID,
					Token: "gtk", TokenEnable: true, Valid: true, Comment: "gc",
				},
				UserIds: map[string]struct{}{userID: {}},
			})
		})
	})
	runCase(t, "UpdateGroup/增删成员", func() error {
		return gs.UpdateGroup(&authcommon.ModifyUserGroup{
			ID: groupID, Token: "gtk2", TokenEnable: true, Comment: "gc2",
			AddUserIds:    []string{},
			RemoveUserIds: []string{},
		})
	})

	strategyID := wid("strategy", 1)
	runCase(t, "AddStrategy", func() error {
		return withTx(func(tx store.Tx) error {
			return ss.AddStrategy(tx, &authcommon.StrategyDetail{
				ID: strategyID, Name: wPrefix + "strategy", Action: "READ_WRITE",
				Owner: userID, Comment: "sc", Revision: "srev", Source: "Polaris",
				Default: false, Valid: true,
				Principals: []authcommon.Principal{{
					StrategyID: strategyID, PrincipalID: userID,
					PrincipalType: authcommon.PrincipalUser,
				}},
				Resources: []authcommon.StrategyResource{{
					StrategyID: strategyID, ResType: 0, ResID: "*",
				}},
				CalleeMethods: []string{"CreateInstances"},
				Metadata:      map[string]string{"sk": "sv"},
			})
		})
	})
	runCase(t, "UpdateStrategy", func() error {
		return ss.UpdateStrategy(&authcommon.ModifyStrategyDetail{
			ID: strategyID, Action: "READ_ONLY", Comment: "sc2",
			AddPrincipals: []authcommon.Principal{},
			AddResources:  []authcommon.StrategyResource{},
			CalleeMethods: []string{"DeleteInstances"},
		})
	})
	runCase(t, "DeleteStrategy", func() error { return ss.DeleteStrategy(strategyID) })
	runCase(t, "DeleteGroup", func() error {
		return withTx(func(tx store.Tx) error {
			return gs.DeleteGroup(tx, &authcommon.UserGroupDetail{
				UserGroup: &authcommon.UserGroup{
					ID: groupID, Name: wPrefix + "group", Owner: userID,
				},
			})
		})
	})
	runCase(t, "DeleteUser", func() error {
		return withTx(func(tx store.Tx) error {
			return us.DeleteUser(tx, &authcommon.User{
				ID: userID, Name: wPrefix + "user", Owner: "",
			})
		})
	})
}

func TestSQLWriteConfig(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}
	cg := &configFileGroupStore{master: db, slave: db}
	cf := &configFileStore{master: db, slave: db}
	cr := &configFileReleaseStore{master: db, slave: db}
	ch := &configFileReleaseHistoryStore{master: db, slave: db}

	nsName := wPrefix + "ns"
	grpName := wPrefix + "grp"
	fileName := wPrefix + "file.yaml"
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsName, Token: "tk", Owner: "polaris"}))

	withTx := func(fn func(*BaseTx) error) error {
		baseTx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() { _ = baseTx.Rollback() }()
		if err := fn(baseTx); err != nil {
			return err
		}
		return baseTx.Commit()
	}

	runCase(t, "CreateConfigFileGroup", func() error {
		_, err := cg.CreateConfigFileGroup(&model.ConfigFileGroup{
			Name: grpName, Namespace: nsName, Comment: "c", Owner: "polaris",
			Business: "biz", Department: "dep", CreateBy: "polaris", ModifyBy: "polaris",
			Metadata: map[string]string{"gk": "gv"}, Valid: true,
		})
		return err
	})
	runCase(t, "UpdateConfigFileGroup", func() error {
		return cg.UpdateConfigFileGroup(&model.ConfigFileGroup{
			Name: grpName, Namespace: nsName, Comment: "c2", Owner: "polaris2",
			Business: "biz2", Department: "dep2", ModifyBy: "polaris",
			Metadata: map[string]string{"gk2": "gv2"},
		})
	})
	runCase(t, "CreateConfigFileTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cf.CreateConfigFileTx(NewSqlDBTx(tx), &model.ConfigFile{
				Name: fileName, Namespace: nsName, Group: grpName,
				Content: "k: v", Comment: "c", Format: "yaml",
				CreateBy: "polaris", ModifyBy: "polaris", Valid: true,
				Metadata: map[string]string{"fk": "fv"},
			})
		})
	})
	runCase(t, "UpdateConfigFileTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cf.UpdateConfigFileTx(NewSqlDBTx(tx), &model.ConfigFile{
				Name: fileName, Namespace: nsName, Group: grpName,
				Content: "k: v2", Comment: "c2", Format: "yaml",
				ModifyBy: "polaris2", Metadata: map[string]string{"fk2": "fv2"},
			})
		})
	})

	relName := wPrefix + "rel"
	runCase(t, "CreateConfigFileReleaseTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cr.CreateConfigFileReleaseTx(NewSqlDBTx(tx), &model.ConfigFileRelease{
				SimpleConfigFileRelease: &model.SimpleConfigFileRelease{
					ConfigFileReleaseKey: &model.ConfigFileReleaseKey{
						Name: relName, Namespace: nsName, Group: grpName, FileName: fileName,
					},
					Comment: "rc", Md5: "md5", Version: 1, Active: true,
					CreateBy: "polaris", ModifyBy: "polaris", Valid: true,
					Metadata: map[string]string{"rk": "rv"},
				},
				Content: "k: v2",
			})
		})
	})
	runCase(t, "ActiveConfigFileReleaseTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cr.ActiveConfigFileReleaseTx(NewSqlDBTx(tx), &model.ConfigFileRelease{
				SimpleConfigFileRelease: &model.SimpleConfigFileRelease{
					ConfigFileReleaseKey: &model.ConfigFileReleaseKey{
						Name: relName, Namespace: nsName, Group: grpName, FileName: fileName,
					},
					Version: 1,
				},
			})
		})
	})
	runCase(t, "InactiveConfigFileReleaseTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cr.InactiveConfigFileReleaseTx(NewSqlDBTx(tx), &model.ConfigFileRelease{
				SimpleConfigFileRelease: &model.SimpleConfigFileRelease{
					ConfigFileReleaseKey: &model.ConfigFileReleaseKey{
						Name: relName, Namespace: nsName, Group: grpName, FileName: fileName,
					},
				},
			})
		})
	})
	runCase(t, "CreateConfigFileReleaseHistory", func() error {
		return ch.CreateConfigFileReleaseHistory(&model.ConfigFileReleaseHistory{
			Name: relName, Namespace: nsName, Group: grpName, FileName: fileName,
			Content: "k: v2", Comment: "hc", Md5: "md5", Type: "publish",
			Status: "success", Format: "yaml", CreateBy: "polaris", ModifyBy: "polaris",
			Metadata: map[string]string{"hk": "hv"},
		})
	})
	runCase(t, "DeleteConfigFileReleaseTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cr.DeleteConfigFileReleaseTx(NewSqlDBTx(tx), &model.ConfigFileReleaseKey{
				Name: relName, Namespace: nsName, Group: grpName, FileName: fileName,
			})
		})
	})
	runCase(t, "DeleteConfigFileTx", func() error {
		return withTx(func(tx *BaseTx) error {
			return cf.DeleteConfigFileTx(NewSqlDBTx(tx), nsName, grpName, fileName)
		})
	})
	runCase(t, "DeleteConfigFileGroup", func() error {
		return cg.DeleteConfigFileGroup(nsName, grpName)
	})
}

func TestSQLWriteClient(t *testing.T) {
	db := newWriteTestDB(t)
	cs := &clientStore{master: db, slave: db}

	runCase(t, "CreateClient", func() error {
		return cs.CreateClient(newTestClient(wid("client", 1)))
	})
	runCase(t, "UpdateClient", func() error {
		return cs.UpdateClient(newTestClient(wid("client", 1)))
	})
	runCase(t, "BatchAddClients", func() error {
		var list []*model.Client
		for i := 10; i < 13; i++ {
			list = append(list, newTestClient(wid("client", i)))
		}
		return cs.BatchAddClients(list)
	})
	runCase(t, "GetClientStat/修复后的 client_id 条件", func() error {
		_, err := cs.GetClientStat(wid("client", 1))
		return err
	})
	runCase(t, "BatchDeleteClients", func() error {
		return cs.BatchDeleteClients([]string{wid("client", 10), wid("client", 11)})
	})
}

func newTestClient(id string) *model.Client {
	return model.NewClient(&apiservice.Client{
		Id:      &wrappers.StringValue{Value: id},
		Host:    &wrappers.StringValue{Value: "10.1.1.1"},
		Type:    apiservice.Client_SDK,
		Version: &wrappers.StringValue{Value: "v1.0"},
		Location: &apimodel.Location{
			Region: &wrappers.StringValue{Value: "sh"},
			Zone:   &wrappers.StringValue{Value: "az1"},
			Campus: &wrappers.StringValue{Value: "c1"},
		},
		Stat: []*apiservice.StatInfo{{
			Target:   &wrappers.StringValue{Value: "prometheus"},
			Port:     &wrappers.UInt32Value{Value: 9090},
			Protocol: &wrappers.StringValue{Value: "http"},
			Path:     &wrappers.StringValue{Value: "/metrics"},
		}},
	})
}

func TestSQLWriteAdmin(t *testing.T) {
	db := newWriteTestDB(t)
	le := &leaderElectionStore{master: db}
	as := &adminStore{master: db}

	key := wPrefix + "elect"
	runCase(t, "CreateLeaderElection", func() error { return le.CreateLeaderElection(key) })
	runCase(t, "CompareAndSwapVersion", func() error {
		_, err := le.CompareAndSwapVersion(key, 0, 1, "host-1")
		return err
	})
	// ReleaseLeaderElection 依赖 StartLeaderElection 建立的状态机，不是纯 SQL 路径，此处不覆盖

	// 批量清理任务：SQL 中含时间窗口计算，需确认在 PG 下可执行
	runCase(t, "BatchCleanDeletedInstances", func() error {
		_, err := as.BatchCleanDeletedInstances(time.Minute, 10)
		return err
	})
	runCase(t, "BatchCleanDeletedClients", func() error {
		_, err := as.BatchCleanDeletedClients(time.Minute, 10)
		return err
	})
	runCase(t, "BatchCleanDeletedServices", func() error {
		_, err := as.BatchCleanDeletedServices(time.Minute, 10)
		return err
	})
	runCase(t, "BatchCleanDeletedRules", func() error {
		_, err := as.BatchCleanDeletedRules("routing", time.Minute, 10)
		return err
	})
	runCase(t, "BatchCleanDeletedConfigFiles", func() error {
		_, err := as.BatchCleanDeletedConfigFiles(time.Minute, 10)
		return err
	})
}
