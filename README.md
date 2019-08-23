# Openlan-go

The OpenLAN project providers one solution for you to access your any site from any where. 

# Version1 (active)

Refer to https://github.com/danieldin95/openlan-go/tree/master/olv1.

                                          FireWall --- CPE D --- Host3
                                              |
                            -----------------OPE--------------------
                            |                 |                    |
                         Firewall          Firewall             FireWall
                            |                 |                    |
                Host1 --- CPE A             CPE B                CPE C --- Host2

# Version2 (abandon)

Refer to https://github.com/danieldin95/openlan-go/tree/master/olv2.

                                            Controller
                                                |
                            ------------------------------------------
                            |                                        |
                            >------------------UDP-------------------<
                            |                                        |
                            >---UDP---<                   >----UDP---<    
                            |         |<------->|<------->|          |
                         Firewall            Firewall             FireWall
                            |                   |                    |
                Host1 --- Endpoint A         Endpoint B            Endpoint C --- Host2
