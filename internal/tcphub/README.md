Tests:



Radar to Client
RadarToClient_test

Host Hub Server
Host Hub Client

Pump "RadarIP" data to the server
Client get the data and:
1. Sends it from the "RadarIP" address to the ClientIPAddr
2. Host a python script as the client to read all info and report it

Issues:
Only start sending information when the client connects in order
to reconcile sent and received totals 




1. Start up the host @ 127.0.0.1:45000
2. Start up the virtual host
