DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` varchar(200) NOT NULL,
  `login` varchar(200) NOT NULL,
  `password` varchar(200) NOT NULL,
  PRIMARY KEY(`id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `sessions`;
CREATE TABLE `sessions` (
  `login` varchar(200) NOT NULL,
  `jwt` varchar(512) NOT NULL,
  `expires_at` DATETIME NOT NULL,
  PRIMARY KEY(`login`),
  FOREIGN KEY (`login`) REFERENCES users(`login`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
