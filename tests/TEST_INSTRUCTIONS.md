## How to run tests

The automated tests use the default `testing` go package and create (and reset) their own test wallets for sending mock transfers, using helper functions from 🗀/helpers. 

The test are separated into 2 files - tests with expectation of success (🗀/correcttransfers_test.go) and tests with expectation of failure (🗀/incorrecttransfers_test.go).

To run the tests, run command line in this directory and type `go test ./.`, which will run all available tests. To show all the test logs, the "-v" flag can be added to the command like this: `go test -v ./.`. To run specific test, you type: `go test (-v) -run NameOfSpecificTest`.

## Available tests:

### Correct transfers
These tests expect successful transfers and valid state of the wallets in the database at the end.

- TestValidTransfer - The test performs a valid transfer and checks for success.
- TestSimultanous - The test attempts to perform two simultaneous valid transfers and checks for success.
- TestDeadlock - The test performs transfers to check whether wallets can simultanously transfer to each other, without causing deadlock.

### Incorrect transfers
These tests "expect failure", meaning that the tests can be successful even if an error occurs, provided it is within the expected range and the reason for transfer failure is valid.

- TestSimultanousOverBalance - The test attempts to perform two simultaneous transfers that together exceed the source wallet's balance. One transfer should finish successfully, while the other will be stopped by balance check.

- TestExampleRaceCondition -  The test performs three transfers in parallel, with one of them potentially failing due to insufficient funds - an implementation of the example from the task sheet.

- TestNonExistingWallets - The test attempts to perform transfers involving non-existing wallets. Test is successful if the transfer with non-existing source wallet fails.

- TestNegativeAmount - The test attempts to perform a transfer with a negative amount. Test is successful if transfer fails.

