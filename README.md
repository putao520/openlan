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

<b>UDP Hole</b>: Endpoint sends periodically UDP packet to keep a hole on firewall. And The hole can receive UDP data from any source such as other Endpoint.

<b>Endpoint</b>: Represent a branch site in one brocast domain, and others Endpoint discover peer Endpoint by Controller, a hole <IPAddress, UDPPort> as unique key.

<b>Host</b>: Under Endpoint or Endpoint self. Controller records all hosts under all Endpoint, and annouces them to Endpoint periodically, and using <IPAddress, HardwareMac> as unique key.

# Endpoint Online

Endpoint MUST send UDP keepalive packet preiodically(default is 5s), to keep a hole on firewall as represent on WAN. 

# Discover Endpoint

When Endpoint received a ARP request from host. Fistly, lookup destination whether on remote Endpoint by local ARP cache. And then learn it, send a host learning packet to Controller. If matched, encapsulation it to remote Endpoint by UDP or drop it.

# Host Learn and Announce

When Controller received a host learning packet from Endpoint, save in local host table and announce it to peer Endpoint.

# Packet Format

We use UDP to carry control and ethernet message between Endpoint, Endpoint and Controller. The UDP port on Controller uses 1002x series, default is 10020, and Endpoint uses 1001x series, default is 10010.

    0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8
    +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
    |            MAGIC                |              Length                |       
    +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
    |                               Payload                                |
    +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
    
    MAGIC: 0xFFFF
    Payload:
        If Dest.MAC isn't ZERO:
            Padded by Ethernet Frame.
        Else:
            0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8 0 1 2 3 4 5 6 7 8
            +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
            |                               0x00                                   |
            +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
            |                0x00             |        Type     |      Resv        |
            +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
            |                               Data                                   |
            +-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+-+-+-+-++-+-+-+-+-+-+-+
            
            Type: 
                Hello          0x01
                Host Learning  0x02
                Host Announce  0x03
                Authentication 0x04
                Acknowledge    0x05
            
