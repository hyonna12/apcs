-- --------------------------------------------------------
-- 호스트:                          lineworldap.iptime.org
-- 서버 버전:                        10.2.27-MariaDB-1:10.2.27+maria~bionic - mariadb.org binary distribution
-- 서버 OS:                        debian-linux-gnu
-- HeidiSQL 버전:                  12.3.0.6589
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- apcs_dev 데이터베이스 구조 내보내기
DROP DATABASE IF EXISTS `apcs_dev`;
CREATE DATABASE IF NOT EXISTS `apcs_dev` /*!40100 DEFAULT CHARACTER SET latin1 */;
USE `apcs_dev`;

-- 테이블 apcs_dev.sessions 구조 내보내기
DROP TABLE IF EXISTS `sessions`;
CREATE TABLE IF NOT EXISTS `sessions` (
  `session_id` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
  `expires` int(11) unsigned NOT NULL,
  `data` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL,
  PRIMARY KEY (`session_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- 내보낼 데이터가 선택되어 있지 않습니다.

-- 테이블 apcs_dev.TN_CTR_ITEM 구조 내보내기
DROP TABLE IF EXISTS `TN_CTR_ITEM`;
CREATE TABLE IF NOT EXISTS `TN_CTR_ITEM` (
  `item_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '물품의 ID',
  `item_height` smallint(5) unsigned DEFAULT 0 COMMENT 'mm 단위',
  `tracking_number` int(10) unsigned DEFAULT NULL COMMENT '운송장 번호',
  `INPUT_DATE` timestamp NULL DEFAULT NULL COMMENT '수납시간',
  `OUTPUT_DATE` timestamp NULL DEFAULT NULL COMMENT '불출시간',
  `delivery_id` bigint(20) unsigned DEFAULT NULL COMMENT '배달부 ID',
  `owner_id` bigint(20) unsigned DEFAULT NULL COMMENT '소유자 ID',
  `c_datetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `u_datetime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`item_id`),
  KEY `FK_TN_CTR_ITEM_TN_INF_DELIVERY` (`delivery_id`),
  KEY `FK_TN_CTR_ITEM_TN_INF_OWNER` (`owner_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=latin1 ROW_FORMAT=DYNAMIC;

-- 내보낼 데이터가 선택되어 있지 않습니다.

-- 테이블 apcs_dev.TN_CTR_SLOT 구조 내보내기
DROP TABLE IF EXISTS `TN_CTR_SLOT`;
CREATE TABLE IF NOT EXISTS `TN_CTR_SLOT` (
  `slot_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '슬롯의 ID',
  `lane` tinyint(3) unsigned DEFAULT NULL COMMENT '슬롯의 열 번호',
  `floor` tinyint(3) unsigned DEFAULT NULL COMMENT '슬롯의 행 번호, 아래부터 1부터 시작',
  `transport_distance` mediumint(9) DEFAULT NULL COMMENT 'mm 단위',
  `slot_enabled` bit(1) DEFAULT NULL COMMENT '슬롯이 사용 가능 여부(고장등, 사용불가), 기본 1',
  `slot_keep_cnt` int(11) unsigned NOT NULL DEFAULT 0 COMMENT '확보 슬롯의 갯수(0이면 사용중). 물건이 든 슬롯의 바로 아래 빈 칸은 1, 그 아래 빈칸은 2, ...',
  `tray_id` bigint(20) unsigned DEFAULT NULL COMMENT '트레이의 ID',
  `item_id` bigint(20) unsigned DEFAULT NULL COMMENT '물품의 ID',
  `check_datetime` timestamp NULL DEFAULT NULL COMMENT '슬롯 체크 시간',
  `c_datetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `u_datetime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`slot_id`),
  KEY `FK_TN_CTR_SLOT_TN_CTR_TRAY` (`tray_id`),
  KEY `FK_TN_CTR_SLOT_TN_CTR_ITEM` (`item_id`)
) ENGINE=InnoDB AUTO_INCREMENT=65 DEFAULT CHARSET=latin1;

-- 내보낼 데이터가 선택되어 있지 않습니다.

-- 테이블 apcs_dev.TN_CTR_TRAY 구조 내보내기
DROP TABLE IF EXISTS `TN_CTR_TRAY`;
CREATE TABLE IF NOT EXISTS `TN_CTR_TRAY` (
  `tray_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '트레이의 ID',
  `tray_occupied` bit(1) DEFAULT b'1' COMMENT '트레이가 사용 중인지 여부',
  `item_id` bigint(20) unsigned DEFAULT NULL COMMENT '물품의 ID',
  `c_datetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `u_datetime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`tray_id`),
  KEY `FK_TN_CTR_TRAY_TN_CTR_ITEM` (`item_id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=latin1 ROW_FORMAT=DYNAMIC;

-- 내보낼 데이터가 선택되어 있지 않습니다.

-- 테이블 apcs_dev.TN_INF_DELIVERY 구조 내보내기
DROP TABLE IF EXISTS `TN_INF_DELIVERY`;
CREATE TABLE IF NOT EXISTS `TN_INF_DELIVERY` (
  `delivery_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '배달부 ID',
  `delivery_company` varchar(50) DEFAULT NULL COMMENT '배달부 소속',
  `c_datetime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `u_datetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`delivery_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=latin1;

-- 내보낼 데이터가 선택되어 있지 않습니다.

-- 테이블 apcs_dev.TN_INF_OWNER 구조 내보내기
DROP TABLE IF EXISTS `TN_INF_OWNER`;
CREATE TABLE IF NOT EXISTS `TN_INF_OWNER` (
  `owner_id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '소유자 ID',
  `phone_num` varchar(50) DEFAULT '000-0000-0000' COMMENT '소유자 연락처',
  `address` varchar(50) DEFAULT NULL COMMENT '주소',
  `password` int(10) unsigned DEFAULT NULL COMMENT '물품 수령 비밀번호',
  `c_datetime` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `u_datetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`owner_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=latin1;

-- 내보낼 데이터가 선택되어 있지 않습니다.

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
