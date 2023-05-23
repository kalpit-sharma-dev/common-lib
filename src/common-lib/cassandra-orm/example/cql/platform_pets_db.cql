// To execute CQL statements from a file use SOURCE 'file_name.cql'
// To execute a CQL file without starting a shell session use cqlsh [-u user -p pwd] -f 'file_name.cql'

CREATE KEYSPACE IF NOT EXISTS platform_pets_db
WITH REPLICATION = {
	'class': 'SimpleStrategy',
	'replication_factor': '1'}
AND DURABLE_WRITES = true;

CREATE TABLE IF NOT EXISTS cats (
    "ID" uuid,
    "Name" text,
    "Age" int,
    "Weight" int,
    "Owners" list<text>,
    PRIMARY KEY ("ID", "Age")
);

CREATE TABLE IF NOT EXISTS cats_by_age (
    "ID" uuid,
    "Name" text,
    "Age" int,
    "Weight" int,
    "Owners" list<text>,
    PRIMARY KEY ("Age", "ID")
);

CREATE TABLE IF NOT EXISTS cats_by_name (
    "ID" uuid,
    "Name" text,
    "Age" int,
    "Weight" int,
    "Owners" list<text>,
    PRIMARY KEY ("Name", "Age", "ID")
);

CREATE TABLE IF NOT EXISTS dogs (
    "Name" text,
    "Age" int,
    "Weight" int,
    "Owners" list<text>,
    "DateOfBirth" timestamp,
    PRIMARY KEY ("Name")
);

INSERT INTO cats("ID", "Name", "Age", "Weight", "Owners") VALUES(4c3766db-94f0-11e8-b068-080027f00fcc, 'Tom', 5, 8, ['Molly']);
INSERT INTO cats("ID", "Name", "Age", "Weight", "Owners") VALUES(112f2b16-301c-4ae6-ad88-bf682dd29a34, 'Jerry', 3, 5, ['Bob', 'Mary']);

INSERT INTO cats_by_age("ID", "Name", "Age", "Weight", "Owners") VALUES(4c3766db-94f0-11e8-b068-080027f00fcc, 'Tom', 5, 8, ['Molly']);
INSERT INTO cats_by_age("ID", "Name", "Age", "Weight", "Owners") VALUES(112f2b16-301c-4ae6-ad88-bf682dd29a34, 'Jerry', 3, 5, ['Bob', 'Mary']);

INSERT INTO cats_by_name("ID", "Name", "Age", "Weight", "Owners") VALUES(4c3766db-94f0-11e8-b068-080027f00fcc, 'Tom', 5, 8, ['Molly']);
INSERT INTO cats_by_name("ID", "Name", "Age", "Weight", "Owners") VALUES(112f2b16-301c-4ae6-ad88-bf682dd29a34, 'Jerry', 3, 5, ['Bob', 'Mary']);

INSERT INTO dogs("Name", "Age", "Weight", "Owners") VALUES('Pluto', 5, 8, ['Polly']);
INSERT INTO dogs("Name", "Age", "Weight", "Owners") VALUES('Charlie', 3, 5, ['Mr.Brown', 'Mrs.Brown']);
