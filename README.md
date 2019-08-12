# Openlan-go

The Golang implements for OpenLAN project.
    
                                               Controller
                                                   |
                            ------------------------------------------
                            |                      |                  |
                         UDP Hole               UDP Hole           UDP Hole
                            |                      |                  |
                         Firewall               Firewall           FireWall
                            |                      |                  |
                Host1 --- Endpoint A           Endpoint B         Endpoint C --- Host2

<b>Controller</b>: Which is running on VPS or server has WAN's IP address.
<b>UDP Hole</b>: Endpoint sends periodically UDP frame to keep a hole on firewall. And The hole can receive UDP data from any source such as other Endpoint.
<b>Endpoint</b>: Represend a branch site in one brocast domain. And others Endpoint discover peer Endpoint by Controller, a hole <IPAddress, UDPPort> as unique key.
<b>Host</b>: A host under Endpoint or Endpoint self. Controller records all hosts for per Endpoint, and annouces them to Endpoint periodically. A Host using <IPAddress, HardwareMac> as unique key.

# Endpoint Online

TODO

# Discover Endpoint

TODO

# Host Learn and Announce

TODO
