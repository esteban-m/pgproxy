### Steps for mysql -> pg proxy

* [X] 1.Proxy mysql
* [X] 2.Parse&support packet
* [X] 3.extract sql
* [ ] 4.corner case handling/fundamental part handling
* [ ] 5.pg protocol
* [ ] 6.mysql -> pg sql conversion: ddl/dml
* [ ] 7.bridge mysql and pg
* [ ] MySQL behaviour
* [ ] Connection pool functions: Connection/Transaction multiplexing
* [ ] SQL translation Cache
* [ ] character conversion: '`' to '"'
* [ ] CREATE TABLE - AUTO_INCREMENT
* [ ] CREATE TABLE - CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_520_ci
* [ ] CREATE TABLE - bigint UNSIGNED
* [ ] CREATE TABLE - index/primary key - USING BTREE
* [ ] CREATE TABLE - INDEX `idx_name`(`field`(20) ASC) USING BTREE
* [ ] CREATE TABLE - datetime + DEFAULT '0000-00-00 00:00:00'
* [ ] CREATE TABLE - tinytext/mediumtext/longtext
* [ ] Conversion note: double-quoted identifiers including column names are case-sensitive. not double-quoted are folded
  to lowercase. This may affect 'create table' and queries(select/etc.)

Top priority:

* [ ] New sql parser with flexible modification support(parse mysql sql, generate postgresql sql)

### Mysql support:

#### TODO

* SSL/Compression
* Auth methods
* statement cache management
* request->response mapping: pipelining + seqId for incomplete payload only
* data type conversion
* More precise sql lib
* More precise mysql/pg protocol lib
* More precise postgresql sql generator from AST
* Logging
    * All SQLs
    * Success SQLs
    * Failed SQLs by pg server
    * Unsupported SQLs by pgproxy
    * modified SQLs and pg server success/failed

#### Will not supported in short term

* Cursor on ResultSet / Large ResultSet

