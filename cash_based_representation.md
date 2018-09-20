# Cash-based representation
Mission of this work was to have a first implementation of a cash-based payment system, where each USC standard denom is uniquely identified by an ID and so that the USC balance of a user is given by aggregation of the set of standard denoms it owns. We made use of precompiles contracts to improve efficiency. Initially, we have assumed to have only token of one standard denom, but this construction can be simply generalized. In the storage each token is represented by 32bytes, in particular the first byte indicate the type of denom and remaings represents the unique id. Tokens are managed like last in first out queue.

## Type precompiled contracts
Here listed precompiled we have defined so far:

- `DataContract`, entry point for all USC balances.
- `USC_Fund`, creates a new set of tokens such they sum up to the amount;
- `USC_Transfer, moves a set of tokens from an account to another;
- `USC_Defund, removes a set of tokens from an account;

Note: In order to `DataContract` was able to access the storage, we needed to modify the `PrecompiledContract` interface allowing the `Run` to take in input `evm *EVM`.

The definitions of the precompiled contracts can be found as usual in `autonity/core/vm/contracts.go`. Instead, the implementations are in `autonity/core/vm/USC.go`.