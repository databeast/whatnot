![Go](https://github.com/databeast/whatnot/workflows/Go/badge.svg)

# whatnot

If you like Etcd's namespace-driven notification subscription model, but have no requirements for long-term storage, then the formality of Etcetera can be dropped in favor of...whatnot.

Whatnot is An Etcd-like Lockable Namespace Manager, allowing you to define an hierarchical namespace of keys (which may contain values), that can be locked with an expiring lease, either individually or for an entire prefix of the key path tree. 

Performance is key, everything is done in-memory with no backing store or persistence. Multiple instances of your application can synchronize the data, so that persistence is perpetual so long as a single instance
remains online.

Leases and Watch notifications occur in much the same way that they do with Etcd and its GO client. Set a watch on a particular element of a namespaced path and recieve notifications of events to that element, or optionally, any element beneath it. Set a lease on a path element and receieve an expiring mutex on modifications to it, and optionally any of its sub-paths

## Optional Build Tags

metrics

errortraces

