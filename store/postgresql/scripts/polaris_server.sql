--
-- Polaris PostgreSQL schema
-- 由 store/mysql/scripts/polaris_server.sql 转换而来
--

/*
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
SET
    timezone = 'UTC';

--
-- Database: polaris_server
--
-- 建库需以超级用户连接到 postgres 库执行，随后再连接到 polaris_server 库执行本文件其余部分：
--     CREATE DATABASE polaris_server ENCODING 'UTF8';
--     \c polaris_server
--

-- --------------------------------------------------------
--
-- Table structure "instance"
--
CREATE TABLE "instance" (
    "id" VARCHAR(128) NOT NULL,
    "service_id" VARCHAR(32) NOT NULL,
    "vpc_id" VARCHAR(64) DEFAULT NULL,
    "host" VARCHAR(128) NOT NULL,
    "port" INTEGER NOT NULL,
    "protocol" VARCHAR(32) DEFAULT NULL,
    "version" VARCHAR(32) DEFAULT NULL,
    "health_status" SMALLINT NOT NULL DEFAULT '1',
    "isolate" SMALLINT NOT NULL DEFAULT '0',
    "weight" SMALLINT NOT NULL DEFAULT '100',
    "enable_health_check" SMALLINT NOT NULL DEFAULT '0',
    "logic_set" VARCHAR(128) DEFAULT NULL,
    "cmdb_region" VARCHAR(128) DEFAULT NULL,
    "cmdb_zone" VARCHAR(128) DEFAULT NULL,
    "cmdb_idc" VARCHAR(128) DEFAULT NULL,
    "priority" SMALLINT NOT NULL DEFAULT '0',
    "revision" VARCHAR(32) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "health_check"
--
CREATE TABLE "health_check" (
    "id" VARCHAR(128) NOT NULL,
    "type" SMALLINT NOT NULL DEFAULT '0',
    "ttl" INTEGER NOT NULL,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "instance_metadata"
--
CREATE TABLE "instance_metadata" (
    "id" VARCHAR(128) NOT NULL,
    "mkey" VARCHAR(128) NOT NULL,
    "mvalue" VARCHAR(4096) NOT NULL,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id", "mkey")
);

-- --------------------------------------------------------
--
-- Table structure "namespace"
--
CREATE TABLE "namespace" (
    "name" VARCHAR(64) NOT NULL,
    "comment" VARCHAR(1024) DEFAULT NULL,
    "token" VARCHAR(64) NOT NULL,
    "owner" VARCHAR(1024) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "service_export_to" TEXT,
    "metadata" TEXT,
    PRIMARY KEY ("name")
);

--
-- Data in the conveyor "namespace"
--
INSERT INTO
    "namespace" (
        "name",
        "comment",
        "token",
        "owner",
        "flag",
        "ctime",
        "mtime"
    )
VALUES
    (
        'Polaris',
        'Polaris-server',
        '2d1bfe5d12e04d54b8ee69e62494c7fd',
        'polaris',
        0,
        '2019-09-06 07:55:07',
        '2019-09-06 07:55:07'
    ),
    (
        'default',
        'Default Environment',
        'e2e473081d3d4306b52264e49f7ce227',
        'polaris',
        0,
        '2021-07-27 19:37:37',
        '2021-07-27 19:37:37'
    );

-- --------------------------------------------------------
--
-- Table structure "routing_config"
--
CREATE TABLE "routing_config" (
    "id" VARCHAR(32) NOT NULL,
    "in_bounds" TEXT,
    "out_bounds" TEXT,
    "revision" VARCHAR(40) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "ratelimit_config"
--
CREATE TABLE "ratelimit_config" (
    "id" VARCHAR(32) NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "disable" SMALLINT NOT NULL DEFAULT '0',
    "service_id" VARCHAR(32) NOT NULL,
    "method" VARCHAR(512) NOT NULL,
    "labels" TEXT NOT NULL,
    "priority" SMALLINT NOT NULL DEFAULT '0',
    "rule" TEXT NOT NULL,
    "revision" VARCHAR(32) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "etime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "ratelimit_revision"
--
CREATE TABLE "ratelimit_revision" (
    "service_id" VARCHAR(32) NOT NULL,
    "last_revision" VARCHAR(40) NOT NULL,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("service_id")
);

-- --------------------------------------------------------
--
-- Table structure "service"
--
CREATE TABLE "service" (
    "id" VARCHAR(32) NOT NULL,
    "name" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "ports" TEXT DEFAULT NULL,
    "business" VARCHAR(64) DEFAULT NULL,
    "department" VARCHAR(1024) DEFAULT NULL,
    "cmdb_mod1" VARCHAR(1024) DEFAULT NULL,
    "cmdb_mod2" VARCHAR(1024) DEFAULT NULL,
    "cmdb_mod3" VARCHAR(1024) DEFAULT NULL,
    "comment" VARCHAR(1024) DEFAULT NULL,
    "token" VARCHAR(2048) NOT NULL,
    "revision" VARCHAR(32) NOT NULL,
    "owner" VARCHAR(1024) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "reference" VARCHAR(32) DEFAULT NULL,
    "refer_filter" VARCHAR(1024) DEFAULT NULL,
    "platform_id" VARCHAR(32) DEFAULT '',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "export_to" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "service_name_uniq" UNIQUE ("name", "namespace")
);

-- --------------------------------------------------------
--
-- Data in the conveyor "service"
--
INSERT INTO
    "service" (
        "id",
        "name",
        "namespace",
        "comment",
        "business",
        "token",
        "revision",
        "owner",
        "flag",
        "ctime",
        "mtime"
    )
VALUES
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        'polaris.checker',
        'Polaris',
        'polaris checker service',
        'polaris',
        '7d19c46de327408d8709ee7392b7700b',
        '301b1e9f0bbd47a6b697e26e99dfe012',
        'polaris',
        0,
        '2021-09-06 07:55:07',
        '2021-09-06 07:55:09'
    );

-- --------------------------------------------------------
--
-- Table structure "service_metadata"
--
CREATE TABLE "service_metadata" (
    "id" VARCHAR(32) NOT NULL,
    "mkey" VARCHAR(128) NOT NULL,
    "mvalue" VARCHAR(4096) NOT NULL,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id", "mkey")
);

-- --------------------------------------------------------
--
-- Table structure "owner_service_map"Quickly query all services under an Owner
--
CREATE TABLE "owner_service_map" (
    "id" VARCHAR(32) NOT NULL,
    "owner" VARCHAR(32) NOT NULL,
    "service" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "circuitbreaker_rule"
--
CREATE TABLE "circuitbreaker_rule" (
    "id" VARCHAR(97) NOT NULL,
    "version" VARCHAR(32) NOT NULL DEFAULT 'master',
    "name" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "business" VARCHAR(64) DEFAULT NULL,
    "department" VARCHAR(1024) DEFAULT NULL,
    "comment" VARCHAR(1024) DEFAULT NULL,
    "inbounds" TEXT NOT NULL,
    "outbounds" TEXT NOT NULL,
    "token" VARCHAR(32) NOT NULL,
    "owner" VARCHAR(1024) NOT NULL,
    "revision" VARCHAR(32) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id", "version"),
    CONSTRAINT "circuitbreaker_rule_name_uniq" UNIQUE ("name", "namespace", "version")
);

-- --------------------------------------------------------
--
-- Table structure "circuitbreaker_rule_relation"
--
CREATE TABLE "circuitbreaker_rule_relation" (
    "service_id" VARCHAR(32) NOT NULL,
    "rule_id" VARCHAR(97) NOT NULL,
    "rule_version" VARCHAR(32) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("service_id")
);

-- --------------------------------------------------------
--
-- Table structure "t_ip_config"
--
CREATE TABLE "t_ip_config" (
    "fip" INTEGER NOT NULL,
    "fareaid" INTEGER NOT NULL,
    "fcityid" INTEGER NOT NULL,
    "fidcid" INTEGER NOT NULL,
    "fflag" SMALLINT DEFAULT '0',
    "fstamp" TIMESTAMP NOT NULL,
    "fflow" INTEGER NOT NULL,
    PRIMARY KEY ("fip")
);

-- --------------------------------------------------------
--
-- Table structure "t_policy"
--
CREATE TABLE "t_policy" (
    "fmodid" INTEGER NOT NULL,
    "fdiv" INTEGER NOT NULL,
    "fmod" INTEGER NOT NULL,
    "fflag" SMALLINT DEFAULT '0',
    "fstamp" TIMESTAMP NOT NULL,
    "fflow" INTEGER NOT NULL,
    PRIMARY KEY ("fmodid")
);

-- --------------------------------------------------------
--
-- Table structure "t_route"
--
CREATE TABLE "t_route" (
    "fip" INTEGER NOT NULL,
    "fmodid" INTEGER NOT NULL,
    "fcmdid" INTEGER NOT NULL,
    "fsetid" VARCHAR(32) NOT NULL,
    "fflag" SMALLINT DEFAULT '0',
    "fstamp" TIMESTAMP NOT NULL,
    "fflow" INTEGER NOT NULL,
    PRIMARY KEY ("fip", "fmodid", "fcmdid")
);

-- --------------------------------------------------------
--
-- Table structure "t_section"
--
CREATE TABLE "t_section" (
    "fmodid" INTEGER NOT NULL,
    "ffrom" INTEGER NOT NULL,
    "fto" INTEGER NOT NULL,
    "fxid" INTEGER NOT NULL,
    "fflag" SMALLINT DEFAULT '0',
    "fstamp" TIMESTAMP NOT NULL,
    "fflow" INTEGER NOT NULL,
    PRIMARY KEY ("fmodid", "ffrom", "fto")
);

-- --------------------------------------------------------
--
-- Table structure "start_lock"
--
CREATE TABLE "start_lock" (
    "lock_id" INTEGER NOT NULL,
    "lock_key" VARCHAR(32) NOT NULL,
    "server" VARCHAR(32) NOT NULL,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("lock_id", "lock_key")
);

--
-- Data in the conveyor "start_lock"
--
INSERT INTO
    "start_lock" ("lock_id", "lock_key", "server", "mtime")
VALUES
    (1, 'sz', 'aaa', '2019-12-05 08:35:49');

-- --------------------------------------------------------
--
-- Table structure "cl5_module"
--
CREATE TABLE "cl5_module" (
    "module_id" INTEGER NOT NULL,
    "interface_id" INTEGER NOT NULL,
    "range_num" INTEGER NOT NULL,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("module_id")
);

--
-- Data in the conveyor "cl5_module"
--
INSERT INTO
    cl5_module (module_id, interface_id, range_num)
VALUES
    (3000001, 1, 0);

-- --------------------------------------------------------
--
-- Table structure "config_file"
--
CREATE TABLE "config_file" (
    "id" BIGSERIAL,
    "namespace" VARCHAR(64) NOT NULL,
    "group" VARCHAR(128) NOT NULL DEFAULT '',
    "name" VARCHAR(128) NOT NULL,
    "content" TEXT NOT NULL,
    "format" VARCHAR(16) DEFAULT 'text',
    "comment" VARCHAR(512) DEFAULT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "config_file_uk_file_uniq" UNIQUE ("namespace", "group", "name")
);

-- --------------------------------------------------------
--
-- Table structure "config_file_group"
--
CREATE TABLE "config_file_group" (
    "id" BIGSERIAL,
    "name" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "comment" VARCHAR(512) DEFAULT NULL,
    "owner" VARCHAR(1024) DEFAULT NULL,
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    "business" VARCHAR(64) DEFAULT NULL,
    "department" VARCHAR(1024) DEFAULT NULL,
    "metadata" TEXT,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    PRIMARY KEY ("id"),
    CONSTRAINT "config_file_group_uk_name_uniq" UNIQUE ("namespace", "name")
);

-- --------------------------------------------------------
--
-- Table structure "config_file_release"
--
CREATE TABLE "config_file_release" (
    "id" BIGSERIAL,
    "name" VARCHAR(128) DEFAULT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "group" VARCHAR(128) NOT NULL,
    "file_name" VARCHAR(128) NOT NULL,
    "format" VARCHAR(16) DEFAULT 'text',
    "content" TEXT NOT NULL,
    "comment" VARCHAR(512) DEFAULT NULL,
    "md5" VARCHAR(128) NOT NULL,
    "version" BIGINT NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    "tags" TEXT,
    "active" SMALLINT NOT NULL DEFAULT '0',
    "description" VARCHAR(512) DEFAULT NULL,
    "release_type" VARCHAR(25) NOT NULL DEFAULT '',
    PRIMARY KEY ("id"),
    CONSTRAINT "config_file_release_uk_file_uniq" UNIQUE ("namespace", "group", "file_name", "name")
);

-- --------------------------------------------------------
--
-- Table structure "config_file_release_history"
--
CREATE TABLE "config_file_release_history" (
    "id" BIGSERIAL,
    "name" VARCHAR(64) DEFAULT '',
    "namespace" VARCHAR(64) NOT NULL,
    "group" VARCHAR(128) NOT NULL,
    "file_name" VARCHAR(128) NOT NULL,
    "content" TEXT NOT NULL,
    "format" VARCHAR(16) DEFAULT 'text',
    "comment" VARCHAR(512) DEFAULT NULL,
    "md5" VARCHAR(128) NOT NULL,
    "type" VARCHAR(32) NOT NULL,
    "status" VARCHAR(16) NOT NULL DEFAULT 'success',
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    "tags" TEXT,
    "version" BIGINT,
    "reason" VARCHAR(3000) DEFAULT '',
    "description" VARCHAR(512) DEFAULT NULL,
    PRIMARY KEY ("id")
);

-- --------------------------------------------------------
--
-- Table structure "config_file_tag"
--
CREATE TABLE "config_file_tag" (
    "id" BIGSERIAL,
    "key" VARCHAR(128) NOT NULL,
    "value" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "group" VARCHAR(128) NOT NULL DEFAULT '',
    "file_name" VARCHAR(128) NOT NULL,
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "config_file_tag_uk_tag_uniq" UNIQUE ("key", "value", "namespace", "group", "file_name")
);

CREATE TABLE "user" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "password" VARCHAR(100) NOT NULL,
    "owner" VARCHAR(128) NOT NULL,
    "source" VARCHAR(32) NOT NULL,
    "mobile" VARCHAR(12) NOT NULL DEFAULT '',
    "email" VARCHAR(64) NOT NULL DEFAULT '',
    "token" VARCHAR(255) NOT NULL,
    "token_enable" SMALLINT NOT NULL DEFAULT 1,
    "user_type" INTEGER NOT NULL DEFAULT 20,
    "comment" VARCHAR(255) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "user_uniq_uniq" UNIQUE ("name", "owner")
);

CREATE TABLE "user_group" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "owner" VARCHAR(128) NOT NULL,
    "token" VARCHAR(255) NOT NULL,
    "comment" VARCHAR(255) NOT NULL,
    "token_enable" SMALLINT NOT NULL DEFAULT 1,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "user_group_uniq_uniq" UNIQUE ("name", "owner")
);

CREATE TABLE "user_group_relation" (
    "user_id" VARCHAR(128) NOT NULL,
    "group_id" VARCHAR(128) NOT NULL,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("user_id", "group_id")
);

CREATE TABLE "auth_strategy" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "action" VARCHAR(32) NOT NULL,
    "owner" VARCHAR(128) NOT NULL,
    "comment" VARCHAR(255) NOT NULL,
    "default" SMALLINT NOT NULL DEFAULT '0',
    "source" VARCHAR(32) NOT NULL,
    "revision" VARCHAR(128) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "auth_strategy_uniq_uniq" UNIQUE ("name", "owner")
);

CREATE TABLE "auth_principal" (
    "strategy_id" VARCHAR(128) NOT NULL,
    "principal_id" VARCHAR(128) NOT NULL,
    "principal_role" INTEGER NOT NULL,
    "extend_info" TEXT,
    PRIMARY KEY ("strategy_id", "principal_id", "principal_role")
);

CREATE TABLE "auth_strategy_resource" (
    "strategy_id" VARCHAR(128) NOT NULL,
    "res_type" INTEGER NOT NULL,
    "res_id" VARCHAR(128) NOT NULL,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("strategy_id", "res_type", "res_id")
);

/* 角色数据 */
CREATE TABLE "auth_role" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "owner" VARCHAR(128) NOT NULL,
    "source" VARCHAR(32) NOT NULL,
    "role_type" INTEGER NOT NULL DEFAULT 20,
    "comment" VARCHAR(255) NOT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "auth_role_uniq_uniq" UNIQUE ("name", "owner")
);

/* 角色关联用户/用户组关系表 */
CREATE TABLE "auth_role_principal" (
    "role_id" VARCHAR(128) NOT NULL,
    "principal_id" VARCHAR(128) NOT NULL,
    "principal_role" INTEGER NOT NULL,
    "extend_info" TEXT,
    PRIMARY KEY ("role_id", "principal_id", "principal_role")
);

/* 鉴权策略中的资源标签关联信息 */
CREATE TABLE "auth_strategy_label" (
    "strategy_id" VARCHAR(128) NOT NULL,
    "key" VARCHAR(128) NOT NULL,
    "value" TEXT NOT NULL,
    "compare_type" VARCHAR(128) NOT NULL,
    PRIMARY KEY ("strategy_id", "key")
);

/* 鉴权策略中的资源标签关联信息 */
CREATE TABLE "auth_strategy_function" (
    "strategy_id" VARCHAR(128) NOT NULL,
    "function" VARCHAR(256) NOT NULL,
    PRIMARY KEY ("strategy_id", "function")
);

-- v1.8.0, support client info storage
CREATE TABLE "client" (
    "id" VARCHAR(128) NOT NULL,
    "host" VARCHAR(100) NOT NULL,
    "type" VARCHAR(100) NOT NULL,
    "version" VARCHAR(32) NOT NULL,
    "region" VARCHAR(128) DEFAULT NULL,
    "zone" VARCHAR(128) DEFAULT NULL,
    "campus" VARCHAR(128) DEFAULT NULL,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

CREATE TABLE "client_stat" (
    "client_id" VARCHAR(128) NOT NULL,
    "target" VARCHAR(100) NOT NULL,
    "port" INTEGER NOT NULL,
    "protocol" VARCHAR(100) NOT NULL,
    "path" VARCHAR(128) NOT NULL,
    PRIMARY KEY ("client_id", "target", "port")
);

-- v1.9.0
CREATE TABLE "config_file_template" (
    "id" BIGSERIAL,
    "name" VARCHAR(128) NOT NULL,
    "content" TEXT NOT NULL,
    "format" VARCHAR(16) DEFAULT 'text',
    "comment" VARCHAR(512) DEFAULT NULL,
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT NULL,
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "config_file_template_uk_name_uniq" UNIQUE ("name")
);

INSERT INTO
    "config_file_template" (
        "name",
        "content",
        "format",
        "comment",
        "create_time",
        "create_by",
        "modify_time",
        "modify_by"
    )
VALUES
    (
        'spring-cloud-gateway-braining',
        '{
        "rules":[
            {
                "conditions":[
                    {
                        "key":"${http.query.uid}",
                        "values":["10000"],
                        "operation":"EQUALS"
                    }
                ],
                "labels":[
                    {
                        "key":"env",
                        "value":"green"
                    }
                ]
            }
        ]
    }',
        'json',
        'Spring Cloud Gateway  染色规则',
        NOW(),
        'polaris',
        NOW(),
        'polaris'
    );

-- v1.12.0
CREATE TABLE "routing_config_v2" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(64) NOT NULL DEFAULT '',
    "namespace" VARCHAR(64) NOT NULL DEFAULT '',
    "policy" VARCHAR(64) NOT NULL,
    "config" TEXT,
    "enable" INTEGER NOT NULL DEFAULT 0,
    "revision" VARCHAR(40) NOT NULL,
    "description" VARCHAR(500) NOT NULL DEFAULT '',
    "priority" SMALLINT NOT NULL DEFAULT '0',
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "etime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "extend_info" VARCHAR(1024) DEFAULT '',
    "metadata" TEXT,
    PRIMARY KEY ("id")
);

CREATE TABLE "leader_election" (
    "elect_key" VARCHAR(128) NOT NULL,
    "version" BIGINT NOT NULL DEFAULT 0,
    "leader" VARCHAR(128) NOT NULL,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("elect_key")
);

-- v1.14.0
CREATE TABLE "circuitbreaker_rule_v2" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL DEFAULT '',
    "enable" INTEGER NOT NULL DEFAULT 0,
    "revision" VARCHAR(40) NOT NULL,
    "description" VARCHAR(1024) NOT NULL DEFAULT '',
    "level" INTEGER NOT NULL,
    "src_service" VARCHAR(128) NOT NULL,
    "src_namespace" VARCHAR(64) NOT NULL,
    "dst_service" VARCHAR(128) NOT NULL,
    "dst_namespace" VARCHAR(64) NOT NULL,
    "dst_method" VARCHAR(128) NOT NULL,
    "config" TEXT,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "etime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id")
);

CREATE TABLE "fault_detect_rule" (
    "id" VARCHAR(128) NOT NULL,
    "name" VARCHAR(64) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL DEFAULT 'default',
    "revision" VARCHAR(40) NOT NULL,
    "description" VARCHAR(1024) NOT NULL DEFAULT '',
    "dst_service" VARCHAR(128) NOT NULL,
    "dst_namespace" VARCHAR(64) NOT NULL,
    "dst_method" VARCHAR(128) NOT NULL,
    "config" TEXT,
    "flag" SMALLINT NOT NULL DEFAULT '0',
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id")
);

/* 服务契约表 */
CREATE TABLE "service_contract" (
    "id" VARCHAR(128) NOT NULL,
    "type" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "service" VARCHAR(128) NOT NULL,
    "protocol" VARCHAR(32) NOT NULL,
    "version" VARCHAR(64) NOT NULL,
    "revision" VARCHAR(128) NOT NULL,
    "flag" SMALLINT DEFAULT 0,
    "content" TEXT,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

/* 服务契约中针对单个接口定义的详细信息描述表 */
CREATE TABLE "service_contract_detail" (
    "id" VARCHAR(128) NOT NULL,
    "contract_id" VARCHAR(128) NOT NULL,
    "type" VARCHAR(128) NOT NULL,
    "namespace" VARCHAR(64) NOT NULL,
    "service" VARCHAR(128) NOT NULL,
    "protocol" VARCHAR(32) NOT NULL,
    "version" VARCHAR(64) NOT NULL,
    "method" VARCHAR(32) NOT NULL,
    "path" VARCHAR(128) NOT NULL,
    "source" INTEGER,
    "content" TEXT,
    "revision" VARCHAR(128) NOT NULL,
    "flag" SMALLINT DEFAULT 0,
    "ctime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);

/* 灰度资源 */
CREATE TABLE "gray_resource" (
    "name" VARCHAR(128) NOT NULL,
    "match_rule" TEXT NOT NULL,
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "create_by" VARCHAR(32) DEFAULT '',
    "modify_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "modify_by" VARCHAR(32) DEFAULT '',
    "flag" SMALLINT DEFAULT 0,
    PRIMARY KEY ("name")
);

CREATE TABLE "lane_group" (
    "id" varchar(128) not null,
    "name" varchar(64) not null,
    "rule" text not null,
    "description" varchar(3000),
    "revision" VARCHAR(40) NOT NULL,
    "flag" SMALLINT default 0,
    "ctime" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "metadata" TEXT,
    PRIMARY KEY ("id"),
    CONSTRAINT "lane_group_name_uniq" UNIQUE ("name")
);

CREATE TABLE "lane_rule" (
    "id" varchar(128) not null,
    "name" varchar(64) not null,
    "group_name" varchar(64) not null,
    "rule" text not null,
    "revision" VARCHAR(40) NOT NULL,
    "description" varchar(3000),
    "enable" SMALLINT,
    "flag" SMALLINT default 0,
    "priority" bigint NOT NULL DEFAULT 0,
    "ctime" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "etime" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "mtime" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id"),
    CONSTRAINT "lane_rule_name_uniq" UNIQUE ("group_name", "name")
);


/* 默认资源信息数据插入 */

-- Create a default master account, password is Polarismesh @ 2021
INSERT INTO
    "user" (
        "id",
        "name",
        "password",
        "source",
        "token",
        "token_enable",
        "user_type",
        "comment",
        "mobile",
        "email",
        "owner"
    )
VALUES
    (
        '65e4789a6d5b49669adf1e9e8387549c',
        'polaris',
        '$2a$10$3izWuZtE5SBdAtSZci.gs.iZ2pAn9I8hEqYrC6gwJp1dyjqQnrrum',
        'Polaris',
        'nu/0WRA4EqSR1FagrjRj0fZwPXuGlMpX+zCuWu4uMqy8xr1vRjisSbA25aAC3mtU8MeeRsKhQiDAynUR09I=',
        1,
        20,
        'default polaris admin account',
        '12345678910',
        '12345678910',
        ''
    );

-- Permissions policy inserted into Polaris-Admin
INSERT INTO
    "auth_strategy" (
        "id",
        "name",
        "action",
        "owner",
        "comment",
        "default",
        "source",
        "revision",
        "flag",
        "ctime",
        "mtime"
    )
VALUES
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        '(用户) polaris的默认策略',
        'READ_WRITE',
        '65e4789a6d5b49669adf1e9e8387549c',
        'default admin',
        1,
        'Polaris',
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        0,
        now(),
        now()
    );

-- Sport rules inserted into Polaris-Admin to access
INSERT INTO
    "auth_strategy_resource" (
        "strategy_id",
        "res_type",
        "res_id",
        "ctime",
        "mtime"
    )
VALUES
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        0,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        1,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        2,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        3,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        4,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        5,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        6,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        7,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        20,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        21,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        22,
        '*',
        now(),
        now()
    ),
    (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        23,
        '*',
        now(),
        now()
    );

-- Insert permission policies and association relationships for Polaris-Admin accounts
INSERT INTO
    auth_principal ("strategy_id", "principal_id", "principal_role") VALUES (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        '65e4789a6d5b49669adf1e9e8387549c',
        1
    );

INSERT INTO
    auth_strategy_function ("strategy_id", "function") VALUES (
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        '*'
    );

/* 默认的全局只读策略 */
INSERT INTO
    "auth_strategy" (
        "id",
        "name",
        "action",
        "owner",
        "comment",
        "default",
        "source",
        "revision",
        "flag",
        "ctime",
        "mtime"
    )
VALUES
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        '全局只读策略',
        'ALLOW',
        '65e4789a6d5b49669adf1e9e8387549c',
        'global resources read onyly',
        1,
        'Polaris',
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        0,
        now(),
        now()
    );

INSERT INTO
    "auth_strategy_resource" (
        "strategy_id",
        "res_type",
        "res_id",
        "ctime",
        "mtime"
    )
VALUES
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        0,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        1,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        2,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        3,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        4,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        5,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        6,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        7,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        20,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        21,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        22,
        '*',
        now(),
        now()
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        23,
        '*',
        now(),
        now()
    );

INSERT INTO
    auth_strategy_function ("strategy_id", "function") VALUES (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        'Describe*'
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        'List*'
    ),
    (
        'bfa04ae1e32a94fbca9ead86e1ecf581',
        'Get*'
    );


/* 默认的全局读写策略 */
INSERT INTO
    "auth_strategy" (
        "id",
        "name",
        "action",
        "owner",
        "comment",
        "default",
        "source",
        "revision",
        "flag",
        "ctime",
        "mtime"
    )
VALUES
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        '全局读写策略',
        'ALLOW',
        '65e4789a6d5b49669adf1e9e8387549c',
        'global resources read and write',
        1,
        'Polaris',
        'fbca9bfa04ae4ead86e1ecf5811e32a9',
        0,
        now(),
        now()
    );

INSERT INTO
    "auth_strategy_resource" (
        "strategy_id",
        "res_type",
        "res_id",
        "ctime",
        "mtime"
    )
VALUES
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        0,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        1,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        2,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        3,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        4,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        5,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        6,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        7,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        20,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        21,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        22,
        '*',
        now(),
        now()
    ),
    (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        23,
        '*',
        now(),
        now()
    );

INSERT INTO
    auth_strategy_function ("strategy_id", "function") VALUES (
        'e3d86e1ecf5812bfa04ae1a94fbca9ea',
        '*'
    );


--
-- 索引
--
CREATE INDEX "instance_service_id_idx" ON "instance" ("service_id");
CREATE INDEX "instance_mtime_idx" ON "instance" ("mtime");
CREATE INDEX "instance_host_idx" ON "instance" ("host");
CREATE INDEX "instance_metadata_mkey_idx" ON "instance_metadata" ("mkey");
CREATE INDEX "routing_config_mtime_idx" ON "routing_config" ("mtime");
CREATE INDEX "ratelimit_config_mtime_idx" ON "ratelimit_config" ("mtime");
CREATE INDEX "ratelimit_config_service_id_idx" ON "ratelimit_config" ("service_id");
CREATE INDEX "ratelimit_revision_service_id_idx" ON "ratelimit_revision" ("service_id");
CREATE INDEX "ratelimit_revision_mtime_idx" ON "ratelimit_revision" ("mtime");
CREATE INDEX "service_namespace_idx" ON "service" ("namespace");
CREATE INDEX "service_mtime_idx" ON "service" ("mtime");
CREATE INDEX "service_reference_idx" ON "service" ("reference");
CREATE INDEX "service_platform_id_idx" ON "service" ("platform_id");
CREATE INDEX "service_metadata_mkey_idx" ON "service_metadata" ("mkey");
CREATE INDEX "owner_service_map_owner_idx" ON "owner_service_map" ("owner");
CREATE INDEX "owner_service_map_name_idx" ON "owner_service_map" ("service", "namespace");
CREATE INDEX "circuitbreaker_rule_mtime_idx" ON "circuitbreaker_rule" ("mtime");
CREATE INDEX "circuitbreaker_rule_relation_mtime_idx" ON "circuitbreaker_rule_relation" ("mtime");
CREATE INDEX "circuitbreaker_rule_relation_rule_id_idx" ON "circuitbreaker_rule_relation" ("rule_id");
CREATE INDEX "t_ip_config_idx_fflow_idx" ON "t_ip_config" ("fflow");
CREATE INDEX "t_route_fflow_idx" ON "t_route" ("fflow");
CREATE INDEX "t_route_idx1_idx" ON "t_route" ("fmodid", "fcmdid", "fsetid");
CREATE INDEX "config_file_release_idx_modify_time_idx" ON "config_file_release" ("modify_time");
CREATE INDEX "config_file_release_history_idx_file_idx" ON "config_file_release_history" ("namespace", "group", "file_name");
CREATE INDEX "config_file_tag_idx_file_idx" ON "config_file_tag" ("namespace", "group", "file_name");
CREATE INDEX "user_owner_idx" ON "user" ("owner");
CREATE INDEX "user_mtime_idx" ON "user" ("mtime");
CREATE INDEX "user_group_owner_idx" ON "user_group" ("owner");
CREATE INDEX "user_group_mtime_idx" ON "user_group" ("mtime");
CREATE INDEX "user_group_relation_mtime_idx" ON "user_group_relation" ("mtime");
CREATE INDEX "auth_strategy_owner_idx" ON "auth_strategy" ("owner");
CREATE INDEX "auth_strategy_mtime_idx" ON "auth_strategy" ("mtime");
CREATE INDEX "auth_strategy_resource_mtime_idx" ON "auth_strategy_resource" ("mtime");
CREATE INDEX "auth_role_owner_idx" ON "auth_role" ("owner");
CREATE INDEX "auth_role_mtime_idx" ON "auth_role" ("mtime");
CREATE INDEX "client_mtime_idx" ON "client" ("mtime");
CREATE INDEX "routing_config_v2_mtime_idx" ON "routing_config_v2" ("mtime");
CREATE INDEX "leader_election_version_idx" ON "leader_election" ("version");
CREATE INDEX "circuitbreaker_rule_v2_name_idx" ON "circuitbreaker_rule_v2" ("name");
CREATE INDEX "circuitbreaker_rule_v2_mtime_idx" ON "circuitbreaker_rule_v2" ("mtime");
CREATE INDEX "fault_detect_rule_name_idx" ON "fault_detect_rule" ("name");
CREATE INDEX "fault_detect_rule_mtime_idx" ON "fault_detect_rule" ("mtime");
CREATE INDEX "service_contract_namespace_service_type_version_protocol_idx" ON "service_contract" ("namespace", "service", "type", "version", "protocol");
CREATE INDEX "service_contract_detail_contract_id_path_method_source_idx" ON "service_contract_detail" ("contract_id", "path", "method", "source");

--
-- 注释
--
COMMENT ON COLUMN "instance"."id" IS 'Unique ID';
COMMENT ON COLUMN "instance"."service_id" IS 'Service ID';
COMMENT ON COLUMN "instance"."vpc_id" IS 'VPC ID';
COMMENT ON COLUMN "instance"."host" IS 'instance Host Information';
COMMENT ON COLUMN "instance"."port" IS 'instance port information';
COMMENT ON COLUMN "instance"."protocol" IS 'Listening protocols for corresponding ports, such as TPC, UDP, GRPC, DUBBO, etc.';
COMMENT ON COLUMN "instance"."version" IS 'The version of the instance can be used for version routing';
COMMENT ON COLUMN "instance"."health_status" IS 'The health status of the instance, 1 is health, 0 is unhealthy';
COMMENT ON COLUMN "instance"."isolate" IS 'Example isolation status flag, 0 is not isolated, 1 is isolated';
COMMENT ON COLUMN "instance"."weight" IS 'The weight of the instance is mainly used for LoadBalance, default is 100';
COMMENT ON COLUMN "instance"."enable_health_check" IS 'Whether to open a heartbeat on an instance, check the logic, 0 is not open, 1 is open';
COMMENT ON COLUMN "instance"."logic_set" IS 'Example logic packet information';
COMMENT ON COLUMN "instance"."cmdb_region" IS 'The region information of the instance is mainly used to close the route';
COMMENT ON COLUMN "instance"."cmdb_zone" IS 'The ZONE information of the instance is mainly used to close the route.';
COMMENT ON COLUMN "instance"."cmdb_idc" IS 'The IDC information of the instance is mainly used to close the route';
COMMENT ON COLUMN "instance"."priority" IS 'Example priority, currently useless';
COMMENT ON COLUMN "instance"."revision" IS 'Instance version information';
COMMENT ON COLUMN "instance"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "instance"."ctime" IS 'Create time';
COMMENT ON COLUMN "instance"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "health_check"."id" IS 'Instance ID';
COMMENT ON COLUMN "health_check"."type" IS 'Instance health check type';
COMMENT ON COLUMN "health_check"."ttl" IS 'TTL time jumping';
COMMENT ON COLUMN "instance_metadata"."id" IS 'Instance ID';
COMMENT ON COLUMN "instance_metadata"."mkey" IS 'instance label of Key';
COMMENT ON COLUMN "instance_metadata"."mvalue" IS 'instance label Value';
COMMENT ON COLUMN "instance_metadata"."ctime" IS 'Create time';
COMMENT ON COLUMN "instance_metadata"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "namespace"."name" IS 'Namespace name, unique';
COMMENT ON COLUMN "namespace"."comment" IS 'Description of namespace';
COMMENT ON COLUMN "namespace"."token" IS 'TOKEN named space for write operation check';
COMMENT ON COLUMN "namespace"."owner" IS 'Responsible for named space Owner';
COMMENT ON COLUMN "namespace"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "namespace"."ctime" IS 'Create time';
COMMENT ON COLUMN "namespace"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "namespace"."service_export_to" IS 'namespace metadata';
COMMENT ON COLUMN "namespace"."metadata" IS 'namespace metadata';
COMMENT ON COLUMN "routing_config"."id" IS 'Routing configuration ID';
COMMENT ON COLUMN "routing_config"."in_bounds" IS 'Service is routing rules';
COMMENT ON COLUMN "routing_config"."out_bounds" IS 'Service main routing rules';
COMMENT ON COLUMN "routing_config"."revision" IS 'Routing rule version';
COMMENT ON COLUMN "routing_config"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "routing_config"."ctime" IS 'Create time';
COMMENT ON COLUMN "routing_config"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "ratelimit_config"."id" IS 'ratelimit rule ID';
COMMENT ON COLUMN "ratelimit_config"."name" IS 'ratelimt rule name';
COMMENT ON COLUMN "ratelimit_config"."disable" IS 'ratelimit disable';
COMMENT ON COLUMN "ratelimit_config"."service_id" IS 'Service ID';
COMMENT ON COLUMN "ratelimit_config"."method" IS 'ratelimit method';
COMMENT ON COLUMN "ratelimit_config"."labels" IS 'Conductive flow for a specific label';
COMMENT ON COLUMN "ratelimit_config"."priority" IS 'ratelimit rule priority';
COMMENT ON COLUMN "ratelimit_config"."rule" IS 'Current limiting rules';
COMMENT ON COLUMN "ratelimit_config"."revision" IS 'Limiting version';
COMMENT ON COLUMN "ratelimit_config"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "ratelimit_config"."ctime" IS 'Create time';
COMMENT ON COLUMN "ratelimit_config"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "ratelimit_config"."etime" IS 'RateLimit rule enable time';
COMMENT ON COLUMN "ratelimit_config"."metadata" IS 'ratelimit rule metadata';
COMMENT ON COLUMN "ratelimit_revision"."service_id" IS 'Service ID';
COMMENT ON COLUMN "ratelimit_revision"."last_revision" IS 'The latest limited limiting rule version of the corresponding service';
COMMENT ON COLUMN "ratelimit_revision"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "service"."id" IS 'Service ID';
COMMENT ON COLUMN "service"."name" IS 'Service name, only under the namespace';
COMMENT ON COLUMN "service"."namespace" IS 'Namespace belongs to the service';
COMMENT ON COLUMN "service"."ports" IS 'Service will have a list of all port information of the external exposure (single process exposing multiple protocols)';
COMMENT ON COLUMN "service"."business" IS 'Service business information';
COMMENT ON COLUMN "service"."department" IS 'Service department information';
COMMENT ON COLUMN "service"."comment" IS 'Description information';
COMMENT ON COLUMN "service"."token" IS 'Service token, used to handle all the services involved in the service';
COMMENT ON COLUMN "service"."revision" IS 'Service version information';
COMMENT ON COLUMN "service"."owner" IS 'Owner information belonging to the service';
COMMENT ON COLUMN "service"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "service"."reference" IS 'Service alias, what is the actual service name that the service is actually pointed out?';
COMMENT ON COLUMN "service"."platform_id" IS 'The platform ID to which the service belongs';
COMMENT ON COLUMN "service"."ctime" IS 'Create time';
COMMENT ON COLUMN "service"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "service"."export_to" IS 'service export to some namespace';
COMMENT ON COLUMN "service_metadata"."id" IS 'Service ID';
COMMENT ON COLUMN "service_metadata"."mkey" IS 'Service label key';
COMMENT ON COLUMN "service_metadata"."mvalue" IS 'Service label Value';
COMMENT ON COLUMN "service_metadata"."ctime" IS 'Create time';
COMMENT ON COLUMN "service_metadata"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "owner_service_map"."owner" IS 'Service Owner';
COMMENT ON COLUMN "owner_service_map"."service" IS 'service name';
COMMENT ON COLUMN "owner_service_map"."namespace" IS 'namespace name';
COMMENT ON COLUMN "circuitbreaker_rule"."id" IS 'Melting rule ID';
COMMENT ON COLUMN "circuitbreaker_rule"."version" IS 'Melting rule version, default is MASTR';
COMMENT ON COLUMN "circuitbreaker_rule"."name" IS 'Melting rule name';
COMMENT ON COLUMN "circuitbreaker_rule"."namespace" IS 'Melting rule belongs to name space';
COMMENT ON COLUMN "circuitbreaker_rule"."business" IS 'Business information of fuse regular';
COMMENT ON COLUMN "circuitbreaker_rule"."department" IS 'Department information to which the fuse regular belongs';
COMMENT ON COLUMN "circuitbreaker_rule"."comment" IS 'Description of the fuse rule';
COMMENT ON COLUMN "circuitbreaker_rule"."inbounds" IS 'Service-tuned fuse rule';
COMMENT ON COLUMN "circuitbreaker_rule"."outbounds" IS 'Service Motoring Fuse Rule';
COMMENT ON COLUMN "circuitbreaker_rule"."token" IS 'Token, which is fucking, mainly for writing operation check';
COMMENT ON COLUMN "circuitbreaker_rule"."owner" IS 'Melting rule Owner information';
COMMENT ON COLUMN "circuitbreaker_rule"."revision" IS 'Melt rule version information';
COMMENT ON COLUMN "circuitbreaker_rule"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "circuitbreaker_rule"."ctime" IS 'Create time';
COMMENT ON COLUMN "circuitbreaker_rule"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "circuitbreaker_rule"."metadata" IS 'circuit_breaker rule metadata';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."service_id" IS 'Service ID';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."rule_id" IS 'Melting rule ID';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."rule_version" IS 'Melting rule version';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."flag" IS 'Logic delete flag, 0 means visible, 1 means that it has been logically deleted';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."ctime" IS 'Create time';
COMMENT ON COLUMN "circuitbreaker_rule_relation"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "t_ip_config"."fip" IS 'Machine IP';
COMMENT ON COLUMN "t_ip_config"."fareaid" IS 'Area number';
COMMENT ON COLUMN "t_ip_config"."fcityid" IS 'City number';
COMMENT ON COLUMN "t_ip_config"."fidcid" IS 'IDC number';
COMMENT ON COLUMN "start_lock"."lock_id" IS '锁序号';
COMMENT ON COLUMN "start_lock"."lock_key" IS 'Lock name';
COMMENT ON COLUMN "start_lock"."server" IS 'SERVER holding launch lock';
COMMENT ON COLUMN "start_lock"."mtime" IS 'Update time';
COMMENT ON TABLE "cl5_module" IS 'To generate SID';
COMMENT ON COLUMN "cl5_module"."module_id" IS 'Module ID';
COMMENT ON COLUMN "cl5_module"."interface_id" IS 'Interface ID';
COMMENT ON COLUMN "cl5_module"."mtime" IS 'Last updated time';
COMMENT ON TABLE "config_file" IS '配置文件表';
COMMENT ON COLUMN "config_file"."id" IS '主键';
COMMENT ON COLUMN "config_file"."namespace" IS '所属的namespace';
COMMENT ON COLUMN "config_file"."group" IS '所属的文件组';
COMMENT ON COLUMN "config_file"."name" IS '配置文件名';
COMMENT ON COLUMN "config_file"."content" IS '文件内容';
COMMENT ON COLUMN "config_file"."format" IS '文件格式，枚举值';
COMMENT ON COLUMN "config_file"."comment" IS '备注信息';
COMMENT ON COLUMN "config_file"."flag" IS '软删除标记位';
COMMENT ON COLUMN "config_file"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file"."modify_by" IS '最后更新人';
COMMENT ON TABLE "config_file_group" IS '配置文件组表';
COMMENT ON COLUMN "config_file_group"."id" IS '主键';
COMMENT ON COLUMN "config_file_group"."name" IS '配置文件分组名';
COMMENT ON COLUMN "config_file_group"."namespace" IS '所属的namespace';
COMMENT ON COLUMN "config_file_group"."comment" IS '备注信息';
COMMENT ON COLUMN "config_file_group"."owner" IS '负责人';
COMMENT ON COLUMN "config_file_group"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file_group"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file_group"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file_group"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "config_file_group"."business" IS 'Service business information';
COMMENT ON COLUMN "config_file_group"."department" IS 'Service department information';
COMMENT ON COLUMN "config_file_group"."metadata" IS '配置分组标签';
COMMENT ON COLUMN "config_file_group"."flag" IS '是否被删除';
COMMENT ON TABLE "config_file_release" IS '配置文件发布表';
COMMENT ON COLUMN "config_file_release"."id" IS '主键';
COMMENT ON COLUMN "config_file_release"."name" IS '发布标题';
COMMENT ON COLUMN "config_file_release"."namespace" IS '所属的namespace';
COMMENT ON COLUMN "config_file_release"."group" IS '所属的文件组';
COMMENT ON COLUMN "config_file_release"."file_name" IS '配置文件名';
COMMENT ON COLUMN "config_file_release"."format" IS '文件格式，枚举值';
COMMENT ON COLUMN "config_file_release"."content" IS '文件内容';
COMMENT ON COLUMN "config_file_release"."comment" IS '备注信息';
COMMENT ON COLUMN "config_file_release"."md5" IS 'content的md5值';
COMMENT ON COLUMN "config_file_release"."version" IS '版本号，每次发布自增1';
COMMENT ON COLUMN "config_file_release"."flag" IS '是否被删除';
COMMENT ON COLUMN "config_file_release"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file_release"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file_release"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file_release"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "config_file_release"."tags" IS '文件标签';
COMMENT ON COLUMN "config_file_release"."active" IS '是否处于使用中';
COMMENT ON COLUMN "config_file_release"."description" IS '发布描述';
COMMENT ON COLUMN "config_file_release"."release_type" IS '文件类型：""：全量 gray：灰度';
COMMENT ON TABLE "config_file_release_history" IS '配置文件发布历史表';
COMMENT ON COLUMN "config_file_release_history"."id" IS '主键';
COMMENT ON COLUMN "config_file_release_history"."name" IS '发布名称';
COMMENT ON COLUMN "config_file_release_history"."namespace" IS '所属的namespace';
COMMENT ON COLUMN "config_file_release_history"."group" IS '所属的文件组';
COMMENT ON COLUMN "config_file_release_history"."file_name" IS '配置文件名';
COMMENT ON COLUMN "config_file_release_history"."content" IS '文件内容';
COMMENT ON COLUMN "config_file_release_history"."format" IS '文件格式';
COMMENT ON COLUMN "config_file_release_history"."comment" IS '备注信息';
COMMENT ON COLUMN "config_file_release_history"."md5" IS 'content的md5值';
COMMENT ON COLUMN "config_file_release_history"."type" IS '发布类型，例如全量发布、灰度发布';
COMMENT ON COLUMN "config_file_release_history"."status" IS '发布状态，success表示成功，fail 表示失败';
COMMENT ON COLUMN "config_file_release_history"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file_release_history"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file_release_history"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file_release_history"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "config_file_release_history"."tags" IS '文件标签';
COMMENT ON COLUMN "config_file_release_history"."version" IS '版本号，每次发布自增1';
COMMENT ON COLUMN "config_file_release_history"."reason" IS '原因';
COMMENT ON COLUMN "config_file_release_history"."description" IS '发布描述';
COMMENT ON TABLE "config_file_tag" IS '配置文件标签表';
COMMENT ON COLUMN "config_file_tag"."id" IS '主键';
COMMENT ON COLUMN "config_file_tag"."key" IS 'tag 的键';
COMMENT ON COLUMN "config_file_tag"."value" IS 'tag 的值';
COMMENT ON COLUMN "config_file_tag"."namespace" IS '所属的namespace';
COMMENT ON COLUMN "config_file_tag"."group" IS '所属的文件组';
COMMENT ON COLUMN "config_file_tag"."file_name" IS '配置文件名';
COMMENT ON COLUMN "config_file_tag"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file_tag"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file_tag"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file_tag"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "user"."id" IS 'User ID';
COMMENT ON COLUMN "user"."name" IS 'user name';
COMMENT ON COLUMN "user"."password" IS 'user password';
COMMENT ON COLUMN "user"."owner" IS 'Main account ID';
COMMENT ON COLUMN "user"."source" IS 'Account source';
COMMENT ON COLUMN "user"."mobile" IS 'Account mobile phone number';
COMMENT ON COLUMN "user"."email" IS 'Account mailbox';
COMMENT ON COLUMN "user"."token" IS 'The token information owned by the account can be used for SDK access authentication';
COMMENT ON COLUMN "user"."user_type" IS 'Account type, 0 is the admin super account, 20 is the primary account, 50 for the child account';
COMMENT ON COLUMN "user"."comment" IS 'describe';
COMMENT ON COLUMN "user"."flag" IS 'Whether the rules are valid, 0 is valid, 1 is invalid, it is deleted';
COMMENT ON COLUMN "user"."ctime" IS 'Create time';
COMMENT ON COLUMN "user"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "user"."metadata" IS 'user metadata';
COMMENT ON COLUMN "user_group"."id" IS 'User group ID';
COMMENT ON COLUMN "user_group"."name" IS 'User group name';
COMMENT ON COLUMN "user_group"."owner" IS 'The main account ID of the user group';
COMMENT ON COLUMN "user_group"."token" IS 'TOKEN information of this user group';
COMMENT ON COLUMN "user_group"."comment" IS 'Description';
COMMENT ON COLUMN "user_group"."flag" IS 'Whether the rules are valid, 0 is valid, 1 is invalid, it is deleted';
COMMENT ON COLUMN "user_group"."ctime" IS 'Create time';
COMMENT ON COLUMN "user_group"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "user_group"."metadata" IS 'user_group metadata';
COMMENT ON COLUMN "user_group_relation"."user_id" IS 'User ID';
COMMENT ON COLUMN "user_group_relation"."group_id" IS 'User group ID';
COMMENT ON COLUMN "user_group_relation"."ctime" IS 'Create time';
COMMENT ON COLUMN "user_group_relation"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "auth_strategy"."id" IS 'Strategy ID';
COMMENT ON COLUMN "auth_strategy"."name" IS 'Policy name';
COMMENT ON COLUMN "auth_strategy"."action" IS 'Read and write permission for this policy, only_read = 0, read_write = 1';
COMMENT ON COLUMN "auth_strategy"."owner" IS 'The account ID to which this policy is';
COMMENT ON COLUMN "auth_strategy"."comment" IS 'describe';
COMMENT ON COLUMN "auth_strategy"."source" IS 'policy rule source';
COMMENT ON COLUMN "auth_strategy"."revision" IS 'Authentication rule version';
COMMENT ON COLUMN "auth_strategy"."flag" IS 'Whether the rules are valid, 0 is valid, 1 is invalid, it is deleted';
COMMENT ON COLUMN "auth_strategy"."ctime" IS 'Create time';
COMMENT ON COLUMN "auth_strategy"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "auth_strategy"."metadata" IS 'policy rule metadata';
COMMENT ON COLUMN "auth_principal"."strategy_id" IS 'Strategy ID';
COMMENT ON COLUMN "auth_principal"."principal_id" IS 'Principal ID';
COMMENT ON COLUMN "auth_principal"."principal_role" IS 'PRINCIPAL type, 1 is User, 2 is Group, 3 is Role';
COMMENT ON COLUMN "auth_principal"."extend_info" IS 'link principal extend info';
COMMENT ON COLUMN "auth_strategy_resource"."strategy_id" IS 'Strategy ID';
COMMENT ON COLUMN "auth_strategy_resource"."res_type" IS 'Resource Type, Namespaces = 0, Service = 1, configgroups = 2';
COMMENT ON COLUMN "auth_strategy_resource"."res_id" IS 'Resource ID';
COMMENT ON COLUMN "auth_strategy_resource"."ctime" IS 'Create time';
COMMENT ON COLUMN "auth_strategy_resource"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "auth_role"."id" IS 'role id';
COMMENT ON COLUMN "auth_role"."name" IS 'role name';
COMMENT ON COLUMN "auth_role"."owner" IS 'Main account ID';
COMMENT ON COLUMN "auth_role"."source" IS 'role source';
COMMENT ON COLUMN "auth_role"."role_type" IS 'role type';
COMMENT ON COLUMN "auth_role"."comment" IS 'describe';
COMMENT ON COLUMN "auth_role"."flag" IS 'Whether the rules are valid, 0 is valid, 1 is invalid, it is deleted';
COMMENT ON COLUMN "auth_role"."ctime" IS 'Create time';
COMMENT ON COLUMN "auth_role"."mtime" IS 'Last updated time';
COMMENT ON COLUMN "auth_role"."metadata" IS 'user metadata';
COMMENT ON COLUMN "auth_role_principal"."role_id" IS 'role id';
COMMENT ON COLUMN "auth_role_principal"."principal_id" IS 'principal id';
COMMENT ON COLUMN "auth_role_principal"."principal_role" IS 'PRINCIPAL type, 1 is User, 2 is Group';
COMMENT ON COLUMN "auth_role_principal"."extend_info" IS 'link principal extend info';
COMMENT ON COLUMN "auth_strategy_label"."strategy_id" IS 'strategy id';
COMMENT ON COLUMN "auth_strategy_label"."key" IS 'tag key';
COMMENT ON COLUMN "auth_strategy_label"."value" IS 'tag value';
COMMENT ON COLUMN "auth_strategy_label"."compare_type" IS 'tag kv compare func';
COMMENT ON COLUMN "auth_strategy_function"."strategy_id" IS 'strategy id';
COMMENT ON COLUMN "auth_strategy_function"."function" IS 'server provider function name';
COMMENT ON COLUMN "client"."id" IS 'client id';
COMMENT ON COLUMN "client"."host" IS 'client host IP';
COMMENT ON COLUMN "client"."type" IS 'client type: polaris-java/polaris-go';
COMMENT ON COLUMN "client"."version" IS 'client SDK version';
COMMENT ON COLUMN "client"."region" IS 'region info for client';
COMMENT ON COLUMN "client"."zone" IS 'zone info for client';
COMMENT ON COLUMN "client"."campus" IS 'campus info for client';
COMMENT ON COLUMN "client"."flag" IS '0 is valid, 1 is invalid(deleted)';
COMMENT ON COLUMN "client"."ctime" IS 'create time';
COMMENT ON COLUMN "client"."mtime" IS 'last updated time';
COMMENT ON COLUMN "client_stat"."client_id" IS 'client id';
COMMENT ON COLUMN "client_stat"."target" IS 'target stat platform';
COMMENT ON COLUMN "client_stat"."port" IS 'client port to get stat information';
COMMENT ON COLUMN "client_stat"."protocol" IS 'stat info transport protocol';
COMMENT ON COLUMN "client_stat"."path" IS 'stat metric path';
COMMENT ON TABLE "config_file_template" IS '配置文件模板表';
COMMENT ON COLUMN "config_file_template"."id" IS '主键';
COMMENT ON COLUMN "config_file_template"."name" IS '配置文件模板名称';
COMMENT ON COLUMN "config_file_template"."content" IS '配置文件模板内容';
COMMENT ON COLUMN "config_file_template"."format" IS '模板文件格式';
COMMENT ON COLUMN "config_file_template"."comment" IS '模板描述信息';
COMMENT ON COLUMN "config_file_template"."create_time" IS '创建时间';
COMMENT ON COLUMN "config_file_template"."create_by" IS '创建人';
COMMENT ON COLUMN "config_file_template"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "config_file_template"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "routing_config_v2"."priority" IS 'ratelimit rule priority';
COMMENT ON COLUMN "routing_config_v2"."metadata" IS 'route rule metadata';
COMMENT ON COLUMN "circuitbreaker_rule_v2"."metadata" IS 'circuit_breaker rule metadata';
COMMENT ON COLUMN "fault_detect_rule"."metadata" IS 'faultdetect rule metadata';
COMMENT ON COLUMN "service_contract"."id" IS '服务契约主键';
COMMENT ON COLUMN "service_contract"."type" IS '服务契约名称';
COMMENT ON COLUMN "service_contract"."namespace" IS '命名空间';
COMMENT ON COLUMN "service_contract"."service" IS '服务名称';
COMMENT ON COLUMN "service_contract"."protocol" IS '当前契约对应的协议信息 e.g. http/dubbo/grpc/thrift';
COMMENT ON COLUMN "service_contract"."version" IS '服务契约版本';
COMMENT ON COLUMN "service_contract"."revision" IS '当前服务契约的全部内容版本摘要';
COMMENT ON COLUMN "service_contract"."flag" IS '逻辑删除标志位 ， 0 位有效 ， 1 为逻辑删除';
COMMENT ON COLUMN "service_contract"."content" IS '描述信息';
COMMENT ON COLUMN "service_contract_detail"."id" IS '服务契约单个接口定义记录主键';
COMMENT ON COLUMN "service_contract_detail"."contract_id" IS '服务契约 ID';
COMMENT ON COLUMN "service_contract_detail"."type" IS '服务契约接口名称';
COMMENT ON COLUMN "service_contract_detail"."namespace" IS '命名空间';
COMMENT ON COLUMN "service_contract_detail"."service" IS '服务名称';
COMMENT ON COLUMN "service_contract_detail"."protocol" IS '当前契约对应的协议信息 e.g. http/dubbo/grpc/thrift';
COMMENT ON COLUMN "service_contract_detail"."version" IS '服务契约版本';
COMMENT ON COLUMN "service_contract_detail"."method" IS 'http协议中的 method 字段, eg:POST/GET/PUT/DELETE, 其他 gRPC 可以用来标识 stream 类型';
COMMENT ON COLUMN "service_contract_detail"."path" IS '接口具体全路径描述';
COMMENT ON COLUMN "service_contract_detail"."source" IS '该条记录来源, 0:SDK/1:MANUAL';
COMMENT ON COLUMN "service_contract_detail"."content" IS '描述信息';
COMMENT ON COLUMN "service_contract_detail"."revision" IS '当前接口定义的全部内容版本摘要';
COMMENT ON COLUMN "service_contract_detail"."flag" IS '逻辑删除标志位, 0 位有效, 1 为逻辑删除';
COMMENT ON TABLE "gray_resource" IS '灰度资源表';
COMMENT ON COLUMN "gray_resource"."name" IS '灰度资源';
COMMENT ON COLUMN "gray_resource"."match_rule" IS '配置规则';
COMMENT ON COLUMN "gray_resource"."create_time" IS '创建时间';
COMMENT ON COLUMN "gray_resource"."create_by" IS '创建人';
COMMENT ON COLUMN "gray_resource"."modify_time" IS '最后更新时间';
COMMENT ON COLUMN "gray_resource"."modify_by" IS '最后更新人';
COMMENT ON COLUMN "gray_resource"."flag" IS '逻辑删除标志位, 0 位有效, 1 为逻辑删除';
COMMENT ON COLUMN "lane_group"."id" IS '泳道分组 ID';
COMMENT ON COLUMN "lane_group"."name" IS '泳道分组名称';
COMMENT ON COLUMN "lane_group"."rule" IS '规则的 json 字符串';
COMMENT ON COLUMN "lane_group"."description" IS '规则描述';
COMMENT ON COLUMN "lane_group"."revision" IS '规则摘要';
COMMENT ON COLUMN "lane_group"."flag" IS '软删除标识位';
COMMENT ON COLUMN "lane_group"."metadata" IS 'lane rule metadata';
COMMENT ON COLUMN "lane_rule"."id" IS '规则 id';
COMMENT ON COLUMN "lane_rule"."name" IS '规则名称';
COMMENT ON COLUMN "lane_rule"."group_name" IS '泳道分组名称';
COMMENT ON COLUMN "lane_rule"."rule" IS '规则的 json 字符串';
COMMENT ON COLUMN "lane_rule"."revision" IS '规则摘要';
COMMENT ON COLUMN "lane_rule"."description" IS '规则描述';
COMMENT ON COLUMN "lane_rule"."enable" IS '是否启用';
COMMENT ON COLUMN "lane_rule"."flag" IS '软删除标识位';
COMMENT ON COLUMN "lane_rule"."priority" IS '泳道规则优先级';

--
-- 自动更新修改时间的触发器（替代 MySQL 的 ON UPDATE CURRENT_TIMESTAMP）
--
CREATE OR REPLACE FUNCTION polaris_set_modify_time()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modify_time = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION polaris_set_mtime()
RETURNS TRIGGER AS $$
BEGIN
    NEW.mtime = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_instance_mtime BEFORE UPDATE ON "instance"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_instance_metadata_mtime BEFORE UPDATE ON "instance_metadata"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_namespace_mtime BEFORE UPDATE ON "namespace"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_routing_config_mtime BEFORE UPDATE ON "routing_config"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_ratelimit_config_mtime BEFORE UPDATE ON "ratelimit_config"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_ratelimit_revision_mtime BEFORE UPDATE ON "ratelimit_revision"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_service_mtime BEFORE UPDATE ON "service"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_service_metadata_mtime BEFORE UPDATE ON "service_metadata"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_circuitbreaker_rule_mtime BEFORE UPDATE ON "circuitbreaker_rule"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_circuitbreaker_rule_relation_mtime BEFORE UPDATE ON "circuitbreaker_rule_relation"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_start_lock_mtime BEFORE UPDATE ON "start_lock"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_cl5_module_mtime BEFORE UPDATE ON "cl5_module"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_config_file_modify_time BEFORE UPDATE ON "config_file"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_config_file_group_modify_time BEFORE UPDATE ON "config_file_group"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_config_file_release_modify_time BEFORE UPDATE ON "config_file_release"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_config_file_release_history_modify_time BEFORE UPDATE ON "config_file_release_history"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_config_file_tag_modify_time BEFORE UPDATE ON "config_file_tag"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_user_mtime BEFORE UPDATE ON "user"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_user_group_mtime BEFORE UPDATE ON "user_group"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_user_group_relation_mtime BEFORE UPDATE ON "user_group_relation"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_auth_strategy_mtime BEFORE UPDATE ON "auth_strategy"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_auth_strategy_resource_mtime BEFORE UPDATE ON "auth_strategy_resource"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_auth_role_mtime BEFORE UPDATE ON "auth_role"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_client_mtime BEFORE UPDATE ON "client"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_config_file_template_modify_time BEFORE UPDATE ON "config_file_template"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_routing_config_v2_mtime BEFORE UPDATE ON "routing_config_v2"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_leader_election_mtime BEFORE UPDATE ON "leader_election"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_circuitbreaker_rule_v2_mtime BEFORE UPDATE ON "circuitbreaker_rule_v2"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_fault_detect_rule_mtime BEFORE UPDATE ON "fault_detect_rule"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_service_contract_mtime BEFORE UPDATE ON "service_contract"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_service_contract_detail_mtime BEFORE UPDATE ON "service_contract_detail"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_gray_resource_modify_time BEFORE UPDATE ON "gray_resource"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_modify_time();
CREATE TRIGGER trg_lane_group_mtime BEFORE UPDATE ON "lane_group"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
CREATE TRIGGER trg_lane_rule_mtime BEFORE UPDATE ON "lane_rule"
    FOR EACH ROW EXECUTE FUNCTION polaris_set_mtime();
