-- Upload 应用数据库表
-- 执行前请确保已选择正确的数据库: USE gbaseadmin;

-- 1. 文件目录
CREATE TABLE IF NOT EXISTS upload_dir (
  id bigint unsigned NOT NULL COMMENT 'ID',
  parent_id bigint unsigned DEFAULT 0 COMMENT '上级目录',
  name varchar(100) NOT NULL COMMENT '目录名称',
  path varchar(500) NOT NULL COMMENT '目录路径',
  sort int DEFAULT 0 COMMENT '排序',
  status tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态:0=禁用,1=启用',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at datetime COMMENT '删除时间',
  created_by bigint unsigned COMMENT '创建人',
  dept_id bigint unsigned COMMENT '部门ID',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件目录';

-- 2. 文件记录
CREATE TABLE IF NOT EXISTS upload_file (
  id bigint unsigned NOT NULL COMMENT 'ID',
  dir_id bigint unsigned DEFAULT 0 COMMENT '所属目录',
  name varchar(255) NOT NULL COMMENT '文件名称',
  url varchar(500) NOT NULL COMMENT '文件地址',
  ext varchar(20) DEFAULT '' COMMENT '文件扩展名',
  size bigint unsigned DEFAULT 0 COMMENT '文件大小',
  mime varchar(100) DEFAULT '' COMMENT 'MIME类型',
  storage tinyint(1) NOT NULL DEFAULT 1 COMMENT '存储类型:1=本地,2=阿里云OSS,3=腾讯云COS',
  is_image tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否图片:0=否,1=是',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at datetime COMMENT '删除时间',
  created_by bigint unsigned COMMENT '创建人',
  dept_id bigint unsigned COMMENT '部门ID',
  PRIMARY KEY (id),
  KEY idx_dir_id (dir_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件记录';

-- 3. 上传配置
CREATE TABLE IF NOT EXISTS upload_config (
  id bigint unsigned NOT NULL COMMENT 'ID',
  name varchar(100) NOT NULL COMMENT '配置名称',
  storage tinyint(1) NOT NULL DEFAULT 1 COMMENT '存储类型:1=本地,2=阿里云OSS,3=腾讯云COS',
  is_default tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否默认:0=否,1=是',
  local_path varchar(500) DEFAULT '' COMMENT '本地存储路径',
  oss_endpoint varchar(255) DEFAULT '' COMMENT 'OSS Endpoint',
  oss_bucket varchar(255) DEFAULT '' COMMENT 'OSS Bucket',
  oss_access_key varchar(255) DEFAULT '' COMMENT 'OSS AccessKey',
  oss_secret_key varchar(255) DEFAULT '' COMMENT 'OSS SecretKey',
  cos_region varchar(100) DEFAULT '' COMMENT 'COS Region',
  cos_bucket varchar(255) DEFAULT '' COMMENT 'COS Bucket',
  cos_secret_id varchar(255) DEFAULT '' COMMENT 'COS SecretId',
  cos_secret_key varchar(255) DEFAULT '' COMMENT 'COS SecretKey',
  max_size int DEFAULT 10 COMMENT '最大文件大小(MB)',
  status tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态:0=禁用,1=启用',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at datetime COMMENT '删除时间',
  created_by bigint unsigned COMMENT '创建人',
  dept_id bigint unsigned COMMENT '部门ID',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='上传配置';

-- 4. 文件目录规则
CREATE TABLE IF NOT EXISTS upload_dir_rule (
  id bigint unsigned NOT NULL COMMENT 'ID',
  dir_id bigint unsigned NOT NULL COMMENT '目录ID',
  category tinyint(1) NOT NULL DEFAULT 1 COMMENT '类别:1=默认,2=类型,3=接口',
  save_path varchar(500) DEFAULT '' COMMENT '保存目录',
  status tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态:0=禁用,1=启用',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at datetime COMMENT '删除时间',
  created_by bigint unsigned COMMENT '创建人',
  dept_id bigint unsigned COMMENT '部门ID',
  PRIMARY KEY (id),
  KEY idx_dir_id (dir_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件目录规则';
