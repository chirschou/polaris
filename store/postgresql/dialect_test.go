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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertPlaceholders(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "无占位符时原样返回",
			input:    "select count(*) from instance",
			expected: "select count(*) from instance",
		},
		{
			name:     "单个占位符",
			input:    "select id from instance where host = ?",
			expected: "select id from instance where host = $1",
		},
		{
			name:     "多个占位符按出现顺序编号",
			input:    "update instance set host = ?, port = ? where id = ?",
			expected: "update instance set host = $1, port = $2 where id = $3",
		},
		{
			name:     "批量插入的值列表",
			input:    "insert into t(a,b) values (?,?),(?,?),(?,?)",
			expected: "insert into t(a,b) values ($1,$2),($3,$4),($5,$6)",
		},
		{
			name:     "编号可超过 9 位数",
			input:    "values (?,?,?,?,?,?,?,?,?,?,?)",
			expected: "values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		},
		{
			// 字符串字面量内的 ? 是数据而非占位符
			name:     "单引号字面量内的问号保持原样",
			input:    "select id from t where name = '?' and flag = ?",
			expected: "select id from t where name = '?' and flag = $1",
		},
		{
			name:     "字面量内含多个问号",
			input:    "select * from t where a = 'a?b?c' and b = ?",
			expected: "select * from t where a = 'a?b?c' and b = $1",
		},
		{
			// '' 是单引号内部的转义写法，不应被当作引号状态切换
			name:     "字面量内的双写单引号转义",
			input:    "select * from t where a = 'it''s ?' and b = ?",
			expected: "select * from t where a = 'it''s ?' and b = $1",
		},
		{
			name:     "反斜杠转义的单引号",
			input:    `select * from t where a = 'x\'? ' and b = ?`,
			expected: `select * from t where a = 'x\'? ' and b = $1`,
		},
		{
			name:     "COALESCE 的空串默认值不受影响",
			input:    "select COALESCE(version, '') from client where id = ?",
			expected: "select COALESCE(version, '') from client where id = $1",
		},
		{
			// 分页参数顺序保持 (offset, limit)
			name:     "OFFSET/LIMIT 占位符",
			input:    "select * from t order by mtime desc offset ? limit ?",
			expected: "select * from t order by mtime desc offset $1 limit $2",
		},
		{
			name:     "ON CONFLICT 语句",
			input:    "insert into t(id,v) values (?,?) on conflict (id) do update set v = EXCLUDED.v",
			expected: "insert into t(id,v) values ($1,$2) on conflict (id) do update set v = EXCLUDED.v",
		},
		{
			name:     "双引号标识符内不含占位符时不受影响",
			input:    `select "group" from config_file where namespace = ?`,
			expected: `select "group" from config_file where namespace = $1`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, convertPlaceholders(tc.input))
		})
	}
}

func TestNormalizeArgs(t *testing.T) {
	t.Run("bool 转换为 0/1", func(t *testing.T) {
		out := normalizeArgs([]interface{}{true, false})
		assert.Equal(t, []interface{}{1, 0}, out)
	})

	t.Run("非 bool 参数保持原值原类型", func(t *testing.T) {
		in := []interface{}{"host", 8080, int64(100), nil}
		out := normalizeArgs(in)
		assert.Equal(t, in, out)
	})

	t.Run("混合参数只转换 bool", func(t *testing.T) {
		out := normalizeArgs([]interface{}{"id-1", true, 3.14, false, nil})
		assert.Equal(t, []interface{}{"id-1", 1, 3.14, 0, nil}, out)
	})

	t.Run("不含 bool 时返回原切片避免额外分配", func(t *testing.T) {
		in := []interface{}{"a", 1}
		assert.Equal(t, &in[0], &normalizeArgs(in)[0])
	})

	t.Run("含 bool 时不修改调用方持有的切片", func(t *testing.T) {
		in := []interface{}{true, "x"}
		out := normalizeArgs(in)
		assert.Equal(t, true, in[0], "入参切片不应被就地修改")
		assert.Equal(t, 1, out[0])
	})

	t.Run("空参数", func(t *testing.T) {
		assert.Empty(t, normalizeArgs(nil))
	})
}
