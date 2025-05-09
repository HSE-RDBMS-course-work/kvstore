# kvstore

Распределённое ключ-значение хранилище,
использующее алгоритм консенсуса **Raft**
и протокол **gRPC** для взаимодействия.

---

## 🚀 Запуск в Docker

```bash
docker run -d \
  -e KVSTORE_USERNAME=admin \
  -e KVSTORE_PASSWORD=password \
  -p 8090:8090 \
  ghcr.io/hse-rdbms-course-work/kvstore:latest
```

---

## 🧩 Запуск распределённого хранилища в docker-compose

```yaml
services:
  node1:
    container_name: node1
    image: ghcr.io/hse-rdbms-course-work/kvstore:latest
    environment:
      KVSTORE_HOST: 0.0.0.0
      KVSTORE_PUBLIC_PORT: 8090
      KVSTORE_INTERNAL_PORT: 3000
      KVSTORE_USERNAME: admin
      KVSTORE_PASSWORD: password
      KVSTORE_ADVERTISED_ADDRESS: node1:3000
    volumes:
      - node1_data:/home/appuser/data
    ports:
      - "8090:8090"
    healthcheck:
      test: ["CMD", "grpc-health-probe", "-addr=localhost:8090", "-connect-timeout", "250ms", "-rpc-timeout", "100ms"]
      interval: 30s
      retries: 3
      timeout: 10s
      start_period: 2s

  node2:
    container_name: node2
    image: ghcr.io/hse-rdbms-course-work/kvstore:latest
    environment:
      KVSTORE_HOST: 0.0.0.0
      KVSTORE_PUBLIC_PORT: 8090
      KVSTORE_INTERNAL_PORT: 3000
      KVSTORE_USERNAME: admin
      KVSTORE_PASSWORD: password
      KVSTORE_ADVERTISED_ADDRESS: node2:3000
    volumes:
      - node2_data:/home/appuser/data
    ports:
      - "8091:8090"
    depends_on:
      node1:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "grpc-health-probe", "-addr=localhost:8090", "-connect-timeout", "250ms", "-rpc-timeout", "100ms"]
      interval: 30s
      retries: 3
      timeout: 10s
      start_period: 5s
    command: ["-join-to", "node1:8090"]

  node3:
    container_name: node3
    image: ghcr.io/hse-rdbms-course-work/kvstore:latest
    environment:
      KVSTORE_HOST: 0.0.0.0
      KVSTORE_PUBLIC_PORT: 8090
      KVSTORE_INTERNAL_PORT: 3000
      KVSTORE_USERNAME: admin
      KVSTORE_PASSWORD: password
      KVSTORE_ADVERTISED_ADDRESS: node3:3000
    volumes:
      - node3_data:/home/appuser/data
    ports:
      - "8092:8090"
    depends_on:
      node1:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "grpc-health-probe", "-addr=localhost:8090", "-connect-timeout", "250ms", "-rpc-timeout", "100ms"]
      interval: 30s
      retries: 3
      timeout: 10s
      start_period: 2s
    command: ["-join-to", "node1:8090"]

volumes:
  node1_data:
  node2_data:
  node3_data:
```
