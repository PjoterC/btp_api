# Token transfer API

A GraphQL API for transferring tokens between virtual wallets, using the gqlgen package. 
The API implements a `transfer` mutation to send tokens from one wallet to another. The conditions of successful transfer are:
- Balance check - balance of the sending wallet after transfer must not be negative.
- Both sender and receiver wallet addresses must exist in the database.
- Transfer amount must be at least 1 token or more.

The transfer mutation schema is as follows:
transfer(from_address: String!, to_address: String!, amount: Int!): Int!

After a successful transfer, the amount of tokens remaining on the source wallet will be returned.

## Installation and running the api (Docker and go language required)

After cloning or downloading the repository, run the command line inside the `BTP_API` directory. First create the container for the database, from the docker-compose file by using the `docker-compose up -d` command in the command line. After pulling the postgres image, a container should be created and automatically run. (NOTE - container is made to restart automatically whenever possible. To change this, edit the docker-compose "restart" section.)

To use the server application, simply enter `go build server.go` and then run the created executable. Provided the database is running, the server should create the initial table for wallets, with the starting wallet as specified by the task requirements.

Both server and database are running on localhost, database on port 5432, while server on port 8080. Ports can be changed in docker-compose and .env file.

## Using the API

There are few possible ways to use the API. 

The first one is to use the default GraphQL playground, provided by the library. It can be accessed by entering the http://localhost:8080/ in the web browser. To initiate a transfer, type a mutation (syntax shown below) on the provided page and press the "execute query" button.

### Examples of mutations:

mutation {
  transfer(from_address: "source", to_address: "destination", amount: 10)
} 

mutation {
  transfer(from_address: "walletA", to_address: "walletB", amount: 1984)
}

mutation {
  transfer(from_address: "testSourceA", to_address: "testDestA", amount: -13)
} 
//example of incorrect transfer - can only transfer positive amount of tokens

`More information about automated tests in 🗀/tests/TEST_INSTRUCTIONS.md`


Alternatively, one can use a web communication tools like Postman, Thunder Client or curl(NOT RECOMMENDED) to send a GraphQL request to the http://localhost:8080/query directly. Most clients have built-in GraphQL editors, so the syntax of the query is identical to the one used in the GraphQL playground.