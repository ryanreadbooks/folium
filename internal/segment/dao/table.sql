CREATE TABLE IF NOT EXISTS alloc_table (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  biz_key VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'biz key identifier',
  cur_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'max id',
  step INT UNSIGNED NOT NULL DEFAULT 0 COMMENT 'step',
  created_at BIGINT NOT NULL DEFAULT 0 COMMENT 'created unix ms',
  updated_at BIGINT NOT NULL DEFAULT 0 COMMENT 'updated unix ms',
  PRIMARY KEY (id),
  UNIQUE KEY uk_key(biz_key)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='segment allocation table';
