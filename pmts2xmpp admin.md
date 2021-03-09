### PMTS2XMPP administrative interface
#### Purpose:
Operational management of PMTS2XMPP alert processor and its corresponding data structures in Redis database

#### Main topics:
 - Channel management
 - Create channel
 - Delete channel
 - Add rule to the channel ruleset
 - Delete rule from the ruleset
 - Modify template attached to the channel
 - Display channel hierarchy (graph or other visual means to see dependencies and possible circular dependencies)

#### Examining message queues in Redis instance
There are several message queues implemented as lists in Redis - incoming alerts (**alerts** list), alerts that weren’t matched by any of the channel rules (**unmatched** list), and alerts that were matched by channels without defined XMPP groups (**undeliverable** list). Administrative interface should allow viewing the number of items in each queue,  examining them, and moving messages from unmatched and undeliverable queues back to incoming.

