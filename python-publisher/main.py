"""
This script is responsible for creating and sending messages
to Redis using either Pub/Sub, Queue (list), or both mechanisms.
"""

import redis
import json
import time
import uuid
import os
import signal
import sys
from datetime import datetime
from typing import Optional

class MessagePublisher:
    """
    This class encapsulates all the logic for generating messages
    and publishing them to Redis.
    """

    def __init__(self):
        # Configuration from environment variables
        self.redis_host = os.getenv('REDIS_HOST', 'localhost')
        self.redis_port = int(os.getenv('REDIS_PORT', '6379'))
        self.publish_interval = float(os.getenv('PUBLISH_INTERVAL', '2.0'))
        self.pattern = os.getenv('PATTERN', 'both')

        # Redis client and internal state
        self.redis_client: Optional[redis.Redis] = None
        self.running = True
        self.message_count = 0

        # Setup signal handlers for graceful shutdown
        signal.signal(signal.SIGTERM, self._signal_handler)
        signal.signal(signal.SIGINT, self._signal_handler)

    def _signal_handler(self, signum, frame):
        print(f"\n[{self._get_timestamp()}] Received signal {signum}. Shutting down gracefully...")
        self.running = False

    def _get_timestamp(self) -> str:
        return datetime.now().isoformat()

    def connect_to_redis(self) -> bool:
        """
        Attempts to connect to Redis with timeout and retries.
        """
        try:
            self.redis_client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                decode_responses=True,
                socket_connect_timeout=5,
                socket_timeout=5,
                retry_on_timeout=True
            )

            # Test the connection
            self.redis_client.ping()
            print(f"[{self._get_timestamp()}] Connected to Redis at {self.redis_host}:{self.redis_port}")
            return True

        except redis.RedisError as e:
            print(f"[{self._get_timestamp()}] Failed to connect to Redis: {e}")
            return False

    def create_message(self) -> dict:
        """
        Generates a unique message payload.
        """
        self.message_count += 1
        return {
            "id": str(uuid.uuid4()),
            "timestamp": self._get_timestamp(),
            "message": f"Hello from Python #{self.message_count}",
            "sender": "python-publisher",
            "count": self.message_count
        }


    def publish_message(self, message: dict, method: str = None) -> bool:
        """
        Publishes a message to Redis using the selected method.
        """
        try:
            if method is None:
                method = self.pattern

            message_json = json.dumps(message)

            if method in ['pubsub', 'both']:
                subscribers = self.redis_client.publish('messages', message_json)
                print(f"[{self._get_timestamp()}] Published to channel 'messages' ({subscribers} subscribers): {message['message']}")

            if method in ['queue', 'both']:
                list_length = self.redis_client.rpush('message_queue', message_json)
                print(f"[{self._get_timestamp()}] Pushed to queue 'message_queue' (length: {list_length}): {message['message']}")

            return True

        except redis.RedisError as e:
            print(f"[{self._get_timestamp()}] Failed to publish message: {e}")
            return False

    def run(self):
        """
        Main loop that creates and publishes messages at regular intervals.
        """
        print(f"[{self._get_timestamp()}] Starting Python Publisher...")
        print(f"[{self._get_timestamp()}] Pattern: {self.pattern}")
        print(f"[{self._get_timestamp()}] Publish interval: {self.publish_interval}s")

        # Retry Redis connection
        max_retries = 5
        for attempt in range(1, max_retries + 1):
            if self.connect_to_redis():
                break
            print(f"[{self._get_timestamp()}] Connection attempt {attempt}/{max_retries} failed. Retrying in 5s...")
            time.sleep(5)
        else:
            print(f"[{self._get_timestamp()}] Failed to connect to Redis after {max_retries} attempts. Exiting.")
            sys.exit(1)

        # Message publishing loop
        while self.running:
            try:
                message = self.create_message()
                self.publish_message(message)
                time.sleep(self.publish_interval)
            except KeyboardInterrupt:
                break
            except Exception as e:
                print(f"[{self._get_timestamp()}] Unexpected error: {e}")
                time.sleep(1)

        print(f"[{self._get_timestamp()}] Publisher stopped. Total messages sent: {self.message_count}")
        if self.redis_client:
            self.redis_client.close()


if __name__ == "__main__":
    publisher = MessagePublisher()
    publisher.run()
