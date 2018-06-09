#! /usr/bin/env python
# -*- coding: utf-8 -*-
import json
import os, sys
import time, datetime

config_path = sys.argv[1]
with open(config_path, 'rt') as f:
    config = json.load(f)

redis_host = config["redis_host"]
redis_port = config["redis_port"]
backup_dir = config["backup_dir"]
project = config["project"]
gopath = os.environ.get("GOPATH") or config["gopath"]
now = str(datetime.datetime.now())
dst_path = os.path.join(backup_dir, now)
src_path = os.path.join(gopath, "src", project, "iserver")
block_db = "blockDB"
tx_db = "txDB"
res = "SUCCESS"

def backup_db(src):
    filename = os.path.basename(src) + ".tar.gz"
    dst = os.path.join(dst_path, filename)
    command = "cd {2} && tar -zcf '{0}' '{1}'".format(dst, src, src_path)
    print("Runing: " + command)
    if os.system(command) != 0:
        res = "FAIL"

def backup_redis():
    filename = "dump.rdb"
    dst = os.path.join(dst_path, filename)
    command = "redis-cli -h {0} -p {1} --rdb '{2}'".format(redis_host, redis_port, dst)
    print("Runing: " + command)
    if os.system(command) != 0:
        res = "FAIL"

if __name__ == "__main__":
    os.makedirs(dst_path)
    backup_db(block_db)
    backup_db(tx_db)
    backup_redis()

print(res)

