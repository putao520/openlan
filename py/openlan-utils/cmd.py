import json
from .client import Client


def with_client(func):
    def decorate(opt, *args, **kws):
        c = Client(opt.ol_server, opt.ol_token)
        return func(c, opt, *args, **kws)

    return decorate


@with_client
def list_user(client, opt):
    resp = client.request("user", "GET")
    print json.dumps(resp.json(), indent=4)


@with_client
def add_user(client, opt):
    if opt.username is None or opt.password is None:
        return

    data = {
        'name': opt.username,
        'password': opt.password
    }
    resp = client.request("user", "POST", data)
    print json.dumps(resp.json(), indent=4)


@with_client
def del_user(client, opt):
    if opt.username is None:
        return

    data = {
        'name': opt.username
    }
    resp = client.request("user", "DELETE", data)

    print resp.text
