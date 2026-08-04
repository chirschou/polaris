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
	"strconv"
	"strings"
)

// convertPlaceholders 将 SQL 中的 ? 占位符按出现顺序转换为 PostgreSQL 的 $N 形式。
//
// 本包的 SQL 语句大量通过字符串拼接动态生成（如 PlaceholdersN 构造批量插入的值列表），
// 拼接时无法预知占位符的全局序号，因此统一在执行入口做一次线性改写。
// 单引号字符串字面量内部的 ? 属于数据而非占位符，必须原样保留。
func convertPlaceholders(query string) string {
	if !strings.ContainsRune(query, '?') {
		return query
	}

	var sb strings.Builder
	sb.Grow(len(query) + 16)

	seq := 0
	inQuote := false
	for i := 0; i < len(query); i++ {
		c := query[i]
		switch {
		case c == '\'':
			// '' 是单引号字面量内部的转义写法，整体跳过，不切换引号状态
			if inQuote && i+1 < len(query) && query[i+1] == '\'' {
				sb.WriteString("''")
				i++
				continue
			}
			inQuote = !inQuote
			sb.WriteByte(c)
		case c == '\\' && inQuote && i+1 < len(query):
			// 反斜杠转义，连同下一个字节一起原样写出
			sb.WriteByte(c)
			sb.WriteByte(query[i+1])
			i++
		case c == '?' && !inQuote:
			seq++
			sb.WriteByte('$')
			sb.WriteString(strconv.Itoa(seq))
		default:
			sb.WriteByte(c)
		}
	}

	return sb.String()
}

// normalizeArgs 把 bool 参数转换为 0/1。
//
// MySQL 驱动会将 Go 的 bool 编码为 0/1，而 lib/pq 发送的是 true/false。
// 本 schema 沿用 MySQL 的建模方式，用 smallint 表示布尔语义（如 flag、health_status、
// isolate），直接发送 true/false 会被 PostgreSQL 拒绝。
func normalizeArgs(args []interface{}) []interface{} {
	var out []interface{}
	for i, arg := range args {
		b, ok := arg.(bool)
		if !ok {
			continue
		}
		// 仅在确实含 bool 时复制一份，避免修改调用方持有的切片
		if out == nil {
			out = make([]interface{}, len(args))
			copy(out, args)
		}
		if b {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
	if out == nil {
		return args
	}
	return out
}
