### Steps for mysql -> pg proxy

*[X] 1.Proxy mysql
*[X] 2.Parse&support packet
*[X] 3.extract sql
*[ ] 4.corner case handling/fundamental part handling
*[ ] 5.pg protocol
*[ ] 6.mysql -> pg sql conversion: ddl/dml
*[ ] 7.bridge mysql and pg
*[ ] 8.Connection/Transaction multiplexing


### Mysql support:

#### TODO
* SSL/Compression
* Auth methods
* statement cache management
* request->response mapping: pipelining + seqId for incomplete payload only
* data type conversion

