import os
import json
import requests

requests.packages.urllib3.disable_warnings()


class Client(object):

    def __init__(self, addr, token, **kws):
        self.addr = addr
        self.token = token
        self.debug = kws.get('debug', False)
        self.default()

    def default(self):
        if self.addr.find(':') < 0:
            self.addr = '{}:{}'.format(self.addr, 10000) 
  
    def request(self, url, method, data=""):
        url = "https://{}/api/{}".format(self.addr, url)
        if self.debug:
            print("{} {} {}".format(method, url, data))

        resp = requests.request(method, url,
                                json=data, verify=False,
                                auth=(self.token, ''))
        if self.debug:
            print("RESPONSE {}".format(resp.text))

        if not resp.ok:
            resp.raise_for_status()

        return resp


if __name__ == '__main__':
    addr = os.environ.get("OL_ADDRESS", "localhost")
    token = os.environ.get("OL_TOKEN", "")

    c = Client(addr, token)
    resp = c.request("user", "GET")
    print(json.dumps(resp.json(), indent=4))

    resp = c.request("link", "GET")
    print(json.dumps(resp.json(), indent=4))

    resp = c.request("neighbor", "GET")
    print(json.dumps(resp.json(), indent=4))
