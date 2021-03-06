import requests
import json
import logging
import time

logger = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)


def main(rpc_endpoint, redis_client, notification_channel, sleep):
    try:
        block_number = int(
            redis_client.get(notification_channel + "::blocknumber")
        )
    except Exception:
        block_number = int(requests.post(
            rpc_endpoint,
            json={"jsonrpc": "2.0", "method": "eth_blockNumber", "params": [],
                  "id":6}
        ).json()["result"], 16)

    while True:
        try:
            res = requests.post(
                rpc_endpoint,
                json={"jsonrpc": "2.0", "method": "eth_getBlockByNumber",
                      "params": [hex(block_number), False], "id": 64}
            )
            next_block = res.json()["result"]
        except KeyError:
            continue
        except json.decoder.JSONDecodeError:
            logger.exception("JSON Decode error for: '%s'", res.content)
            time.sleep(1)
            continue

        if next_block is None:
            time.sleep(5)
            continue
        new_block_number = int(next_block['number'], 16)
        if new_block_number != block_number:
            raise ValueError(
                "Received block does not match request block number"
            )
        if next_block['logsBloom'].strip("0") == "x":
            logger.info("Skipping block %s (empty)", block_number)
            block_number += 1
            time.sleep(0.1)
            continue
        message = json.dumps({
            "hash": next_block['hash'],
            "number": new_block_number,
            "bloom": next_block['logsBloom']
        })
        redis_client.lpush(notification_channel, message)
        logger.info("Publishing block %s: %s", block_number, message)
        block_number += 1
        redis_client.set(notification_channel + "::blocknumber", block_number)
        time.sleep(0.1)

if __name__ == "__main__":
    import argparse
    import redis
    parser = argparse.ArgumentParser()
    parser.add_argument("rpc_endpoint")
    parser.add_argument("redis_host")
    parser.add_argument("redis_queue")
    parser.add_argument("--sleep", type=int, default=2)
    args = parser.parse_args()
    if not args.redis_queue.startswith("queue://"):
        raise ValueError("redis queue must start with queue://")
    redis_host = args.redis_host.split(":")
    if len(redis_host) > 1:
        redis_client = redis.Redis(redis_host[0], redis_host[1])
    else:
        redis_client = redis.Redis(redis_host[0])
    main(args.rpc_endpoint, redis_client, args.redis_queue[len("queue://"):], args.sleep)
