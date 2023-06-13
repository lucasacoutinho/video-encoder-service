# Video Encoder Service

## Bootstrap application
1. ```docker compose up -d```

## Setup Queues

> Consumer Queue
1. The "videos" queue is created automaticatly
> DLX
1. Create an fanout exchage named dlx
2. Attach an queue named "videos-failed" on the exchange
> Result
1. Create an queue named "videos-result" and attach it to the "amq.direct"
