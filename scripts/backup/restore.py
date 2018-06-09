#! /usr/bin/env python
# -*- coding: utf-8 -*-
import json
import os, sys
import time, datetime

config_path = sys.argv[1]
time_dir = sys.argv[2]
with open(config_path, 'rt') as f:
    config = json.load(f)

redis_host = config["redis_host"]
redis_port = config["redis_port"]
backup_dir = config["backup_dir"]
project = config["project"]
gopath = os.environ.get("GOPATH") or config["gopath"]
dst_path = os.path.join(backup_dir, time_dir)
src_path = os.path.join(gopath, "src", project, "iserver")
block_db = "blockDB"
tx_db = "txDB"
res = "SUCCESS"

def restore_db(src):
    filename = os.path.basename(src) + ".tar.gz"
    dst = os.path.join(dst_path, filename)
    command = "cd {1} && tar -zxf '{0}'".format(dst, src_path)
    print("Runing: " + command)
    if os.system(command) != 0:
        res = "FAIL"

def restore_redis():
    filename = "dump.rdb"
    dst = os.path.join(dst_path, filename)
    command = "rdb --command protocol '{2}' | redis-cli -h {0} -p {1} --pipe".format(redis_host, redis_port, dst)
    print("Runing: " + command)
    if os.system(command) != 0:
        res = "FAIL"

if __name__ == "__main__":
    restore_db(block_db)
    restore_db(tx_db)
    restore_redis()

print(res)

