# pgproxy

### Project Goal

* Proxy MySQL to PostgreSQL
    * MySQL SQL Sniffer
    * Proxy MySQL client to PostgreSQL backend
* PostgreSQL connection pool
* Using filtering unsupported SQL out method to indicate SQL translation(Not always translate SQLs)
    * Internal functions translation
    * Data type mapping
    * SQL Hints

### Workflow

1. Incoming request: Command/SQL/Parameters
2. Parse SQL
3. SQL translation(optional) + Request data mapping
4. Send postgresql request and receive response
5. Response data mapping
6. Respond MySQL request with converted response

### Dependencies

#### Sql Parser

* [ ] https://github.com/blastrain/vitess-sqlparser
* [ ] https://github.com/arana-db/parser

