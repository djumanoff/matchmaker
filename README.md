# Match Maker

**Match maker** is a small web application to build teams, based on the rating of the players, which was calculated by rating each other. 

## Requirements

- Go Runtime
- MySQL database

## Getting Started

1. Go to root directory
2. Run sql init queries in `init.sql` file
3. Run:
 
    `go get`
     
4. Edit `config.env` file then run command: 

    `go run main.go -c=config.env`
    