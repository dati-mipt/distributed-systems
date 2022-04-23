# Distributed Systems course

### Course program
__Topic 1. Models of distributed systems__:
* Two generals and Byzantium generals problems
* System models of network, node crash and timing behavior
* Availability and Fault Tolerance

__Topic 2. Time, clocks and ordering__:
* Usage of physical time for ordering
* Usage of logical time ordering
* Happens-before and causality relation

__Topic 3. Logical Time and Broadcast protocols__:
* Lamport clocks
* Vector clocks
* Broadcast ordering models and implementations: FIFO, causal, total order, etc…
* Gossip protocols

__Topic 4. Replication__:
* Retrying state updated
* Idempotence
* Reconciliation
* Concurrent writes
* State machine replication

__Topic 5. Quorums__:
* Quorums fault tolerance scheme
* Linearizable fault tolerant register implementation

__Topic 6. Strong Eventual Consistency and Conflict-free Replicated Data Types__:
* Convergent Replicated Data Types
* Commutative Replicated Data Types
* Complex CRDTs: texts, graphs, etc…
* Collaboration Software

__Topic 7. Eventual Consistency (several lectures)__:
* CAP theorem
* Formal model of histories, visibility and arbitration order on events
* Session guarantees
* Eventual, PRAM, Causal, Sequential and Strong Consistency and their relations
* Return-Value consistency
* Key concepts of implementations of counters, registers and key-value stores with different consistency guarantees with examples
* Algebraic proofs that such implementations satisfy consistency guarantees

__Topic 8. Consensus (several lectures)__:
* FLP theorem
* Problem Statement
* RAFT
* PAXOS
* Improvements

__Topic 9. Formal verification of distributed systems algorithms via TLA+__

__Topic 10. Modern blockchain systems__:
* Consensus algorithms and its verifications:
    * Proof-of-work
    * Proof-of-stake
* Smart contracts


### Home assignments
* [Home Assignment 1](https://github.com/dati-mipt/distributed-systems/blob/master/assignments/hw1/README.md)
