package lib

const initSqlQueries = `
CREATE TABLE IF NOT EXISTS players (
  email varchar(255) NOT NULL DEFAULT '',
  displayName varchar(255) NOT NULL DEFAULT '',
  password varchar(255) NOT NULL DEFAULT '',
  rating float(8,2) NOT NULL DEFAULT '0.00',
  ratingCount int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS teams (
  teamId int(11) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL DEFAULT '',
  rating float(8,2) NOT NULL DEFAULT '0.00',
  PRIMARY KEY (teamId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS ratings (
  rateeEmail varchar(255) NOT NULL DEFAULT '',
  raterEmail varchar(255) NOT NULL DEFAULT '',
  rating float(8,2) NOT NULL DEFAULT '0.00',
  PRIMARY KEY (rateeEmail,raterEmail)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS matches (
  matchId int(11) unsigned NOT NULL AUTO_INCREMENT,
  numOfTeams int(11) NOT NULL DEFAULT '0',
  numOfPlayersPerTeam int(11) NOT NULL DEFAULT '0',
  totalPlayers int(11) NOT NULL DEFAULT '0',
  date datetime DEFAULT NULL,
  rating float(8,2) NOT NULL DEFAULT '0.00',
  PRIMARY KEY (matchId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS matchPlayers (
  matchId int(11) unsigned NOT NULL DEFAULT '0',
  playerEmail varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (matchId,playerEmail)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS matchTeams (
  matchId int(11) unsigned NOT NULL DEFAULT '0',
  teamId int(11) unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (matchId, teamId)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS teamPlayers (
  teamId int(11) unsigned NOT NULL DEFAULT '0',
  playerEmail varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (teamId,playerEmail)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`