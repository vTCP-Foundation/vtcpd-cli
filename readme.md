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
        *   `--address <address>`: Contractor address. Multiple can be specified.
        *   `--eq <equivalent_ID>`: Equivalent ID.
        *   `--amount <sum>`: Payment amount.
        *   `--payload <data>`: (Optional) Additional data for the transaction.
    *   **Example:** `vtcpd-cli payment --address "ipv4:1.2.3.4:5678" --eq 0 --amount 100 --payload "Order 123"`

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

### Address Format
Addresses in the API use the following format: `<type_code>-<address>`
* `type_code`: Numeric code representing the address type
  * `12`: IPv4 address
* `address`: The actual address (e.g., IP:port for IPv4)

### **Main API (`server.go`)**

*   **Equivalents**
    *   `GET /api/v1/node/equivalents/`
        *   **Description:** Lists all available equivalents (currencies/tokens).
        *   **Path Parameters:** None.
        *   **Query Parameters:** None explicitly in `server.go`, but the handler might support them.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/equivalents/`
        *   **Response:** JSON object containing a count and a list of equivalents.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 2,
                    "equivalents": ["0", "1"]
                }
            }
            ```

*   **Channels**
    *   `POST /api/v1/node/contractors/init-channel/`
        *   **Description:** Initializes a new payment channel. This is a two-step process for creating a channel between two participants:
            1.  **Initiator:** One participant calls this endpoint, passing the `contractor_address` parameter(s) in the query string (can be repeated for multiple addresses). The `crypto_key` and `contractor_id` parameters are not passed. The response returns the `channel_id` and `crypto_key`.
            2.  **Participant:** The second participant calls this endpoint, passing the same `contractor_address`, as well as the `crypto_key` and `contractor_id` parameters (obtained in the first step). Only after this will the channel be opened.
        *   **Request Parameters (query):**
            *   `contractor_address` (required, can be repeated)
            *   `crypto_key` (optional, only for the second step)
            *   `contractor_id` (optional, only for the second step)
        *   **Example (Step 1 - Initiator):**
            ```bash
            curl -X POST "http://localhost:PORT/api/v1/node/contractors/init-channel/?contractor_address=12-5.6.7.8:5001"
            ```
        *   **Example (Step 2 - Participant):**
            ```bash
            curl -X POST "http://localhost:PORT/api/v1/node/contractors/init-channel/?contractor_address=12-1.2.3.4:5000&crypto_key=initiator_public_key&contractor_id=channel-uuid-from-initiator"
            ```
        *   **Response (Step 1):**
            ```json
            {
              "data": {
                "channel_id": "channel-uuid-123",
                "crypto_key": "initiator_public_key"
              }
            }
            ```
        *   **Response (Step 2):**
            ```json
            {
              "data": {
                "channel_id": "channel-uuid-456",
                "crypto_key": "participant_public_key"
              }
            }
            ```
    *   `GET /api/v1/node/contractors/channels/`
        *   **Description:** Lists all payment channels.
        *   **Query Parameters:** Pagination might be supported (e.g., `offset=0&count=10`).
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/channels/?offset=0&count=10"`
        *   **Response:** JSON object containing a count and a list of channels.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "channels": [
                        {
                            "channel_id": "channel-uuid-123",
                            "channel_addresses": "12-1.2.3.4:5000,12-5.6.7.8:5001"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/channels/{contractor_id}/`
        *   **Description:** Detailed information about a specific channel.
        *   **Path Parameters:** `contractor_id` (Channel/Contractor ID). Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/channels/333/`
        *   **Response:** JSON object containing detailed channel information including addresses, crypto keys, and status.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "channel_id": "channel-uuid-123",
                    "channel_addresses": ["12-1.2.3.4:5000"],
                    "channel_confirmed": "true",
                    "channel_crypto_key": "self_crypto_key",
                    "channel_contractor_crypto_key": "contractor_crypto_key"
                }
            }
            ```
    *   `GET /api/v1/node/channel-by-address/`
        *   **Description:** Information about a channel by contractor address.
        *   **Query Parameters:** Expects `contractor_address` (e.g., `contractor_address=12-1.2.3.4:5000`). Handled via `r.FormValue("contractor_address")` in the respective `routes` handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/channel-by-address/?contractor_address=12-1.2.3.4:5000"`
        *   **Response:** JSON object containing channel information for the specified address.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "channel_id": "channel-uuid-123",
                    "channel_confirmed": "true"
                }
            }
            ```
    *   `PUT /api/v1/node/channels/{contractor_id}/set-addresses/`
        *   **Description:** Sets/updates addresses for a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Request Parameters (query):**
            *   `contractor_address` (required, can be repeated)
        *   **Example:**
            `curl -X PUT "http://localhost:PORT/api/v1/node/channels/333/set-addresses/?contractor_address=12-1.2.3.5:5001"`
        *   **Response:** JSON object with success status and updated channel information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "channel_id": "channel-uuid-123",
                    "channel_addresses": ["12-1.2.3.5:5001"],
                    "channel_confirmed": "true",
                    "channel_crypto_key": "self_crypto_key",
                    "channel_contractor_crypto_key": "contractor_crypto_key"
                }
            }
            ```
    *   `PUT /api/v1/node/channels/{contractor_id}/set-crypto-key/`
        *   **Description:** Sets the cryptographic key for a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Request Parameters (query):**
            *   `crypto_key` (required)
            *   `channel_id_on_contractor_side` (optional)
        *   **Example:**
            `curl -X PUT "http://localhost:PORT/api/v1/node/channels/333/set-crypto-key/?crypto_key=new_key"`
        *   **Response:** JSON object with success status and updated channel information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "channel_id": "channel-uuid-123",
                    "channel_addresses": ["12-1.2.3.4:5000"],
                    "channel_confirmed": "true",
                    "channel_crypto_key": "new_key",
                    "channel_contractor_crypto_key": "contractor_key_if_known_or_unchanged"
                }
            }
            ```
    *   `PUT /api/v1/node/channels/{contractor_id}/regenerate-crypto-key/`
        *   **Description:** Regenerates the cryptographic key for your side of the channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl -X PUT http://localhost:PORT/api/v1/node/channels/333/regenerate-crypto-key/`
        *   **Response:** JSON object with success status and new crypto key information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "channel_id": "channel-uuid-123",
                    "channel_addresses": ["12-1.2.3.4:5000"],
                    "channel_confirmed": "true",
                    "channel_crypto_key": "new_regenerated_key",
                    "channel_contractor_crypto_key": "contractor_crypto_key"
                }
            }
            ```
    *   `DELETE /api/v1/node/channels/{contractor_id}/remove/`
        *   **Description:** Deletes/closes a channel.
        *   **Path Parameters:** `contractor_id`. Parsed via `mux.Vars(r)["contractor_id"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/channels/333/remove/`
        *   **Response:** JSON object with success status and confirmation of channel removal.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "status": "success",
                    "msg": "Channel removed successfully."
                }
            }
            ```

*   **Contractors**
    *   `GET /api/v1/node/contractors/{equivalent}/`
        *   **Description:** Lists contractors for the specified equivalent.
        *   **Path Parameters:** `equivalent` (Equivalent ID). Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/0/`
        *   **Response:** JSON object containing a count and a list of contractors.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "contractors": [
                        {
                            "contractor_id": "contractor-uuid-abc",
                            "contractor_addresses": "12-1.2.3.4:5000"
                        }
                    ]
                }
            }
            ```

*   **Settlement Lines**
    *   `GET /api/v1/node/contractors/settlement-lines/{equivalent}/`
        *   **Description:** Lists settlement lines for the specified equivalent.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/0/`
        *   **Response:** JSON object containing a count and a list of settlement lines.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "settlement_lines": [
                        {
                            "contractor_id": "contractor-uuid-abc",
                            "contractor": "Contractor Name/ID",
                            "state": "Active",
                            "own_keys_present": "true",
                            "contractor_keys_present": "true",
                            "max_negative_balance": "1000",
                            "max_positive_balance": "5000",
                            "balance": "500"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/contractors/settlement-lines/{offset}/{count}/{equivalent}/`
        *   **Description:** Lists settlement lines with pagination.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/0/10/0/`
        *   **Response:** JSON object containing a count and a list of settlement lines with pagination metadata.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "settlement_lines": [
                        {
                            "contractor_id": "contractor-uuid-abc",
                            "contractor": "Contractor Name/ID",
                            "state": "Active",
                            "own_keys_present": "true",
                            "contractor_keys_present": "true",
                            "max_negative_balance": "1000",
                            "max_positive_balance": "5000",
                            "balance": "500"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/contractors/settlement-lines/equivalents/all/`
        *   **Description:** Lists settlement lines across all equivalents.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/contractors/settlement-lines/equivalents/all/`
        *   **Response:** JSON object containing a count and a list of settlement lines grouped by equivalent.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "equivalents": [
                        {
                            "equivalent": "0",
                            "count": 1,
                            "settlement_lines": [
                                {
                                    "contractor_id": "contractor-uuid-abc",
                                    "contractor": "Contractor Name/ID",
                                    "state": "Active",
                                    "own_keys_present": "true",
                                    "contractor_keys_present": "true",
                                    "max_negative_balance": "1000",
                                    "max_positive_balance": "5000",
                                    "balance": "500"
                                }
                            ]
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/contractors/settlement-line-by-id/{equivalent}/`
        *   **Description:** Gets a settlement line by its ID.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `id` (Settlement line ID). Handled via `r.FormValue("id")` in the handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/settlement-line-by-id/0/?id=333"`
        *   **Response:** JSON object containing settlement line details.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "Active",
                        "own_keys_present": "true",
                        "contractor_keys_present": "true",
                        "audit_number": "5",
                        "max_negative_balance": "1000",
                        "max_positive_balance": "5000",
                        "balance": "500"
                    }
                }
            }
            ```
    *   `GET /api/v1/node/contractors/settlement-line-by-address/{equivalent}/`
        *   **Description:** Gets a settlement line by contractor address.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Request Parameters (query):**
            *   `contractor_address` (Contractor addresses, can be repeated). Handled via `r.FormValue("contractor_address")` in the handler.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/contractors/settlement-line-by-address/0/?contractor_address=12-1.2.3.4:5000"`
        *   **Response:** JSON object containing settlement line details for the specified address.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "Active",
                        "own_keys_present": "true",
                        "contractor_keys_present": "true",
                        "audit_number": "5",
                        "max_negative_balance": "1000",
                        "max_positive_balance": "5000",
                        "balance": "500"
                    }
                }
            }
            ```
    *   `POST /api/v1/node/contractors/{contractor_id}/init-settlement-line/{equivalent}/`
        *   **Description:** Initializes a new settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Parameters (query):**
            *   `max_positive_balance` (optional)
            *   `max_negative_balance` (optional)
        *   **Example:**
            `curl -X POST "http://localhost:PORT/api/v1/node/contractors/333/init-settlement-line/0/?max_positive_balance=1000"`
        *   **Response:** JSON object containing the created settlement line information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "new-line-uuid-123",
                        "state": "Pending",
                        "own_keys_present": "false",
                        "contractor_keys_present": "false",
                        "audit_number": "0",
                        "max_negative_balance": "0",
                        "max_positive_balance": "1000",
                        "balance": "0"
                    }
                }
            }
            ```
    *   `PUT /api/v1/node/contractors/{contractor_id}/settlement-lines/{equivalent}/` (Handler `SetMaxPositiveBalance`)
        *   **Description:** Sets the maximum positive balance for a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Parameters (query):**
            *   `amount` (required)
        *   **Example:**
            `curl -X PUT "http://localhost:PORT/api/v1/node/contractors/333/settlement-lines/0/?amount=2000"`
        *   **Response:** JSON object with success status and updated balance information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "Active",
                        "own_keys_present": "true",
                        "contractor_keys_present": "true",
                        "audit_number": "5",
                        "max_negative_balance": "1000",
                        "max_positive_balance": "2000",
                        "balance": "500"
                    }
                }
            }
            ```
    *   `DELETE /api/v1/node/contractors/{contractor_id}/close-incoming-settlement-line/{equivalent}/` (Handler `ZeroOutMaxNegativeBalance`)
        *   **Description:** Closes the incoming part of a settlement line (zeros out max negative balance).
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/contractors/333/close-incoming-settlement-line/0/`
        *   **Response:** JSON object with success status and updated balance information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "Active",
                        "own_keys_present": "true",
                        "contractor_keys_present": "true",
                        "audit_number": "6", 
                        "max_negative_balance": "0",
                        "max_positive_balance": "2000",
                        "balance": "500" 
                    }
                }
            }
            ```
    *   `PUT /api/v1/node/contractors/{contractor_id}/keys-sharing/{equivalent}/` (Handler `PublicKeysSharing`)
        *   **Description:** Initiates public key exchange for a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X PUT http://localhost:PORT/api/v1/node/contractors/333/keys-sharing/0/`
        *   **Response:** JSON object with success status and key sharing information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "KeysSharing",
                        "own_keys_present": "true",
                        "contractor_keys_present": "false", 
                        "audit_number": "6",
                        "max_negative_balance": "0",
                        "max_positive_balance": "2000",
                        "balance": "500"
                    }
                }
            }
            ```
    *   `DELETE /api/v1/node/contractors/{contractor_id}/remove-settlement-line/{equivalent}/`
        *   **Description:** Deletes a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl -X DELETE http://localhost:PORT/api/v1/node/contractors/333/remove-settlement-line/0/`
        *   **Response:** JSON object with success status and confirmation of removal.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "status": "success",
                    "msg": "Settlement line removed successfully."
                }
            }
            ```
    *   `PUT /api/v1/node/contractors/{contractor_id}/reset-settlement-line/{equivalent}/`
        *   **Description:** Resets the state of a settlement line.
        *   **Path Parameters:** `contractor_id`, `equivalent`. Parsed via `mux.Vars(r)["contractor_id"]`, `mux.Vars(r)["equivalent"]`.
        *   **Request Parameters (query):**
            *   `audit_number` (required)
            *   `max_negative_balance` (required)
            *   `max_positive_balance` (required)
            *   `balance` (required)
        *   **Example:**
            `curl -X PUT "http://localhost:PORT/api/v1/node/contractors/333/reset-settlement-line/0/?audit_number=1&max_negative_balance=0&max_positive_balance=500&balance=-500"`
        *   **Response:** JSON object with success status and updated settlement line state.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "settlement_line": {
                        "id": "line-uuid-xyz",
                        "state": "Active",
                        "own_keys_present": "true",
                        "contractor_keys_present": "true",
                        "audit_number": "1",
                        "max_negative_balance": "0",
                        "max_positive_balance": "500",
                        "balance": "-500"
                    }
                }
            }
            ```

*   **Transactions**
    *   `POST /api/v1/node/contractors/transactions/{equivalent}/`
        *   **Description:** Creates and sends a payment (transaction).
        *   **Path Parameters:** `equivalent` (Equivalent/currency ID).
        *   **Request Parameters (query):**
            *   `contractor_address` (required, recipient contractor address)
            *   `amount` (required, payment amount)
            *   `payload` (optional, additional transaction data)
        *   **Example:** `curl -X POST "http://localhost:PORT/api/v1/node/contractors/transactions/0/?contractor_address=12-1.2.3.4.:5000&amount=100&payload=Order123"`
        *   **Response:** JSON object containing the transaction UUID.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "transaction_uuid": "tx-uuid-abcdef"
                }
            }
            ```
    *   `GET /api/v1/node/contractors/transactions/max/{equivalent}/`
        *   **Description:** Calculates the maximum flow for the specified equivalent (likely for *all* contractors or for one specified via query).
        *   **Path Parameters:** `equivalent` (Equivalent/currency ID).
        *   **Request Parameters (query):**
            *   `contractor_address` (contractor address, can be repeted)
        *   **Example:** `curl -X GET "http://localhost:PORT/api/v1/node/contractors/transactions/max/0/?contractor_address=12-1.2.3.4:500"`
        *   **Response:** JSON object containing the max flow calculation results.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "records": [
                        {
                            "address_type": "12",
                            "contractor_address": "1.2.3.4:5000",
                            "max_amount": "10000"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/transactions/{command_uuid}/`
        *   **Description:** Gets the transaction status by the UUID of the command that initiated it (e.g., the UUID returned in the POST request to create the transaction).
        *   **Path Parameters:** `command_uuid` (Command UUID).
        *   **Example:** `curl -X GET http://localhost:PORT/api/v1/node/transactions/cmd-uuid-12345/`
        *   **Response:** JSON object containing the transaction UUID if found.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1, 
                    "transaction_uuid": "tx-uuid-abcdef"
                }
            }
            ```

*   **Stats**
    *   `GET /api/v1/node/stats/total-balance/{equivalent}/`
        *   **Description:** Total balance for the specified equivalent.
        *   **Path Parameters:** `equivalent`. Parsed via `mux.Vars(r)["equivalent"]`.
        *   **Example:** `curl http://localhost:PORT/api/v1/node/stats/total-balance/0/`
        *   **Response:** JSON object containing total balance information.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "total_max_negative_balance": "10000",
                    "total_negative_balance": "2000",
                    "total_max_positive_balance": "15000",
                    "total_positive_balance": "3000"
                }
            }
            ```

*   **History**
    *   `GET /api/v1/node/history/transactions/payments/{offset}/{count}/{equivalent}/`
        *   **Description:** Payment history.
        *   **Path Parameters:** `offset`, `count`, `equivalent`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`, `mux.Vars(r)["equivalent"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`. Handled via `r.FormValue(...)`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/payments/0/10/0/?date_from=2023-01-01T00:00:00Z&amount_from=50"`
        *   **Response:** JSON object containing a count and a list of payment transaction objects with pagination metadata.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "records": [
                        {
                            "transaction_uuid": "tx-uuid-1",
                            "unix_timestamp_microseconds": "1678886400000000",
                            "contractor": "contractor-uuid-abc",
                            "operation_direction": "outgoing",
                            "amount": "100",
                            "balance_after_operation": "400",
                            "payload": "Order 123"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/history/transactions/payments-all/{offset}/{count}/`
        *   **Description:** Payment history across all equivalents.
        *   **Path Parameters:** `offset`, `count`. Parsed via `mux.Vars(r)["offset"]`, `mux.Vars(r)["count"]`.
        *   **Query Parameters:** `date_from`, `date_to`, `amount_from`, `amount_to`.
        *   **Example:** `curl "http://localhost:PORT/api/v1/node/history/transactions/payments-all/0/20/"`
        *   **Response:** JSON object containing a count and a list of payment transaction objects across all equivalents with pagination metadata.
        *   **Response Body (JSON Example):**
            ```json
            {
                "data": {
                    "count": 1,
                    "records": [
                        {
                            "equivalent": "0",
                            "transaction_uuid": "tx-uuid-2",
                            "unix_timestamp_microseconds": "1678886500000000",
                            "contractor": "contractor-uuid-xyz",
                            "operation_direction": "incoming",
                            "amount": "50",
                            "balance_after_operation": "450",
                            "payload": "Invoice 456"
                        }
                    ]
                }
            }
            ```
    *   `GET /api/v1/node/history/transactions/payments/additional/{offset}/{count}/{equivalent}/`

### **Testing API (`server_testing.go`). Can be used only in testing build mode**

*   `PUT /api/v1/node/subsystems-controller/{flags}/`
    *   **Description:** Sets testing flags for the subsystems controller.
    *   **Path Parameters:** `flags` (String representing the flags).
    *   **Request Parameters (query):**
        *   `forbidden_address` (optional)
        *   `forbidden_amount` (optional)
    *   **Example:** `curl -X PUT "http://localhost:TEST_PORT/api/v1/node/subsystems-controller/some_flags_value?forbidden_address=12-1.2.3.4:5000&forbidden_amount=100"`
    *   **Response:** JSON object with success status and applied flags.
    *   **Response Body (JSON Example):**
        ```json
        {
            "data": {
                "status": "success",
                "msg": "Flags 'some_flags_value' applied to subsystems controller."
            }
        }
        ```
*   `PUT /api/v1/node/settlement-lines-influence/{flags}/`
    *   **Description:** Sets testing flags for settlement lines influence.
    *   **Path Parameters:** `flags` (String representing the flags).
    *   **Request Parameters (query):**
        *   `first_parameter` (optional)
        *   `second_parameter` (optional)
        *   `third_parameter` (optional)
    *   **Example:** `curl -X PUT "http://localhost:TEST_PORT/api/v1/node/settlement-lines-influence/another_flags_value?first_parameter=val1&second_parameter=val2"`
    *   **Response:** JSON object with success status and applied flags.
    *   **Response Body (JSON Example):**
        ```json
        {
            "data": {
                "status": "success",
                "msg": "Flags 'another_flags_value' applied to settlement lines influence."
            }
        }
        ```