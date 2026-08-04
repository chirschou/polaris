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

import "github.com/polarismesh/polaris/store"

type Tx struct {
	delegateTx *BaseTx
}

func NewSqlDBTx(delegateTx *BaseTx) store.Tx {
	return &Tx{
		delegateTx: delegateTx,
	}
}

func (t *Tx) Commit() error {
	return t.delegateTx.Commit()
}

func (t *Tx) Rollback() error {
	return t.delegateTx.Rollback()
}

func (t *Tx) GetDelegateTx() interface{} {
	return t.delegateTx
}

func (t *Tx) CreateReadView() error {
	tx := t.delegateTx
	// PostgreSQL 没有 MySQL 的 START TRANSACTION WITH CONSISTENT SNAPSHOT，
	// 等价语义由 REPEATABLE READ 隔离级别提供；该语句必须先于事务内任何其它语句执行。
	if _, err := tx.Exec("SET TRANSACTION ISOLATION LEVEL REPEATABLE READ"); err != nil {
		return err
	}
	// REPEATABLE READ 下快照在首条语句时才建立，这里主动触发一次以对齐 MySQL 的立即建快照行为
	_, err := tx.Exec("SELECT 1")
	return err
}
