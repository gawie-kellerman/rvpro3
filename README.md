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
Message must have a commom identification header.

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

