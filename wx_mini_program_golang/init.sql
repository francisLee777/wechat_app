CREATE DATABASE IF NOT EXISTS francis DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE francis;
ALTER DATABASE francis CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- auto-generated definition
create table IF NOT EXISTS user_info
(
    id            bigint unsigned auto_increment comment '自增id' primary key,
    openId        varchar(128) default ''                not null comment '微信提供的openId',
    user_nickName varchar(128) default ''                not null comment '创建人name',
    user_icon_url varchar(512) default ''                not null comment '头像url',
    status        int          default 1                 not null comment '1-正常状态  2-删号',
    create_time datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最后更新时间',
    family_id     bigint unsigned default 0                 not null comment '所属家庭ID',
    role          varchar(32)     default 'member'          not null comment '角色: admin/member',
    unique uk_open_id (openId)
) comment '用户信息' ENGINE = InnoDB
                     DEFAULT CHARSET = utf8mb4
                     COLLATE = utf8mb4_general_ci;

-- 家庭表
create table IF NOT EXISTS family
(
    id            bigint unsigned auto_increment comment '自增id' primary key,
    name          varchar(128) default ''                not null comment '家庭名称',
    owner_id      varchar(128) default ''                not null comment '创建人openId',
    create_time datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最后更新时间'
) comment '家庭表' ENGINE = InnoDB
                   DEFAULT CHARSET = utf8mb4
                   COLLATE = utf8mb4_general_ci;

-- 食谱表
create table IF NOT EXISTS recipe
(
    id            bigint unsigned auto_increment comment '自增id' primary key,
    family_id     bigint unsigned default 0                 not null comment '家庭id',
    name          varchar(128)  default ''                not null comment '食谱名称',
    images        json                                    comment '成品照片路径列表',
    content       text                                    comment '食谱具体内容',
    creator_id     varchar(128) default '' comment '创建者openId',
    creator_name   varchar(128) default '' comment '创建者昵称',
    creator_avatar varchar(512) default '' comment '创建者头像',
    sort_order    bigint        default 0                 not null comment '排序字段',
    create_time datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最后更新时间',
    index idx_family (family_id)
) comment '食谱表' ENGINE = InnoDB
                   DEFAULT CHARSET = utf8mb4
                   COLLATE = utf8mb4_general_ci;

-- 留言表
create table IF NOT EXISTS message
(
    id                 bigint unsigned auto_increment comment '自增id' primary key,
    family_id          bigint unsigned default 0                 not null comment '家庭id',
    user_id            varchar(128)  default ''                not null comment '留言者openId',
    user_name          varchar(128)  default ''                not null comment '留言者昵称',
    user_avatar        varchar(512)  default ''                not null comment '留言者头像',
    content            text                                    comment '留言内容',
    parent_id          bigint unsigned default null              comment '父留言ID',
    reply_to_user_id   varchar(128)  default ''                comment '回复给谁',
    reply_to_user_name varchar(128)  default ''                comment '回复给谁的昵称',
    create_time datetime default CURRENT_TIMESTAMP not null comment '创建时间',
    index idx_family_msg (family_id),
    index idx_parent (parent_id)
) comment '留言表' ENGINE = InnoDB
                   DEFAULT CHARSET = utf8mb4
                   COLLATE = utf8mb4_general_ci;

-- 每日菜单表
create table IF NOT EXISTS menu
(
    id          bigint unsigned auto_increment comment '自增id' primary key,
    family_id   bigint unsigned default 0                 not null comment '家庭id，关联 family 表',
    date        date                                      not null comment '日期，格式 YYYY-MM-DD',
    meal_type   tinyint(4)                                not null comment '餐别：1-早餐, 2-午餐, 3-晚餐',
    recipe_id   bigint unsigned default 0                 not null comment '食谱id，关联 recipe 表',
    user_id     varchar(128)    default ''                not null comment '点菜人openId',
    remark      varchar(255)    default '' comment '备注信息',
    create_time datetime        default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time datetime        default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '更新时间',
    index idx_family_date (family_id, date)
) comment '每日菜单表' ENGINE = InnoDB
                       DEFAULT CHARSET = utf8mb4
                       COLLATE = utf8mb4_general_ci;
