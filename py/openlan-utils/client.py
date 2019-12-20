import os
import json
import requests

requests.packages.urllib3.disable_warnings()


class Client(object):

    def __init__(self, addr, token):
        self.addr = addr
        self.token = token

    def request(self, url, method, data=""):
        url = "https://{}/api/{}".format(self.addr, url)
        return requests.request(method, url,
                                json=data, verify=False,
                                auth=(self.token, ''))


if __name__ == '__main__':
    addr = os.environ.get("OL_ADDRESS", "localhost:10000")
    token = os.environ.get("OL_TOKEN", "")

    c = Client(addr, token)
    resp = c.request("user", "GET")
    print json.dumps(resp.json(), indent=4)

    resp = c.request("link", "GET")
    print json.dumps(resp.json(), indent=4)

    resp = c.request("neighbor", "GET")
    print json.dumps(resp.json(), indent=4)