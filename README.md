# Banano Site API


Easy [BANANO](https://banano.cc/) payments on the [Banano Site](https://banano.site/) project pages.

*banano-site-api* is a server program that hold username and banano address KV pairs.
Registering new user accounts and query for information by username is provided.

## Compile

 - *banano-site-api* is written in Go. You can compile it for you **current platform**:

   ```$ make build```
   
- or **Cross-compile** it for *linux_amd64*, *linux_arm* or *linux_arm64* platforms with:
 
   ```$ make all-build```
   
*TODO: docker images, binary releases*

## Running

 - Create configuration file (*banano.json* sample config file is provided):
 
 ```json
 {
  "AppPort":"8080",    # port the daemon will run on
  "AppDb":"banano.db", # KV store for accounts - /path/to/some/file.db
  "AppUser":"api",     # Basic authentication user for some actions
  "AppPass":"secret"   # Basic authentication password for some actions
}
 ```

 - Start the *banano-site-api* daemon (example is  forOSX). This will also create DB:
 
 ```bash
 $ ./bin/darwin_amd64/banano-site-api
 ```
 
 - Check port *8080* for the config above:
 
 ```
 $ curl http://api:secret@127.0.0.1:8080/api/v1/accounts
[]
 ```


## How it works?

 *banano-site-api* is a HTTP server with the following endpoints (see *app.go* for details):
 
### Public (non-protected) endpoints

 - **GET /u/:username** - get user information in **JSON** format - *username* and *address*
 
 ```
 $ curl http://127.0.0.1:8080/u/somebody

 {"username":"somebody","address":"ban_12345"}
 ```

 - **POST /api/v1/account** - save user information (input data in **JSON** format)
 
 ```
 $ curl --header "Content-Type: application/json" \
 --request POST \
 --user api:secret \
 --data '{"username":"somebody","address":"ban_12345"}' \
 http://127.0.0.1:8080/api/v1/account
  
 {"username":"somebody","address":"ban_12345"}
 ```
  
 
### Basic Authentication protected maintenance endpoints

 - **GET /api/v1/accounts** - List all saved users
 
 ```
 $ curl http://api:secret@127.0.0.1:8080/api/v1/accounts
 
 [{"username":"somebody","address":"ban_12345"}, ...]
 ```
 - **PUT /api/v1/account/:username** - Update user address information
 
 ```
 $ curl --header "Content-Type: application/json" \
 --request PUT \
 --user api:secret \
 --data '{"address":"ban_1333555"}' \
 http://127.0.0.1:8080/api/v1/account/somebody
 
 {"username":"somebody","address":"ban_1333555"}
 ```
 
 - **DELETE /api/v1/account/:username** - Delete user account
 
 ```
 $curl -X DELETE http://api:secret@127.0.0.1:8080/api/v1/account/somebody
 
 {"result":"success"}
 ```

## Security

 - Set proper username and strong enough password in you config file
 - Very simple requests rate limiting is provided
 - No other information, than *username* and *banano address* are stored in the DB

## Contributing

 - Please open an issue or PR if you have a question or suggestion.
