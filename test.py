from types import resolve_bases
import requests


def stream_test():
    url = "http://localhost:8080/chat"

    with requests.post(
        url,
        stream=True,
        headers={"Content-Type": "text/event-stream"},
        json={"someshit": 1},
    ) as response:
        for line in response.iter_lines():
            if line:
                print(f"Received message: {line.decode('utf-8')}")


def non_stream_test():
    url = "http://localhost:8080/chat"

    with requests.post(url) as response:

        print(f"Received message: {response.content}")


if __name__ == "__main__":
    stream_test()
    non_stream_test()
