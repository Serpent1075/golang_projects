CREATE TABLE IF NOT EXISTS urmyuserinfo (
		loginuuid  char(40) NOT NULL UNIQUE PRIMARY KEY,
		deviceos varchar(50),
		email   varchar(70),
		name varchar(25)
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
UPDATE urmyuserinfo SET tester='true';
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

UPDATE