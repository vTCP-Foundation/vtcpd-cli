# vTCP CLI Documentation

## Building the Project

This project uses a `Makefile` for common development tasks. Ensure you have `make` installed on your system.

**Available `make` commands:**

*   `make build`: Builds the project and places the binary in the `build` directory.
*   `make build-testing`: Builds the project in testing mode and places the binary in the `build` directory.
*   `make clean`: Removes the `build` directory.
*   `make test`: Runs all tests in the project.
*   `make run`: Builds and runs the application.
*   `make deps`: Downloads and installs Go module dependencies.
*   `make fmt`: Formats the Go source code.
*   `make lint`: Lints the Go source code using `golangci-lint`.
    *   Note: You need to have `golangci-lint` installed: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
*   `make config`: Copies the `conf.yaml` (example configuration) to the `build` directory.
*   `make all`: Runs `fmt`, `lint`, `test`, `build`, and `config` targets sequentially. This is a good command to run for a full check and build.

**Example Build Process:**
1.  Install dependencies: `make deps`
2.  Format and lint: `make fmt && make lint`
3.  Run tests: `make test`
4.  Build the binary: `make build`
    *   The executable will be located at `build/vtcpd-cli`.

## Command Line Interface (CLI)

General command format: `vtcpd-cli <command> [--type <sub-command>] [flags]`

### **Node Management Commands**

1.  **`start`**
    *   **Description:** Starts the vTCP node.
    *   **Flags:** None specific to the command itself. The node starts based on the configuration file.
    *   **Example:** `vtcpd-cli start`

2.  **`stop`**
    *   **Description:** Stops the vTCP node.
    *   **Flags:** None.
    *   **Example:** `vtcpd-cli stop`

3.  **`http`**
    *   **Description:** Starts the HTTP server for an already running vTCP node to manage it via API.
    *   **Flags:** None.
    *   **Example:** `vtcpd-cli http`

4.  **`start-http`**
    *   **Description:** Starts the vTCP node and then the HTTP server.
    *   **Flags:** None.
    *   **Example:** `vtcpd-cli start-http`

### **Node Interaction Commands**

For these commands, many flags are global and are set for use by internal handlers. The primary logic for differentiating actions is performed via the `--type` flag.

5.  **`channels`**
    *   **Description:** Manages payment channels.
    *   **Main Types (`--type`):** (The `--type <type>` flag is required to define the action for `channels` command)
        *   `init`: Initialize a new channel.
            *   **Flags:**
                *   `--address <address>`: Contractor address. Multiple can be specified.
        *   `list`: List all channels.
            *   **Flags:**
                *   `--offset <number>`: Offset for paginated output.
                *   `--count <number>`: Number of items to output.
        *   `info`: Information about a specific channel.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID or channel ID.
        *   `set-addresses`: Set/update addresses for a channel.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID or channel ID.
                *   `--address <address>`: Contractor address. Multiple can be specified.
        *   `set-crypto-key`: Set the cryptographic key for a channel.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID or channel ID.
                *   `--crypto-key <key>`: Cryptographic key.
        *   `regenerate-crypto-key`: Regenerate the cryptographic key for your side of the channel.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID or channel ID.
        *   `remove`: Delete/close a channel.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID or channel ID.
    *   **Examples:**
        *   Initialize a channel: `vtcpd-cli channels --type init --address ipv4:127.0.0.1:5001`
        *   List channels: `vtcpd-cli channels --type list --offset 0 --count 10`
        *   Channel information: `vtcpd-cli channels --type info --contractorID "channel-uuid"`
        *   Set key: `vtcpd-cli channels --type set-crypto-key --contractorID "channel-uuid" --crypto-key "new_key"`

6.  **`settlement-lines`**
    *   **Description:** Manages settlement lines.
    *   **Main Types (`--type`):** (The `--type <type>` flag is required to define the action for `settlement-lines` command)
        *   `init`: Initialize a new settlement line.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID for the new settlement line.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) for the new line.
        *   `set`: Set the maximum positive balance (contractor's debt to you).
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID of the settlement line to modify.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) of the line.
                *   `--amount <sum>`: Amount for setting the maximum positive balance.
        *   `close-incoming`: Close the incoming part of the line (zero out your debt to the contractor).
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID of the settlement line.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) of the line.
        *   `share-keys`: Exchange public keys.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID of the settlement line.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) of the line.
        *   `delete`: Delete the settlement line.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID of the settlement line.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) of the line.
        *   `reset`: Reset the state of the settlement line (audit, balances).
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID of the settlement line.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token) of the line.
                *   `--audit-number <number>`: Audit number for the reset.
                *   `--balance <sum>`: Balance for the reset (this is the `balance` field. Note: `reset` also requires `incoming_amount` and `outgoing_amount`, which are not currently set by separate flags, this might need adjustment or they are passed differently).
        *   `list-all`: List all settlement lines across all equivalents.
            *   **Flags:** (No specific flags for this type beyond `--type list-all`)
        *   `list-portions`: List settlement lines with pagination for a specific equivalent.
            *   **Flags:**
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token).
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of items to output.
        *   `list`: List settlement lines for a specific equivalent (usually the first page).
            *   **Flags:**
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token).
        *   `by-id`: Get a settlement line by its ID.
            *   **Flags:**
                *   `--contractorID <ID>`: Settlement line ID.
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token).
        *   `by-address`: Get a settlement line by the contractor's address.
            *   **Flags:**
                *   `--eq <equivalent_ID>`: Equivalent ID (currency/token).
                *   `--address <address>`: Contractor address.
    *   **Examples:**
        *   Initialize line: `vtcpd-cli settlement-lines --type init --contractorID "contractor-uuid" --eq 0`
        *   Set max positive balance: `vtcpd-cli settlement-lines --type set --contractorID "contractor-uuid" --eq 0 --amount 1000`
        *   List lines: `vtcpd-cli settlement-lines --type list --eq 0`
        *   Reset line: `vtcpd-cli settlement-lines --type reset --contractorID "contractor-uuid" --eq 0 --audit-number 1 --balance 0` (Note: passing `incoming_amount` and `outgoing_amount` for `reset` needs clarification).

7.  **`max-flow`**
    *   **Description:** Calculates the maximum flow.
    *   **Main Types (`--type`):** (The `--type <type>` flag is required to define the action for `max-flow` command)
        *   `calculate-fully`: Full calculation of maximum flow.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID.
                *   `--eq <equivalent_ID>`: Equivalent ID.
        *   `calculate-partly-step-1`: Partial calculation, step 1.
            *   **Flags:**
                *   `--contractorID <ID>`: Contractor ID.
                *   `--eq <equivalent_ID>`: Equivalent ID.
        *   `calculate-partly-step-2`: Partial calculation, step 2.
            *   **Flags:**
                *   `--channel-id-on-contractor-side <ID>`: Channel ID on the contractor's side.
    *   **Examples:**
        *   Calculate fully: `vtcpd-cli max-flow --type calculate-fully --contractorID "contractor-uuid" --eq 0`

8.  **`payment`**
    *   **Description:** Creates and sends a payment.
    *   **Flags Used:**
        *   `--contractorID <recipient_ID>`: Recipient contractor ID.
        *   `--eq <equivalent_ID>`: Equivalent ID.
        *   `--amount <sum>`: Payment amount.
        *   `--payload <data>`: (Optional) Additional data for the transaction.
    *   **Example:** `vtcpd-cli payment --contractorID "recipient-uuid" --eq 0 --amount 100 --payload "Order 123"`

9.  **`history`**
    *   **Description:** Views transaction history.
    *   **Main Types (`--type`):** (The `--type <type>` flag is required to define the action for `history` command)
        *   `payments`: Payment history for a specific equivalent.
            *   **Flags:**
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of records.
                *   `--eq <equivalent_ID>`: Equivalent ID.
                *   `--history-from <date>`: Start date of history (RFC3339 format, e.g., "2023-01-01T00:00:00Z").
                *   `--history-to <date>`: End date of history (RFC3339 format).
                *   `--amount-from <sum>`: Minimum amount for filtering.
                *   `--amount-to <sum>`: Maximum amount for filtering.
        *   `payments-all`: Payment history across all equivalents.
            *   **Flags:**
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of records.
                *   `--history-from <date>`: Start date of history (RFC3339 format, e.g., "2023-01-01T00:00:00Z").
                *   `--history-to <date>`: End date of history (RFC3339 format).
                *   `--amount-from <sum>`: Minimum amount for filtering.
                *   `--amount-to <sum>`: Maximum amount for filtering.
        *   `payments-additional`: Additional payment history.
            *   **Flags:**
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of records.
                *   `--eq <equivalent_ID>`: Equivalent ID.
                *   `--history-from <date>`: Start date of history (RFC3339 format, e.g., "2023-01-01T00:00:00Z").
                *   `--history-to <date>`: End date of history (RFC3339 format).
                *   `--amount-from <sum>`: Minimum amount for filtering.
                *   `--amount-to <sum>`: Maximum amount for filtering.
        *   `settlement-lines`: History of operations with settlement lines.
            *   **Flags:**
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of records.
                *   `--eq <equivalent_ID>`: Equivalent ID.
                *   `--history-from <date>`: Start date of history (RFC3339 format, e.g., "2023-01-01T00:00:00Z").
                *   `--history-to <date>`: End date of history (RFC3339 format).
                *   `--amount-from <sum>`: Minimum amount for filtering.
                *   `--amount-to <sum>`: Maximum amount for filtering.
        *   `contractor`: History of operations with a specific contractor.
            *   **Flags:**
                *   `--offset <number>`: Offset for pagination.
                *   `--count <number>`: Number of records.
                *   `--eq <equivalent_ID>`: Equivalent ID.
                *   `--history-from <date>`: Start date of history (RFC3339 format, e.g., "2023-01-01T00:00:00Z").
                *   `--history-to <date>`: End date of history (RFC3339 format).
                *   `--amount-from <sum>`: Minimum amount for filtering.
                *   `--amount-to <sum>`: Maximum amount for filtering.
                *   `--contractorID <ID>`: Contractor ID.
    *   **Examples:**
        *   Payment history: `vtcpd-cli history --type payments --eq 0 --offset 0 --count 20 --history-from "2023-10-01T00:00:00Z"`
        *   History by contractor: `vtcpd-cli history --type contractor --contractorID "contractor-uuid" --eq 0`

10. **`remove-outdated-crypto`**
    *   **Description:** Removes outdated cryptographic data from the node.
    *   **Flags:** None.
    *   **Example:** `vtcpd-cli remove-outdated-crypto`

## REST API Endpoints

### **Main API (`server.go`)**

*   **Equivalents**
    *   `GET /api/v1/node/equivalents/`
        *   **Description:** Lists all available equivalents (currencies/tokens).
        *   **Path Parameters:** None.
        *   **Query Parameters:** None explicitly in `server.go`, but the handler might support them.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/equivalents/`

*   **Channels**
    *   `POST /api/v1/node/contractors/init-channel/`
        *   **Description:** Initializes a new payment channel.
        *   **Request Body (JSON):** Expects parameters like contractor addresses (e.g., `{"addresses": ["ipv4:1.2.3.4:5000"]}`).
        *   **Example:** `curl -X POST -H "Content-Type: application/json" -d '{"addresses": ["ipv4:1.2.3.4:5000"]}' http://localhost:PORT/api/v1/node/contractors/init-channel/`
    *   `GET /api/v1/node/contractors/channels/`
        *   **Description:** Lists all payment channels.
        *   **Query Parameters:** Pagination might be supported (e.g., `offset=0&count=10`).
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/channels/?offset=0&count=10"`
    *   `GET /api/v1/node/channels/{contractor_id}/`
        *   **Description:** Detailed information about a specific channel.
        *   **Path Parameters:** `contractor_id` (Channel/Contractor ID). Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/channels/333/`
    *   `GET /api/v1/node/channel-by-address/`
        *   **Description:** Information about a channel by contractor address.
        *   **Query Parameters:** Expects `address` (e.g., `address=ipv4:1.2.3.4:5000`). Handled via `r.FormValue("address")` in the respective `routes` handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/channel-by-address/?address=ipv4:1.2.3.4:5000"`
    *   `PUT /api/v1/node/channels/{contractor_id}/set-addresses/`
        *   **Description:** Sets/updates addresses for a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Request Body (JSON):** Expects new addresses (e.g., `{"addresses": ["ipv4:new.ip:port"]}`).
        *   **Example:** `curl -X PUT -H "Content-Type: application/json" -d '{"addresses": ["ipv4:1.2.3.5:5001"]}' http://localhost:PORT/api/v1/node/channels/333/set-addresses/`
    *   `PUT /api/v1/node/channels/{contractor_id}/set-crypto-key/`
        *   **Description:** Sets the cryptographic key for a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Request Body (JSON):** Expects the key (e.g., `{"crypto_key": "contractor_key"}`).
        *   **Example:** `curl -X PUT -H "Content-Type: application/json" -d '{"crypto_key": "new_key"}' http://localhost:PORT/api/v1/node/channels/333/set-crypto-key/`
    *   `PUT /api/v1/node/channels/{contractor_id}/regenerate-crypto-key/`
        *   **Description:** Regenerates the cryptographic key for your side of the channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl -X PUT http://localhost:PORT/api/v1/node/channels/333/regenerate-crypto-key/`
    *   `DELETE /api/v1/node/channels/{contractor_id}/remove/`
        *   **Description:** Deletes/closes a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/channels/333/remove/`

*   **Contractors**
    *   `GET /api/v1/node/contractors/{equivalent}/`
        *   **Description:** Lists contractors for the specified equivalent.
        *   **Path Parameters:** `equivalent` (Equivalent ID). Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/0/`

*   **Settlement Lines**
    *   `GET /api/v1/node/contractors/settlement-lines/{equivalent}/`
        *   **Description:** Lists settlement lines for the specified equivalent.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/0/`
    *   `GET /api/v1/node/contractors/settlement-lines/{offset}/{count}/{equivalent}/`
        *   **Description:** Lists settlement lines with pagination.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/0/10/0/`
    *   `GET /api/v1/node/contractors/settlement-lines/equivalents/all/`
        *   **Description:** Lists settlement lines across all equivalents.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/equivalents/all/`
    *   `GET /api/v1/node/contractors/settlement-line-by-id/{equivalent}/`
        *   **Description:** Gets a settlement line by its ID.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `id` (Settlement line ID). Handled via `r.FormValue("id")` in the handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/settlement-line-by-id/0/?id=333"`
    *   `GET /api/v1/node/contractors/settlement-line-by-address/{equivalent}/`
        *   **Description:** Gets a settlement line by contractor address.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `addresses` (Contractor addresses, comma-separated). Handled via `r.FormValue("addresses")` in the handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/settlement-line-by-address/0/?addresses=ipv4:1.2.3.4:5000"`
    *   `POST /api/v1/node/contractors/{contractor_id}/init-settlement-line/{equivalent}/`
        *   **Description:** Initializes a new settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Body (JSON):** May require `max_positive_balance`, `max_negative_balance`.
        *   **Example:** `curl -X POST -H "Content-Type: application/json" -d '{"max_positive_balance": "1000"}' http://localhost:PORT/api/v1/node/contractors/333/init-settlement-line/0/`
    *   `PUT /api/v1/node/contractors/{contractor_id}/settlement-lines/{equivalent}/` (Handler `SetMaxPositiveBalance`)
        *   **Description:** Sets the maximum positive balance for a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Body (JSON):** Expects `amount` (e.g., `{"amount": "2000"}`).
        *   **Example:** `curl -X PUT -H "Content-Type: application/json" -d '{"amount": "2000"}' http://localhost:PORT/api/v1/node/contractors/333/settlement-lines/0/`
    *   `DELETE /api/v1/node/contractors/{contractor_id}/close-incoming-settlement-line/{equivalent}/` (Handler `ZeroOutMaxNegativeBalance`)
        *   **Description:** Closes the incoming part of a settlement line (zeros out max negative balance).
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/contractors/333/close-incoming-settlement-line/0/`
    *   `PUT /api/v1/node/contractors/{contractor_id}/keys-sharing/{equivalent}/` (Handler `PublicKeysSharing`)
        *   **Description:** Initiates public key exchange for a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X PUT http://localhost:PORT/api/v1/node/contractors/333/keys-sharing/0/`
    *   `DELETE /api/v1/node/contractors/{contractor_id}/remove-settlement-line/{equivalent}/`
        *   **Description:** Deletes a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/contractors/333/remove-settlement-line/0/`
    *   `PUT /api/v1/node/contractors/{contractor_id}/reset-settlement-line/{equivalent}/`
        *   **Description:** Resets the state of a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Body (JSON):** Expects `audit_number`, `max_negative_balance`, `max_positive_balance`, `balance`.
        *   **Example:** `curl -X PUT -H "Content-Type: application/json" -d '{"audit_number":"1","max_negative_balance":"0","max_positive_balance":"500","balance":"-500"}' http://localhost:PORT/api/v1/node/contractors/333/reset-settlement-line/0/`

*   **Stats**
    *   `GET /api/v1/node/stats/total-balance/{equivalent}/`
        *   **Description:** Total balance for the specified equivalent.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/stats/total-balance/0/`

*   **History**
    *   `GET /api/v1/node/history/transactions/payments/{offset}/{count}/{equivalent}/`
        *   **Description:** Payment history.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`. Handled via `r.FormValue(...)`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/payments/0/10/0/?date_from=2023-01-01T00:00:00Z&amount_from=50"`
    *   `GET /api/v1/node/history/transactions/payments-all/{offset}/{count}/`
        *   **Description:** Payment history across all equivalents.
        *   **Path Parameters:** `offset`, `count`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/payments-all/0/20/"`
    *   `GET /api/v1/node/history/transactions/payments/additional/{offset}/{count}/{equivalent}/`
        *   **Description:** Additional payment history.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/payments/additional/0/5/0/"`
    *   `GET /api/v1/node/history/transactions/settlement-lines/{offset}/{count}/{equivalent}/`
        *   **Description:** History of operations with settlement lines.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/settlement-lines/0/15/0/"`
    *   `GET /api/v1/node/history/contractors/{offset}/{count}/{equivalent}/`
        *   **Description:** History of operations with a contractor.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `contractor_id`, `date_from`, `date_to`, `amount_from`, `amount_to`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/contractors/0/10/0/?contractor_id=333"`

*   **Optimization**
    *   `DELETE /api/v1/node/remove-outdated-crypto/`
        *   **Description:** Removes outdated cryptographic data.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/remove-outdated-crypto/`
    *   `POST /api/v1/node/regenerate-all-keys/`
        *   **Description:** Regenerates all cryptographic keys for the node.
        *   **Example:** `curl -X POST http://localhost:PORT/api/v1/node/regenerate-all-keys/`

*   **Control**
    *   `POST /api/v1/ctrl/stop/`
        *   **Description:** Stops the node and all its operations.
        *   **Example:** `curl -X POST http://localhost:PORT/api/v1/ctrl/stop/`

### **Testing API (`server_testing.go`). Can be used only in testing build mode**

*   `PUT /api/v1/node/subsystems-controller/{flags}/`
    *   **Description:** Sets testing flags for the subsystems controller.
    *   **Path Parameters:** `flags` (String representing the flags). Parsed via `mux.Vars(r)["flags"]`.
    *   **Example:** `curl -X PUT http://localhost:TEST_PORT/api/v1/node/subsystems-controller/some_flags_value`
*   `PUT /api/v1/node/settlement-lines-influence/{flags}/`
    *   **Description:** Sets testing flags for settlement lines influence.
    *   **Path Parameters:** `flags`. Parsed via `mux.Vars(r)["flags"]`.
    *   **Example:** `curl -X PUT http://localhost:TEST_PORT/api/v1/node/settlement-lines-influence/another_flags_value`
*   `PUT /api/v1/node/make-node-busy/`
    *   **Description:** Puts the node into a "busy" state for testing purposes.
    *   **Example:** `curl -X PUT http://localhost:TEST_PORT/api/v1/node/make-node-busy/`