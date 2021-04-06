#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>          /* See NOTES */
#include <sys/socket.h>
#include <linux/udp.h>
 
#include <netinet/in.h>
#include <netinet/ip.h>
 
#include <linux/xfrm.h>
#include <linux/ipsec.h>
#include <linux/pfkeyv2.h>
 
 
int main(int argc, char *argv[])
{
        int on = 1;
        struct xfrm_userpolicy_info policy;
        int type = UDP_ENCAP_ESPINUDP;
 
        struct sockaddr_in addr = {
                .sin_family = AF_INET,
                .sin_port = htons(4500),
                .sin_addr = {
                        .s_addr = INADDR_ANY,
                },
        };
 
 
        int fd = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
        if (fd == -1)
                return -1;
 
        if (setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, (void*)&on, sizeof(on)) < 0) {
                printf("unable to set SO_REUSEADDR on socket: %s", strerror(errno));
                return -1;
        }
 
        /* bind the socket */
        if (bind(fd, (struct sockaddr*)&addr, sizeof(addr)) == -1) {
                printf("unable to bind socket: %s", strerror(errno));
                return -1;
        }
 
        memset(&policy, 0, sizeof(policy));
        policy.action = XFRM_POLICY_ALLOW;
        policy.sel.family = AF_INET;
 
        policy.dir = XFRM_POLICY_OUT;
        if (setsockopt(fd, IPPROTO_IP, IP_XFRM_POLICY, &policy, sizeof(policy)) < 0) {
                printf("unable to set XFRM_POLICY on socket: %s\n",
                                strerror(errno));
                return -1;
        }
        policy.dir = XFRM_POLICY_IN;
        if (setsockopt(fd, IPPROTO_IP, IP_XFRM_POLICY, &policy, sizeof(policy)) < 0) {
                printf("unable to set XFRM_POLICY2 on socket: %s\n",
                                strerror(errno));
                return -1;
        }
 
 
        if (setsockopt(fd, IPPROTO_UDP, UDP_ENCAP, &type, sizeof(type)) < 0) {
                printf("unable to set UDP_ENCAP: %s\n", strerror(errno));
                return -1;
        }
 
        while (1) {
                pause();
        }
 
        return 0;
}
