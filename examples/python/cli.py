# -*- coding: utf-8 -*-


import requests
import hashlib
import time
import json
import getopt, sys

class Client(object):
    def __init__(self, domain, secret, ssl):
        self.domain = domain
        self.secret = secret
        self.host = 'http://' + domain
        if ssl:
            self.host = 'https://' + domain

    @staticmethod
    def hash(query):
        m = hashlib.new('md5')
        keys = list(query.keys())
        keys.sort()
        for key in keys:
            m.update(query[key].encode(encoding='UTF-8'))
        m.update(secret.encode(encoding='UTF-8'))
        return m.hexdigest()

    def query_dns(self, q, blur=False):
        payload = {}
        payload['t'] = str(int(time.time()))
        print('t=', int(time.time()), payload['t'])
        payload['q'] = q
        if blur:
            payload['blur'] = '1'
        else:
            payload['blur'] = '0'
        payload['hash'] = Client.hash(payload)

        r = requests.get(self.host + '/data/dns', payload)
        print('resp:', r.text)
        j = json.loads(r.text)
        if r.status_code == 200:
            return j['result'], j['message']
        else:
            return None, j['message']

    def query_http(self, q, blur):
        payload = {}
        payload['t'] = ''.format('{}', int(time.time()))
        payload['q'] = q
        if blur:
            payload['blur'] = '1'
        else:
            payload['blur'] = '0'
        payload['hash'] = Client.hash(payload)

        r = requests.get(self.host + '/data/http', payload)
        print('resp:', r.text)
        j = json.loads(r.text)
        if r.status_code == 200:
            return j['result'], j['message']
        else:
            return None, j['message']


def usage():
    print('cli: query dns/http log result')
    sys.exit()

if __name__ == '__main__':
    try:
        options, args = getopt.getopt(sys.argv[1:], "hs:d:t:q:l:", [ "help", "blur", "ssl", "domain=", "secret=", "query=", "type=dns" ])
    except getopt.GetoptError as err:
        # print help information and exit:
        print(err) # will print something like "option -a not recognized"
        usage()
        sys.exit(2)

    secret = ''
    domain = ''
    ssl = False
    type = 'dns'
    blur = False
    query = ''
    for option, value in options:
        if option in ("-h", "--help"):
            usage()
        if option in ("-s", "--secret"):
            secret = str(value)
        if option in ("-d", "--domain"):
            domain = str(value)
        if option in ("-q", "--query"):
            query = str(value)
        if option in ("-t", "--type"):
            cmd = str(value)
        if option in ("-b", "--bulr"):
            bulr = True
        if option in ("-l", "--ssl"):
            ssl = True
    if secret == '' or domain == '' or query == '':
        print("secret,domain,query are required")
        sys.exit()

    c = Client(domain, secret, ssl)
    if type == 'dns':
        r, m = c.query_dns(query, blur)
        print(m, r)
    elif type == 'http':
        r,m = c.query_http(query, blur)
        print(m, r)
    else:
        print('unknown type', type, 'only support dns/http')

