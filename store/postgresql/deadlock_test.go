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

// 验证 base_db.go 中 errMsg 里的 PostgreSQL 错误关键词确实能匹配 lib/pq 抛出的错误。
//
// errMsg 决定 Retry 是否重试。移植时把 MySQL 的 "Deadlock" 换成了 PG 的
// "deadlock detected" / 40P01，若关键词与驱动实际返回的文本对不上，
// 死锁重试会静默失效——这里制造真实死锁来确认。

import (
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/polarismesh/polaris/common/model"
)

// TestPGDeadlockErrorMatchesRetryKeyword 制造真实死锁，确认错误文本能被 errMsg 命中
func TestPGDeadlockErrorMatchesRetryKeyword(t *testing.T) {
	db := newWriteTestDB(t)
	ns := &namespaceStore{master: db, slave: db}

	// 两行数据，供两个事务交叉加锁
	nsA, nsB := wPrefix+"dl-a", wPrefix+"dl-b"
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsA, Token: "tk", Owner: "polaris"}))
	require.NoError(t, ns.AddNamespace(&model.Namespace{Name: nsB, Token: "tk", Owner: "polaris"}))

	lockRow := func(tx *BaseTx, name string) error {
		_, err := tx.Exec("UPDATE namespace SET owner = ? WHERE name = ?", "dl", name)
		return err
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		deadlock error
	)
	record := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if deadlock == nil {
			deadlock = err
		}
	}

	// tx1 先锁 A，tx2 先锁 B，然后互相去锁对方持有的行
	step := make(chan struct{})
	wg.Add(2)
	go func() {
		defer wg.Done()
		tx, err := db.Begin()
		if err != nil {
			record(err)
			return
		}
		defer func() { _ = tx.Rollback() }()
		if err := lockRow(tx, nsA); err != nil {
			record(err)
			return
		}
		close(step)
		time.Sleep(200 * time.Millisecond)
		record(lockRow(tx, nsB))
	}()
	go func() {
		defer wg.Done()
		tx, err := db.Begin()
		if err != nil {
			record(err)
			return
		}
		defer func() { _ = tx.Rollback() }()
		<-step
		if err := lockRow(tx, nsB); err != nil {
			record(err)
			return
		}
		time.Sleep(200 * time.Millisecond)
		record(lockRow(tx, nsA))
	}()
	wg.Wait()

	require.NotNil(t, deadlock, "应当制造出死锁；未产生说明测试本身没构成循环等待")
	t.Logf("PostgreSQL 返回的死锁错误: %v", deadlock)

	// 核心断言：该错误必须被判定为可重试，否则 Retry 会直接放弃
	require.True(t, isRetryableErr(deadlock),
		"死锁错误未被判定为可重试，Retry 将不会重试。实际错误=%v", deadlock)

	// SQLSTATE 判断必须独立生效——lib/pq 的 Error() 不含 SQLSTATE，
	// 若只靠文本匹配，一旦 PG 改措辞或换 locale 就会失效
	var pqErr *pq.Error
	require.ErrorAs(t, deadlock, &pqErr)
	require.Equal(t, pq.ErrorCode("40P01"), pqErr.Code, "死锁的 SQLSTATE 应为 40P01")
	_, bySQLState := retryableSQLState[pqErr.Code]
	require.True(t, bySQLState, "应能仅凭 SQLSTATE 判定为可重试")
	t.Logf("SQLSTATE=%s，文本=%q", pqErr.Code, pqErr.Message)
}

// TestRetryTriggersOnPGDeadlockText 确认 Retry 对该错误文本确实会重试
func TestRetryTriggersOnPGDeadlockText(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过重试计时测试")
	}
	attempts := 0
	start := time.Now()
	Retry("pg-deadlock", func() error {
		attempts++
		if attempts >= 3 {
			return nil
		}
		// 使用 PostgreSQL 死锁错误的真实文本形态
		return &fakeErr{"pq: deadlock detected (SQLSTATE 40P01)"}
	})
	require.Equal(t, 3, attempts, "应当重试到成功")
	require.Greater(t, time.Since(start), 5*time.Millisecond, "重试之间应有退避等待")
}

type fakeErr struct{ s string }

func (e *fakeErr) Error() string { return e.s }

// TestSerializationFailureRetryable 确认 REPEATABLE READ 下的序列化冲突同样可重试。
// CreateReadView 会把事务设为 REPEATABLE READ，该场景下并发更新会抛 40001，
// 其文本是 "could not serialize access due to concurrent update"，不含 SQLSTATE。
func TestSerializationFailureRetryable(t *testing.T) {
	bySQLState := &pq.Error{Code: "40001", Message: "could not serialize access due to concurrent update"}
	require.True(t, isRetryableErr(bySQLState), "序列化冲突应判定为可重试")

	// 即便拿不到 *pq.Error（例如被包装成普通 error），文本兜底也应生效
	require.True(t, isRetryableErr(&fakeErr{"pq: could not serialize access due to concurrent update"}),
		"文本兜底应能识别序列化冲突")

	// 非冲突类错误不应触发重试
	require.False(t, isRetryableErr(&pq.Error{Code: "23505", Message: "duplicate key value"}),
		"唯一键冲突不应重试")
	require.False(t, isRetryableErr(&fakeErr{"pq: syntax error at or near \"foo\""}),
		"语法错误不应重试")
}
