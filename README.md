# Openlan-go

The Golang implements for OpenLAN project.
The OpenLAN project providers one solution for you to access your any site from any where. 

# Version1 

Refer to https://github.com/danieldin95/openlan-go/tree/master/olv1.

                                                  OPE
                                                   |
                            ---------------------------------------------
                            |                      |                    |
                         Firewall               Firewall             FireWall
                            |                      |                    |
                Host1 --- CPE A                  CPE B                CPE C --- Host2

# Version2

Refer to https://github.com/danieldin95/openlan-go/tree/master/olv2.

                                               Controller
                                                   |
                            ---------------------------------------------
                            |                      |                    |
                            >---UDP---<                      >----UDP---<    
                            |         |<---------->|<------->|          |
                         Firewall               Firewall             FireWall
                            |                      |                    |
                Host1 --- Endpoint A           Endpoint B             Endpoint C --- Host2
                            |                                           |
                            >---------------------UDP-------------------<
