/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_asset   */
/******************************************/
CREATE TABLE `t_asset` (
  `code` varchar(128) NOT NULL,
  `addr` varchar(128) NOT NULL,
  `attrs` blob NOT NULL,
  `version` int(11) NOT NULL,
  `utc_st` bigint(20) NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_candidates   */
/******************************************/
CREATE TABLE `t_candidates` (
  `addr` varchar(128) NOT NULL,
  `votes` bigint(20) NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_context   */
/******************************************/
CREATE TABLE `t_context` (
  `lm_key` varchar(128) NOT NULL,
  `lm_val` blob NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`lm_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_equity   */
/******************************************/
CREATE TABLE `t_equity` (
  `code` varchar(128) NOT NULL,
  `id` varchar(128) NOT NULL,
  `addr` varchar(128) NOT NULL,
  `equity` varchar(128) NOT NULL,
  `version` int(11) NOT NULL,
  `utc_st` bigint(20) NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`,`addr`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_kv   */
/******************************************/
CREATE TABLE `t_kv` (
  `lm_key` varchar(128) NOT NULL,
  `lm_val` mediumblob NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`lm_key`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_meta_data   */
/******************************************/
CREATE TABLE `t_meta_data` (
  `id` varchar(128) NOT NULL,
  `code` varchar(128) NOT NULL,
  `addr` varchar(128) NOT NULL,
  `attrs` blob NOT NULL,
  `version` int(11) NOT NULL,
  `utc_st` bigint(20) NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
;

/******************************************/
/*   DatabaseName = lemochain   */
/*   TableName = t_tx   */
/******************************************/
CREATE TABLE `t_tx` (
  `thash` varchar(128) NOT NULL,
  `phash` varchar(128) NOT NULL,
  `bhash` varchar(128) NOT NULL,
  `height` bigint(20) NOT NULL,
  `faddr` varchar(128) NOT NULL,
  `taddr` varchar(128) NOT NULL,
  `tx` blob NOT NULL,
  `flag` int(11) NOT NULL,
  `utc_st` bigint(22) NOT NULL,
  `st` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `package_time` bigint(20) DEFAULT NULL,
  `asset_code` varchar(128) DEFAULT NULL,
  `asset_id` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`thash`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC
;
