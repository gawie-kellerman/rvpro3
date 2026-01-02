## UDP Flow

1. Receive UDP
2. Preprocess UDP

## Serial Flow

1. Receive Serial
2. Parse Serial

Receive -> Parse -> Process
Receive -> Parse -> Process

UDP Mode
UDPKeepAliveService 
-> UDPDataService
-> UDPDataProcessor -> UDPDataMapper
  -> Send data to TCP client
-> MessageHub
-> MessageHandler
-> PVRMessageHandler
-> StatisticsHandler
-> ObjectListHandler
-> TriggerHandler
-> DiagnosticsHandler

Serial UMRR Mode
-> SerialDataService
-> SerialDataMapper
  -> SerialDataParser
-> MessageHub...

Serial Bits Mode
-> SerialBitsDataService
-> SerialBitsDataMapper
-> MessageHub
-> ...


MessageHub sends to channel array
MessageHandler listens to channel
Message must have a common identification header.

UDPDataProcessor uses UDPDataMapper
... latch onto UDPDataService events
... meaning it is not tightly coupled
... UDPDataProcessor must have event sink to allow TCP Client
    TCP Client interested in data buffer and Port header
... Must be able to make call to throw away message that will not be used in downline



BIG IDEA:
Instruction Executor

Methods:
ExecuteIncoming()
ExecuteIdle()
ExecuteStartup()

Must be able to handle instruction responses
Must be able to handle the lack of instruction responses

The gist:
It gets loaded with 1 or more instructions to execute against
a radar.  
The instruction requests gets retried if not received baded on a
retry count


InstructionService
Properties
1. queue
2. 

Methods:
Enqueue(inFront bool)
OnHandleResponse(*self, queueItem, instruction) shouldPop?
OnHandleIdle

InstructionQueueItem
1. MaxRetries
2. RetryCount
3. RetryOn
4. RadarIP
5. RadarPort
6. SequenceNo
7. Instruction




## Metrics
```bash
# Get list of available metric sections
curl -s "localhost:8080/metrics/sections" | jq

# Get detail of metric section(s)
curl -s "localhost:8080/metrics/section?sn=UDP.Metric&sn=Radar.192.168.11.12:55555" | jq

# Stop Radars
curl -X PUT "localhost:8080/executor/radars/stop"

# Start Radars
curl -X PUT "localhost:8080/executor/radars/start"

# Radars Status
curl -s "localhost:8080/executor/radars/status" | jq
```

BIG TODOs:

1. For remote/vs/local, switch Keep Alive and UDP Data off
2. For fail safe, switch SDLC writing off
3. Get version number in through the build




## Config/GlobalConfig
The Config is a simple key value pair of strings separated by an
equals (=) sign.  Every key is unique.  Duplicating a key will mean
that the last entry is the value for the key.
In a config file, blank lines are ignored.  So too are lines
that start with #, which is seen as a comment.  

There are 3 categories of config.  

### Key Categories
#### Global Key
Global keys start with ```Global.```.   It is single value entries.
Notes:
* Object(s) using these keys expect these values to be configured.
* The objects generally are singleton by design
* Having the settings are mandatory.  That does not mean that the objects are required.

Example:
```text
Global.UDP.KeepAlive.CastIP = 239.144.0.0:60000
Global.UDP.KeepAlive.RepeatMillis = 60000
Global.Http.Host = localhost:8080
```

#### Default Key
Default keys serve as default settings where specific indexed
key overrides are not supplied.

Example:
```text
# To specify that object list should generally be switched off
Default.ObjectListPath = 
```

#### Indexed Key
Indexed Keys provide the ability to override **Default Key(s)**.
Indexed Kes and Default Keys are partners, while Global Keys are loose standing.

When provided a default key 
```text
Default.ObjectListPath = 
```

A possible override is 
```text
Radar.192.168.11.12.ObjectListPath = /media/path/whatever
````

When an 'indexed' object needs its settings it asks the Config
for an indexed key using its unique entity name, entity index and config key where, 
as per the example:

* entity name is Radar
* entity index is 192.168.11.12
* config key is ObjectListPath

Which will first search for a specific key ```Radar.192.168.11.12.ObjectListPath```.
Failure to find this key will result in a search where:

* ```Default``` is used as the entity key
* config key is ObjectListPath

Resulting in a key ```Default.ObjectListPath```

Failure to also find a default value will result in a panic.

### VERY IMPORTANT

#### Key Conflict
If you consider the key ```Default.ObjectListPath``` then it is important to note that, given
the design of Default and Indexed keys that ```ObjectListPath``` must be unique as you cannot have two
different indexed entity (object) types with the key name ```ObjectListPath```.  You have to make
the key part unique across different object types e.g.

```text
Default.ObjectTypeA.ObjectListPath
vs
Default.ObjectTypeB.ObjectListPath
```

In other words: Even though your specific key is unique by virtue of the index, you may have
clashes at the Default.  So beware.

#### Key Strategy
The strategy is, that we anticipate to never know what will be included/excluded from running and 
instead of creating a horde of different configurations, we rather create sane defaults for every
possible item (as Globals or Defaults).

This not only allows none or partial configuration and successful execution, but further
allows for overriding selective values even after successful configuration due to a 
multistep configuration of:

1. Initialize defaults and globals on program started
2. Override settings from loaded configuration
3. Override settings from command line arguments
4. Override settings from with the running application e.g. web service

It further leaves the option to combine/load multiple configuration files to tweak behavior, which
will assist in simplifying configuration of specific services/objects within the application.
