# Whatnot - an Etcd-like Distributed Locking Namespace Manager

![Go](https://github.com/databeast/whatnot/workflows/Go/badge.svg)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/a57e480a071f4017a3692adaf7f1da17)](https://app.codacy.com/gh/databeast/whatnot?utm_source=github.com&utm_medium=referral&utm_content=databeast/whatnot&utm_campaign=Badge_Grade)
[![Go Report Card](https://goreportcard.com/badge/github.com/databeast/whatnot)](https://goreportcard.com/report/github.com/databeast/whatnot)

If you like Etcd's namespace-driven notification subscription model, but have no requirements for long-term storage,
then the formality of Etcetera can be dropped in favor of... Whatnot.

Whatnot is An Etcd-like Lockable Namespace Manager, allowing you to define a hierarchical namespace of keys (which may
contain values), that can be locked with an expiring lease, either individually or for an entire prefix of the key path
tree.

Performance is key, everything is done in-memory with no backing store or persistence. Multiple instances of your
application can synchronize the data, so that persistence is perpetual so long as a single instance remains online.

Leases and Watch notifications occur in much the same way that they do with Etcd and its Go client. Set a watch on a
particular element of a namespaced path and recieve notifications of events to that element, or optionally, any element
beneath it. Set a lease on a path element and receieve an expiring mutex on modifications to it, and optionally any of
its sub-paths.

## Suggested Use Cases  

Originally inspired by the Kubernetes REST API and its hierarchical ordering of resources, I wanted something similar, but with
lower latency, and without storing the full resources into the system itself. I do a lot of microservices development and a key 
design factor in those is being able to extend out the datamodel, without having to backport it to existing, stable components every time.

Etcd's capability to arrange keys in a directory-like namespace, and apply mutex-style leases to a both a single key and its descendents
appealed to me for this reason - allowing other components to extend a resource and be notified of it without other components having to 
know of these items specifically. Etcd's ability to subscribe ('watch') a key, and send notifications to subscribers of leases and changes 
to a single key or all of its children as well, was immensely useful for this design pattern.

Caching and Resource Locking are two tasks that are almost inseparable, and this is where Whatnot is intended to excel - acting as 
a distributed state coordinator, especially for microservice-based application architectures, where multiple elastically-scaling
instances of a service exist together in coordinated deployments, and the overhead of an external service (redis, etc) to provide
this functionality is an unnecessary resource overhead when all is required from these services is the coordination of state.

### Why use this over [ProjectName]

#### Over Etcd ?

Simply put, Etcd's storage quorum functionality placed an upper limit on the number of notifications per second, and their 
accumulated latency, causing some notifications to be placed into a 'get around to eventually' queue. Those limitations
combined with the resource requirements of Etcd were what inspired this project in the first place.

#### Over Redis ?

Redis gives tools to build some complex setups that resemble this, and it has close to the required speed too, but they require use of advanced commands and data modelling
to approximate the same functionality. All of this places Redis on a scale larger than what I required here.  
  
#### Over Consul ?

Consul, like Etcd, has the same hierarchical keyspace support, but focuses on persisting data with fairly low levels of state changes. 

### Discouraged Use cases

Whatnot was created with the intention of creating a specialized package that incorporates certain features of larger, more
popular software projects, without the unwanted overhead of their additional featureset - hence.

#### Whatnot is not intended for persisting data.

Although its possible to write your own serializer to save the state of a Whatnot namespace (perhaps during application shutdown)
the internal system itself is not intended to incorporate any kind of storage backend - Other software accomplishes this requirement
and use pattern far better already. So although Whatnot includes support for values on keys, those values should never be the only
canonical store of that value.

* Speed is always in contention to Accuracy
* Extremely large numbers of keys

## Key Functionality

#### Individual Namespaces

* Completely separate notification channels

#### Native Hierarchy Support

Just as with Etcd, Keys are organized into directory-like tree structures, where every key can have sub-keys.

Just as with directory paths, this gives us the concept of both Absolute Paths and Relative Paths.

#### Leases and Prefix Leases

Returns a native Go `context.Context` object and `cancel()` function - once you have obtained a lock, the rest of your code doesnt 
need to care about Whatnot, just follow regular Go patterns for working within a scope that has a deadline. Leases will forcibly 
unlock the corresponding element once their deadline is reached, so it is up to your code to obey that channel signal and cease any 
further operations that could induce race conditions.

### Emphasis on lightweight operation

* No additional infrastructure
* No storage persistence/synchronization latency
* No huge memory requirements
* not intended to be used for billions of key

### Emphasis on Structure and Hierarchy

* Etcd's explicit directory-path style namespaces were a huge inspiration

### Decentralized

* Support for direct clustering between your applications instances via Gossip, Raft and gRPC

### Built without storage persistence and centralization in mind.

* Designed originally to coordinate cluster cache invalidation

## Resource Management

Whatnot defaults to a _highly_ reactive concurrency model, whereby each distinctly unique element in the namespace
recieves its own dedicated routine to pass up event notifications to its parent.

