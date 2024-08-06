CREATE TABLE IF NOT EXISTS urmysajupalja (
	loginuuid   char(40) PRIMARY KEY NOT NULL,
	yearChun varchar(1),
	yearJi varchar(1),
	monthChun varchar(1),
	monthJi varchar(1),
	dayChun varchar(1),
	dayJi varchar(1),
	timeChun varchar(1),
	timeJi varchar(1)
);

CREATE TABLE IF NOT EXISTS urmysajudaesaeun (
	loginuuid   char(40) PRIMARY KEY NOT NULL,
	daeunChun varchar(1),
	daeunJi varchar(1),
	saeunChun varchar(1),
	saeunJi varchar(1)
);


CREATE TAble IF NOT EXISTS urmybirthday (
	birthday varchar(20) PRIMARY KEY NOT NULL,
	urmyyear smallint,
	urmymonth smallint,
	urmyday smallint,
	urmytime varchar(5)
);

CREATE TABLE IF NOT EXISTS urmyuserinfo (
		loginuuid  char(40) NOT NULL UNIQUE PRIMARY KEY,
		updatedprofile BOOLEAN,
		deviceos varchar(50),
		email   varchar(70),
		phoneno  varchar(20),
		phonecode varchar(6),
		name varchar(25),
		description varchar(255),
		country varchar(50),
		city varchar(50),
		residentcountry varchar(50) DEFAULT 'none',
		residentstate varchar(50) DEFAULT 'none',
		residentcity varchar(50) DEFAULT 'none',
		birthday varchar(20) NOT NULL REFERENCES urmybirthday(birthday),
		birthdayupdate date,
		mbti varchar(4) NOT NULL,
		gender BOOLEAN NOT NULL,
		urmycount smallint DEFAULT 0,
		tester BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS urmyunusualuser (
		loginuuid  char(40) NOT NULL UNIQUE PRIMARY KEY,
		updatedprofile BOOLEAN,
		deviceos varchar(50),
		email   varchar(70),
		phoneno  varchar(20),
		phonecode varchar(6),
		name varchar(25),
		description varchar(255),
		country varchar(50),
		city varchar(50),
		residentcountry varchar(50) DEFAULT 'none',
		residentstate varchar(50) DEFAULT 'none',
		residentcity varchar(50) DEFAULT 'none',
		birthday varchar(20) NOT NULL REFERENCES urmybirthday(birthday),
		birthdayupdate date,
		mbti varchar(4) NOT NULL,
		gender BOOLEAN NOT NULL,
		urmycount smallint DEFAULT 0,
		tester BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS urmyusersetting (
		loginuuid  char(70) NOT NULL UNIQUE PRIMARY KEY,
		minage smallint,
		maxage smallint,
		delreq BOOLEAN
);

CREATE TABLE IF NOT EXISTS urmycompensationlist (
	email varchar(70) NOT NULL UNIQUE PRIMARY KEY
);


CREATE MATERIALIZED VIEW urmyuserinfo_kr
AS
SELECT u.loginuuid, u.name, u.description, u.residentcountry, u.residentstate, u.residentcity, u.birthday, u.birthdayupdate, u.mbti, u.gender, u.tester, u.urmycount, b.urmyyear, b.urmymonth, b.urmyday, b.urmytime
FROM urmyuserinfo as u LEFT JOIN urmybirthday as b ON u.birthday = b.birthday
WITH NO DATA;


CREATE TABLE IF NOT EXISTS urmyusersagreement (
			loginuuid char(40) PRIMARY KEY NOT NULL,
			isoverage   BOOLEAN,
			urmyprivacy  BOOLEAN,
			urmyservice  BOOLEAN,
			urmyillegal  BOOLEAN,
			createdAt DATE
);


CREATE TABLE IF NOT EXISTS urmyservice (
	serviceid varchar(6) PRIMARY KEY NOT NULL,
	servicename varchar(10) NOT NULL,
	serviceprice numeric(10,5) NOT NULL
);

CREATE TABLE IF NOT EXISTS urmysales (
	invoiceid SERIAL UNIQUE PRIMARY KEY NOT NULL,
	serviceid varchar(6) NOT NULL REFERENCES urmyservice(serviceid),
	purchaseuser char(40) NOT NULL REFERENCES urmyuserinfo(loginuuid),
	purchasedate DATE NOT NULL
);

CREATE TABLE IF NOT EXISTS urmyrefundrequest (
	refundid SERIAL UNIQUE PRIMARY KEY NOT NULL,
	invoiceid INTEGER NOT NULL REFERENCES urmysales(invoiceid),
	refunduser char(40) NOT NULL REFERENCES urmyuserinfo(loginuuid),
	refundrequestdate DATE NOT NULL,
	refunded BOOLEAN
);


CREATE TABLE IF NOT EXISTS urmyrefundcommit (
	refunsucceededid SERIAL UNIQUE PRIMARY KEY NOT NULL,
	refundid INTEGER NOT NULL REFERENCES urmyrefundrequest(refundid),
	refundcommiteddate date NOT NULL,
	refundstate bool NOT NULL
);

CREATE TABLE IF NOT EXISTS urmyuserprofile (
			id SERIAL PRIMARY KEY,
			loginuuid char(40) NOT NULL REFERENCES urmyuserinfo(loginuuid),
			profilename  varchar(100),
			updateddate DATE,
			profilesequence INTEGER
);


CREATE TABLE IF NOT EXISTS urmyuserdelreq (
	loginuuid  char(40) NOT NULL UNIQUE PRIMARY KEY ,
	email   varchar(70),
	delreqedate DATE
);

CREATE TABLE IF NOT EXISTS urmyuserdeleted (
		invoiceid SERIAL UNIQUE PRIMARY KEY NOT NULL,
		loginuuid  char(40) NOT NULL,
		country varchar(50),
		city varchar(50),
		birthday varchar(20) NOT NULL REFERENCES urmybirthday(birthday),
		mbti varchar(4) NOT NULL,
		gender BOOLEAN NOT NULL,
		deldate DATE
);
CREATE TABLE IF NOT EXISTS urmycandidatereport (
		caseid SERIAL PRIMARY KEY,
		reportinguser char(40)  NOT NULL REFERENCES urmyuserinfo(loginuuid),
		reporteduser  char(40)  NOT NULL REFERENCES urmyuserinfo(loginuuid),
		reporttype varchar(10),
		contents varchar(500),
		registereddate DATE,
		states char(10)
);

CREATE TABLE IF NOT EXISTS urmychatreport (
		caseid SERIAL PRIMARY KEY,
		reportinguser char(40)  NOT NULL REFERENCES urmyuserinfo(loginuuid),
		reporteduser  char(40)  NOT NULL REFERENCES urmyuserinfo(loginuuid),
		chatroomid varchar(70) NOT NULL,
		reporttype varchar(10),
		messages varchar(300),
		messagetime varchar(20),
		contents varchar(500),
		registereddate DATE,
		states char(10)
);

CREATE TABLE IF NOT EXISTS urmyadminuserinfo (
      loginid   varchar(30) PRIMARY KEY,
      userpassword  varchar(20),
      username varchar(25),
      urmydepartment varchar(50)
);

CREATE TABLE IF NOT EXISTS urmyinquiry (
      caseid SERIAL PRIMARY KEY,
      userid  char(40) NOT NULL,
      inquirytype varchar(20) NOT NULL,
      title varchar(50),
      inquirystate varchar(20)
);

CREATE TABLE IF NOT EXISTS urmyinquirycontent (
      contentsid SERIAL PRIMARY KEY,
      caseid integer NOT NULL REFERENCES urmyinquiry(caseid),
      contents varchar(500),
	  managerid varchar(50),
      registereddate TIMESTAMP
);


CREATE INDEX urmybirthday_urmyyear_index_brin_idx ON urmybirthday USING brin(urmyyear);
CREATE INDEX urmyuserinfo_birthday_index_brin_idx ON urmyuserinfo USING brin(birthday);
CREATE INDEX urmyrefundrequest_invoiceid_index_brin_idx ON urmyrefundrequest USING brin(invoiceid);
CREATE INDEX urmyrefundrequest_refunduser_index_brin_idx ON urmyrefundrequest USING brin(refunduser);
CREATE INDEX urmysales_serviceid_index_brin_idx ON urmysales USING brin(serviceid);
CREATE INDEX urmysales_purchaseuser_index_brin_idx ON urmysales USING brin(purchaseuser);
CREATE INDEX urmyrefundcommit_refundid_index_brin_idx ON urmyrefundcommit USING brin(refundid);
CREATE INDEX urmycandidatereport_reporteduser_index_brin_idx ON urmycandidatereport USING brin(reporteduser);
CREATE INDEX urmychatreport_reporteduser_index_brin_idx ON urmychatreport USING brin(reporteduser);
CREATE INDEX urmyinquirycontent_caseid_index_brin_idx ON urmyinquirycontent USING brin(caseid);


ALTER TABLE urmyuserinfo ALTER COLUMN tester SET DEFAULT false;
ALTER TABLE urmyuserinfo ALTER COLUMN tester SET NOT NULL;
ALTER TABLE urmyuserinfo ALTER COLUMN tester TYPE BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS urmyusers (
			loginuuid char(30) PRIMARY KEY NOT NULL,
			notificationtoken  char(170),
			platform char(3),
			lastlogin DATE
);

CREATE INDEX urmyuserinfobirthindex ON urmyuserinfo USING btree (birthday);
CREATE INDEX urmybirthyearindex ON urmybirthday USING btree (urmyyear);
SET TIMEZONE='ASIA/SEOUL';
SET enable_seqscan = OFF;

CREATE USER koreaogh WITH SUPERUSER CREATEROLE CREATEDB REPLICATION LOGIN INHERIT BYPASSRLS PASSWORD 'test1234';
CREATE USER koreaogh WITH SUPERUSER CREATEROLE CREATEDB REPLICATION LOGIN INHERIT BYPASSRLS PASSWORD 'test1234';
CREATE DATABASE urmydb with owner koreaogh encoding 'UTF8';
CREATE DATABASE urmydb with owner urmyojh encoding 'UTF8';
UPDATE urmysaju SET yearChun='ja', yearJi='chuk', monthChun='in', monthJi='myo', dayChun='jin', dayJi='sa', timeChun='O', timeJi='mi' WHERE phoneNo=(SELECT phoneNo FROM urmyusers WHERE loginid='tgja1075@gmail.com')
UPDATE urmyuserinfo SET description="RG9u4oCZdCB0b3VjaCBtZQ==";

INSERT INTO urmyusers (loginId, password, nickname, name, phoneNo, gender, birthday, createdat) VALUES ('tgja1075', 'qwer1234', 'whitewhale', 'johnoh', '01011112222', true, '910807', current_timestamp)
UPDATE urmysaju SET loginid='tgja1420@gmail.com' WHERE loginid='tjga1420@gmail.com';

INSERT INTO urmyuserinfo (loginuuid, email, name, birthday, gender, mbti) VALUES ("", $2, $3, SELECT birthday form urmybirthday WHERE birthday=$4, $5, $6)

SELECT u.loginuuid, u.name, u.birthday, u.gender FROM urmyuserinfo as u RIGHT JOIN urmybirthday as b ON (b.urmyyear >= 1895 AND b.birthday < 1998) WHERE u.gender = false

SELECT DISTINCT u.loginuuid, u.name, u.birthday, u.gender FROM urmyuserinfo as u INNER JOIN urmybirthday as b ON u.gender = false;

SELECT birthday, urmyyear FROM urmybirthday WHERE urmyyear BETWEEN 1895 AND 1997;

SELECT DISTINCT u.loginuuid, u.name, u.birthday, u.gender FROM urmyuserinfo as u INNER JOIN (SELECT birthday, urmyyear FROM urmybirthday WHERE urmyyear BETWEEN 1895 AND 1997) as b ON u.gender = false;

SELECT DISTINCT u.loginuuid, u.name, u.birthday, u.gender FROM urmyuserinfo as u INNER JOIN (SELECT birthday, urmyyear FROM urmybirthday WHERE urmyyear BETWEEN 1990 AND 1992) as b ON u.gender = false;

SELECT u.loginuuid, u.name, u.gender, b.birthday, b.urmyyear FROM urmyuserinfo as u Left JOIN urmybirthday as b on u.birthday  = b.birthday where  u.gender = false and b.urmyyear between 1986 and 1998;

SELECT u.loginuuid, u.name, u.gender, b.birthday, b.urmyyear
FROM urmyuserinfo as u 
Left JOIN urmybirthday as b on u.birthday  = b.birthday 
where  u.gender = false and b.urmyyear between 1986 and 1998
order by RANDOM() limit 100;

SELECT DISTINCT u.loginuuid, u.name, u.birthday, u.gender FROM urmyuserinfo as u INNER JOIN urmybirthday as b ON u.gender = $1 INNER JOIN (SELECT birthday, urmyyear FROM urmybirthday) ub ON u.birthday = ub.birthday WHERE ub.urmyyear BETWEEN $2 AND $3;

SELECT u.loginuuid, u.name, u.gender, b.birthday, b.urmyyear FROM urmyuserinfo as u Left JOIN urmybirthday as b on u.birthday  = b.birthday where  u.gender = $1 and b.urmyyear between $2 and $3order by RANDOM() limit $4;


SELECT table_name, is_updatable, is_insertable_into FROM information_schema.views WHERE table_name = urmyuserinfo;

pg_dump -h 127.0.0.1 -U koreaogh -d urmydb -W > ./urmydump.dump

////////
DELETE FROM urmyuserprofile Where loginuuid='XynSxQD035Xc6uGyPD98fNkfFlm1';


INSERT INTO urmyuserinfo (serviceid, servicename, price) VALUES ("s10000", "premium", 20000);
INSERT INTO urmyuserinfo (serviceid, servicename, price) VALUES ("s10002", "top100", 200000);
INSERT INTO urmyuserinfo (serviceid, servicename, price) VALUES ("s10003", "mbtiemoticon", 2000);
INSERT INTO urmyinquirycontent(caseid, contents, registereddate) VALUES ("2", "tested", "2022-09-07T09:08:11");

INSERT INTO urmyinquirycontent(caseid, contents, registereddate) VALUES (1, 'dGVzdDEyMzR0ZXN0MTIzNHRlc3QxMjM0dGVzdDEyMzQ=', '2022-09-08T09:08:11');
INSERT INTO urmyinquirycontent(caseid, managerid, contents, registereddate) VALUES (1, 'qwer1234','dGVzdDEyMzR0ZXN0MTIzNHRlc3QxMjM0dGVzdDEyMzQ=', '2022-09-08T09:08:11');