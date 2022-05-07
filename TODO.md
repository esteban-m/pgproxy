### Steps for mysql -> pg proxy

[X] 1. Proxy mysql
2. Parse&support packet
3. extract sql
4. corner case handling/fundamental part handling
5. pg protocol
6. mysql -> pg sql conversion: ddl/dml
7. bridge mysql and pg
8. Connection/Transaction multiplexing




Mysql support:

Done
* connection phase
* command phase

TODO
* SSL
* Auth methods
* statement cache management
* Prepared statement matching
* request->response mapping: pipelining + seqId for incomplete payload only
* Mysql: over 16MB support
