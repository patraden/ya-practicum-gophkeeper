# ya-practicum-gophkeeper
![GoReportCard](https://goreportcard.com/badge/github.com/patraden/ya-practicum-gophkeeper)
[![codecov](https://codecov.io/gh/patraden/ya-practicum-gophkeeper/graph/badge.svg?token=9XQT17LJDH)](https://codecov.io/gh/patraden/ya-practicum-gophkeeper)

Final Diploma project

### Project Overview

### Project Requirements
- https://buf.build/docs/cli/installation/
- https://podman.io/get-started


### Dev enviroment setup
```bash
# create infra pod
make pod
# generate tls certificates.
make podman-build-certgen
make podman-certgen-run
# start server infra components
make podman-postgres-run
make podman-keycloak-run
make podman-minio-run
# start gophkeeper server
make podman-build-server
make podman-server-run
# alternatively you can run server locally
make run-server-local

# client operations:
# unseal server as admin:
./dev/scripts/unseal.sh

# install client app
go run ./client install -p "$(pwd)/.gophkeeper"
# register new user
go run ./client register -u patraden -p password
```

