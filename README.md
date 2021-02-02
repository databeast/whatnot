

# Whatnot - an Etcd-like Distributed Locking Namespace Manager
![Go](https://github.com/databeast/whatnot/workflows/Go/badge.svg)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/a57e480a071f4017a3692adaf7f1da17)](https://app.codacy.com/gh/databeast/whatnot?utm_source=github.com&utm_medium=referral&utm_content=databeast/whatnot&utm_campaign=Badge_Grade)


If you like Etcd's namespace-driven notification subscription model, but have no requirements for long-term storage, then the formality of Etcetera can be dropped in favor of...whatnot.

Whatnot is An Etcd-like Lockable Namespace Manager, allowing you to define an hierarchical namespace of keys (which may contain values), that can be locked with an expiring lease, either individually or for an entire prefix of the key path tree. 

Performance is key, everything is done in-memory with no backing store or persistence. Multiple instances of your application can synchronize the data, so that persistence is perpetual so long as a single instance
remains online.

Leases and Watch notifications occur in much the same way that they do with Etcd and its GO client. Set a watch on a particular element of a namespaced path and recieve notifications of events to that element, or optionally, any element beneath it. Set a lease on a path element and receieve an expiring mutex on modifications to it, and optionally any of its sub-paths

## Key Functionality



## Why use this over <ProjectName>

Emphasis on lightweight
* No additional infrastructure
* No storage persistence/synchronization latency
* No huge memory requirements
* not intended to be used for billions of key

Emphasis on Structure and Hierarchy
* Etcd's explicity directory-path style namespaces were a huge inspiration


Decentralized
* 

Built without storage persistence and centralization in mind.

Designed originally to coordinate cluster cache invalidation

## Optional Build Tags

metrics

errortraces

