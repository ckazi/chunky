# Chunky - PostgreSQL Data Integrity Checker

This program is designed to identify and manage corrupted rows in PostgreSQL databases. Using Go, it supports multithreading to efficiently handle large data sets and offers options for logging and deleting corrupted entries.

![image](https://github.com/ckazi/chunky/assets/45271263/96d997e1-8f43-4fac-a1f7-228a6f05cba4)

## Features

- **Multi-threading**: Utilizes multiple threads to improve performance on large databases.
- **Flexible database connection**: Customize database connection details through command-line flags.
- **Corrupted row management**: Detects corrupted rows and optionally deletes them, logging all activities to a file.

### Prerequisites

Before running the program, ensure you have:
- Go installed on your system.
- Access to a PostgreSQL database.

### Installation

Clone the repository to your local machine:

```bash
git clone https://github.com/ckazi/chunky.git
cd chunky
go mod init github.com/ckazi/chunky
go get github.com/jackc/pgx/v5
```

## Usage
To run the program, use the following command with necessary flags:
```bash
go run main.go -h [host] -p [port] -U [user] -pwd [password] -dbname [dbname] -table [table] -threads [number_of_threads] -file [output_file]
```

## Flags

Flags configure the program's behavior. Here's a detailed look at each:

Mandatory Flags
- -dbname: Specifies the name of the database. There is no default value, and this flag must be provided.
- -table: Specifies the table to check for corrupted rows. This flag is also mandatory and has no default value.

Optional Flags
- -h: Host address of the database server (default: "localhost").
- -p: Port number on which the database server is running (default: 5432).
- -U: Username for database access (default: "postgres").
- -pwd: Password for the specified user. It is optional and defaults to an empty string, but it is highly recommended for secure database access.
- -c: Column to order the data by (default: "id").
- -limit: Maximum number of rows to fetch per query (default: 5000).
- -offset: Starting offset for querying rows (default: 0).
- -threads: Number of threads to use for parallel processing (default: 8).
- -file: File path where results will be logged (default: "result.txt").
- -del: A boolean flag to enable deletion of corrupted rows; set this to true to delete rows (default: false).

## Example Commands

Here's how you might run the program in a typical scenario:

**To check the database without deleting any entries:**
```bash
go run main.go -dbname mydatabase -table usertable -U admin -pwd securepassword -file check_log.txt
```
**To run the program with deletion of corrupted rows enabled:**
```bash
go run main.go -dbname mydatabase -table usertable -U admin -pwd securepassword -del true -file deletion_log.txt
```

### Monitoring
![image](https://github.com/ckazi/chunky/assets/45271263/34b63594-1570-4eda-909f-85545386adf8)
You can see the process of the check by executing the following SQL query on your database:
```sql
SELECT pid, usename, application_name, client_addr, backend_start, query_start, state, query FROM pg_stat_activity WHERE state = 'active';
```


### Building the Project
```bash
git clone https://github.com/ckazi/chunky.git
cd chunky
go mod init github.com/ckazi/chunky.git
go get github.com/jackc/pgx/v5
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"'
```

  
  ### Sponsor
Are you enjoying this project?
[Buy me a beer!]
(monero **41ugNNZ5erdfj8ofHFhkb2gtwnpsB25digy6DWP1kCgRTJVbg6p7E6YMWbza7iCSMWaeuk9Qkeqzya8mCQcQDymH7P2tgZ5** ) 🍻
