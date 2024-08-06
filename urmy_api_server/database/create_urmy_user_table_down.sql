DROP MATERIALIZED VIEW urmyuserinfo_kr;
DROP TABLE urmyrefundcommit;
DROP TABLE urmyrefundrequest;
DROP TABLE urmysales;
DROP TABLE urmyservice;
DROP TABLE urmychatreport;
DROP TABLE urmycandidatereport;
DROP TABLE urmyinquirycontent;
DROP TABLE urmyinquiry;
DROP TABLE urmyuserinfo;
DROP TABLE urmyunusualuser;
DROP TABLE urmyuserdeleted;
DROP TABLE urmybirthday;

DROP TABLE urmysajupalja;
DROP TABLE urmysajudaesaeun;
DROP TABLE urmyusersagreement;
DROP TABLE urmyuserprofile;
DROP TABLE urmyusersetting;
DROP TABLE urmycompensationlist;
DROP TABLE urmyuserdelreq;
DROP TABLE urmyadminuserinfo;









DROP INDEX urmybirthday_urmyyear_index_brin_idx;
DROP INDEX urmyuserinfo_birthday_index_brin_idx;
DROP INDEX urmyrefundrequest_invoiceid_index_brin_idx;
DROP INDEX urmyrefundrequest_refunduser_index_brin_idx;
DROP INDEX urmysales_serviceid_index_brin_idx;
DROP INDEX urmysales_purchaseuser_index_brin_idx;
DROP INDEX urmyrefundcommit_refundid_index_brin_idx;

DROP TABLE urmyusers;
DROP INDEX urmyuserinfobirthindex;
DROP INDEX urmybirthyearindex;