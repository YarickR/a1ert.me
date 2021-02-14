**PMTS2XMPP**

#### Purpose: 
to retrieve alerts sent by Alertmanager and stored in Redis, filter them
according to set of rules, and deliver them to various groups to the
XMPP server(s) of choice

#### Requirements (Production): 
Linux, ability to run Go binaries

#### Requirements (Development): 
access to public Go repositories, ability to build Go binaries

### Configuration file format:
JSON containing following fields:

    {
	    "xmppServer": "<uri of XMPP server to deliver alerts to>",
	    "jidLogin": "<user on that server>",
	    "jidNick": "<Visible name of the user>",
	    "jidPwd": "<password for that user>",
	    "redisURI": "redis://<hostname>[:<port>]",
	    "groupsURI": "<XMPP groups URI>"
    }

### 

### Redis data structures:
All Redis-related structures reside in the database \#0 . All keys are
retrieved during startup. Additionally, **settings** key is retrieved
after each new packet is received, but before processing the packet, to
apply new channel definitions.

<table>
<tbody>
<tr class="header">
<td><strong>settings</strong></td>
<td>HASH</td>
<td><p><strong>channel_version</strong> uint</p>
<p>Any update of any channel should bump this number for the daemon to reload all channel definitions</p>
<p><strong>last_channel_id</strong> uint</p>
<p>ID of the last channel, to be able to preallocate required amount of channel structures</p>
<p><strong>shutdown</strong> bool</p>
<p>Setting to ‘1’ or ‘true’ will shut down the daemon</p></td>
</tr>
<tr class="odd">
<td><strong>channel_defs</strong></td>
<td>LIST</td>
<td>List of channels and their parameters (in JSON format) to filter and send messages to</td>
</tr>
<tr class="even">
<td><strong>alerts</strong></td>
<td>LIST</td>
<td>Message queue for incoming alerts. Each alert is in the JSON format, see below</td>
</tr>
<tr class="odd">
<td><strong>undelivered</strong></td>
<td>LIST</td>
<td>List of alerts that matched one or more rules in one or more channels, but no channels with matched rules were assigned to XMPP groups</td>
</tr>
<tr class="even">
<td><strong>unmatched</strong></td>
<td>LIST</td>
<td>List of alerts that didn’t match any rule</td>
</tr>
</tbody>
</table>

#### Channel:

    {
	    id		uint32, # channel id , used in rules for other channels
	    version	uint32, # channel definition version, should be bumped on each change for daemon to reload it’s properties
	    label	uint32, # channel description, for human consumption only, may be empty
	    group	string, # group to deliver alerts matched by any of the channel rules. May be empty
	    format	string, # message format to use , should contain Go templates . If empty, format for channel \#0 is used, if that is empty, default format ‘{{.}}’ is used
	    rules	[ rule, ... ]
    }

#### Rule:

    {
	    id 		uint32, # optional channel-specific rule id to reference this rule from rules in other channels. See condfrom
	    src 	uint32 | [ uint32, ... ],# channel id(s) used as alert sources to match against condition
	    cond 	string, # condition(s) to match alerts coming from source channels. Rules need either this field or **condfrom**
	    condfrom 	string # “channel_id:rule_id” think of it as a symlink to other rule's condition , to avoid duplication of similar rules in different channels
    }

Alert is a more or less generic JSON, matching language could operate on
field in the structure

### 

### Alert matching rule language:

It’s a simple Polish notation based set of functions, operators, and
field accessors . Typical rule looks like:

    regex "critical|warning" .labels.severity
    hasany ["gmru" "lootdog" "mailer" "mrac"] .labels.projects

etc. Currently supported functions, with required number of
arguments:

**true**: 0 returns true
**false**: 0 returns false
**eq**: 2 converts arguments to numbers, returns true if they are equal
**ne**: 2 same as **eq** , returns true if different
**lt**: 2 same as **eq**, returns if first argument is numerically less than second
**le**: 2 same as **lt**, but less or equal
**gt**: 2 same as **lt**, strictly greater
**ge**: 2 same as **gt**, greater or equal
**and**: 2 true if both arguments evaluate to true
**or**: 2 true if either argument is true
**not**: 1 evaluates arg as bool and returns negated value
**regex**: 2 both arguments are strings, first is a pattern, second is a field or result of a function to match pattern against . Returns true if pattern matches
**has**: 2 first argument is a string, second is a field or function result interpreted as an array. Returns true if any element of the array is equal (as string) to the first argument.
**hasany**: 2 both arguments are interpreted as arrays, returns true if any element of the first argument is present in the second array
**hasall**: 2 same as **hasany**, but returns true if all elements are present. Order is irrelevant
**since**: 1 returns number of seconds between current time and argument interpreted as date
**join**: 2 first argument is a string, second is an array (field or function result) with elements interpreted as strings. Returns single string - all elements of an array merged with first argument as a delimiter

Fields are specified using Golang template notation, using dots as context pointers , with leading dot meaning the whole alert is a context. For example, if the following message is picked up from **alerts** list in Redis:

    {
	    "received": 1613262923,
	    "message": {
	    "receiver": "itt_email"
	    "status": "firing",
	    "alerts": [
		    {
		    "status": "firing",
		    "labels": {
			    "alertname": "low_disk_space",
			    "device": "md18",
			    "fstype": "ext4",
			    "instance": "SRV106593.ko.ifc",
			    "job": "telegraf_ifc",
			    "mode": "rw",
			    "path": "/opt",
			    "severity": "minor",
			    "hosts": [
				    "SRV106593.ko.ifc"
			    ],
			    "projects": [
				    "hustle"
			    ],
			    "services": [
				    "db-5.hustle.srv."
			    ]
		    },
		    "annotations": {
			    "description": "/opt free space is 23.46%",
			    "summary": "/opt < 25%"
		    },
		    "startsAt": "2021-02-14T00:35:13.429828962Z",
		    "endsAt": "0001-01-01T00:00:00Z",
		    "fingerprint": "f3a395ecc7adaa40"
		    }
	    ],
	    "groupLabels": {
		    "alertname": "low_disk_space",
		    "instance": "SRV106593.ko.ifc"
	    },
	    "commonLabels": {
		    "alertname": "low_disk_space",
		    "device": "md18",
		    "fstype": "ext4",
		    "instance": "SRV106593.ko.ifc",
		    "job": "telegraf_ifc",
		    "mode": "rw",
		    "path": "/opt",
		    "severity": "minor"
	    },
	    "commonAnnotations": {
		    "description": "/opt free space is 23.46%",
		    "summary": "/opt < 25%"
	    },
	    "externalURL": "http://pmts-1.infra.srv:9093",
	    "version": "4",
	    "groupKey": "{}:{alertname=\"low_disk_space\", instance=\"SRV106593.ko.ifc\"}",
	    "truncatedAlerts": 0
	    }
    }

, then rules defined in channel \#0 will be applied for every item in the **alerts** list, and, if any of them will return true, then this item will run against rules in every channel using channel \#0 as it’s source. Rule to compare value of `labels.severity` field will look like
`regex “\^critical\$” .labels.severity` 
**eq** works with numbers, so string comparison here is done using **regex**. 
**has**/**hasall**/**hasany** compare strings using Golang “==” operator


