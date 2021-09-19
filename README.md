# cloudStorage
A simple cloud storage implementation in Go

Install and run: (go v 1.17)
mysql database should include a "cloudStorage" schema.
Execute cmd/migrate to set up database schema.
> go run github.com/EatonEmmerich/cloudStorage/cmd/migrate@latest
> go run github.com/EatonEmmerich/cloudStorage@latest

App will attempt to connect on port 3306 for mysql server and serve on port 8084 by default
