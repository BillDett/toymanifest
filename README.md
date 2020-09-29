# Toy Manifest Service

## Pushing layer blobs
```
$ curl -X POST \
    -H "Content-Type: application/vnd.oci.image.layer.v1.tar+gzip" \
    http://localhost:8080/upload \
    --data-binary @1-sha256-534a5505201da9ddb334b5b2fcb3cec45fcafccd8e91b93ad4852e1a1bb318c1.tar.gz
```

## Pulling layer blobs
```
curl http://localhost:8080/layer/65cdcb6a68363de5d6f33582c9a00a70b5503368ff1913ce677fd63648689340 --output test
```

