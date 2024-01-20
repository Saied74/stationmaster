-- Host: localhost    Database: stationmaster
-- ------------------------------------------------------
-- Server version	10.5.19-MariaDB-0+deb11u2

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `defaults`
--

DROP TABLE IF EXISTS `defaults`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `defaults` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `kee` varchar(20) NOT NULL,
  `val` varchar(100) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_defaults_kee` (`kee`)
) ENGINE=InnoDB AUTO_INCREMENT=29 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `defaults`
--

LOCK TABLES `defaults` WRITE;
/*!40000 ALTER TABLE `defaults` DISABLE KEYS */;
INSERT INTO `defaults` VALUES (1,'mode','CW'),(2,'band','10m'),(3,'qrzkey','99576ae24c5d4fd31a1c39f79f99c430'),(4,'contest','No'),(5,'sent','599'),(6,'exch','NJ'),(7,'contestname','SKCC'),(9,'split','Off'),(10,'xfreq','14.200000'),(11,'rfreq','14.200000'),(12,'10mxfreq','28.011600'),(13,'10mrfreq','28.011600'),(14,'15mrfreq','21.006700'),(15,'15mxfreq','21.006700'),(16,'20mxfreq','14.240000'),(17,'20mrfreq','14.240000'),(18,'40mrfreq','7.015000'),(19,'40mxfreq','7.015000'),(20,'80mxfreq','3.694000'),(21,'80mrfreq','3.694000'),(22,'160mrfreq','2.000000'),(23,'160mxfreq','2.000000'),(25,'WWVxfreq','10.000000'),(26,'WWVrfreq','10.000000'),(27,'Auxrfreq','10.497000'),(28,'Auxxfreq','10.497000');
/*!40000 ALTER TABLE `defaults` ENABLE KEYS */;
UNLOCK TABLES;

--

