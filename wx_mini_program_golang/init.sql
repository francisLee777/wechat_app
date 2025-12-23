CREATE DATABASE IF NOT EXISTS francis;
USE francis;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100)
);

INSERT INTO users (name, email) VALUES
('Francis', 'francis@example.com'),
('Alice', 'alice@example.com');


-- auto-generated definition
create table kitchen_order
(
    id            bigint unsigned auto_increment comment '自增id' primary key,
    unique_code   varchar(128)  default ''                not null comment '订单号[全局唯一]',
    openId        varchar(128)  default ''                not null comment '创建人id',
    create_name   varchar(128)  default ''                not null comment '创建人name',
    info          varchar(2048) default ''                not null comment '订单内容，暂时不做索引直接怼进去',
    status        int           default 1                 not null comment '订单状态， 1-新建[提单]  2-已支付',
    target_period varchar(32)   default ''                not null comment '订餐的目标时间，取值只有9个,已校验.明天中午、后天早上、今天晚上等',
    extra         varchar(2048) default ''                not null comment '扩展字段',
    create_time   timestamp     default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time   timestamp     default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最后更新时间',
    unique uk_order_code (unique_code),
    unique idx_order_creator (openId, target_period)
) comment '餐厅订单表[状态表]';


create table user_info
(
    id            bigint unsigned auto_increment comment '自增id' primary key,
    openId        varchar(128) default ''                not null comment '微信提供的openId',
    user_nickName varchar(128) default ''                not null comment '创建人name',
    user_icon_url varchar(512) default ''                not null comment '头像url',
    status        int          default 1                 not null comment '1-正常状态  2-删号',
    create_time   timestamp    default CURRENT_TIMESTAMP not null comment '创建时间',
    update_time   timestamp    default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP comment '最后更新时间',
    unique uk_open_id (openId)
) comment '用户信息';
