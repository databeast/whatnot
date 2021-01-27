/*
Whatnot is a low-footprint, high-speed, cluster-friendly volatile system for providing etcd-like
distributed semaphores on hierarchical resource identifiers with subscribable watch notifications.

It is intended to replace the functionality of systems like Redis and Etcd, in situations where
data persistence over long periods of time is not an issue, additional infrastructure is not desired,
and peer-to-peer sharing of data is a preferable solution for extreme low-latency

Whatnot (a far more informal rendering of 'etcetera') was driven by a desire to utilize the namespace
subscription capabilities of Etcd, without the investment in storage and memory it required to maintain
persistent data, a feature I did not require at the time.
*/
package whatnot
