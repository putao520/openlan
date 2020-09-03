# controller
Manage Multiple vSwitch and Uniform Configuration by CLI or UI.

# feature 

1. Point can discover vSwitch by controller, and controller gives a vSwitch base point's location；
2. If Point access to vSwitch is missed, and vSwitch NEED forward auth to Controller;
3. When add new vSwitch on Controller, Controller MUST call vSwitch's API to set it's master to Controller;
4. If vSwitch has master, it NEED send accessed point and status to master, and include changing or modifying;
5. If vSwitch find network is missed, and vSwitch NEED forward request to Controller;
6. When vSwitch connected to Controller, MUST send point, link and other information at lease once;
7. When this information changed, MUST send a update tp vSwitch;
8. A heartbeat MUST be keep alive between Controller and vSwitch to detect vSwitch whether closed or disconnect;

