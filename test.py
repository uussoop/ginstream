import requests


def stream_listener():
    url = "http://localhost:8080/stream"

    with requests.get(url, stream=True) as response:
        for line in response.iter_lines():
            if line:
                print(f"Received message: {line.decode('utf-8')}")


if __name__ == "__main__":
    stream_listener()
