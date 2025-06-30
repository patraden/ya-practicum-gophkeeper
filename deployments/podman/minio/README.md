### minio useful admin commands
```bash
# create alias:
mc alias set local https://localhost:9000 gophkeeper gophkeeper --insecure
# check alias created:
mc alias list
# list buckets
mc ls local --insecure
# display full info
mc admin info --json local
# get redis notify config:
mc admin config get local notify_redis
# apply config:
mc admin config set local notify_redis enable=on format=namespace address=localhost:6379 key=minioevents region=eu-central-1
# confirm event notifications setup:
mc event list local/mysecondbucket arn:minio:sqs:eu-central-1:gophkeeper:redis --insecure
```