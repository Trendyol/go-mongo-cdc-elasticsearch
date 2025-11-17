# Testing Guide

Bu dokümanda `go-mongo-cdc-elasticsearch` projesinin testlerini nasıl çalıştıracağınız açıklanmaktadır.

## Test Türleri

### 1. Unit Tests
Bireysel fonksiyonları ve modülleri test eder.

### 2. Integration Tests
MongoDB ve Elasticsearch ile gerçek entegrasyon testleri.

## Integration Testlerini Çalıştırma

### Yöntem 1: Docker Compose ile (go-dcp-elasticsearch benzeri) ⭐

Bu yöntem `go-dcp-elasticsearch` projesindeki gibi çalışır.

```bash
# Proje kök dizininde
cd /Users/mehmet.alak/go/src/opensource/go-mongo-cdc-elasticsearch

# Tüm servisleri başlat ve testleri çalıştır
make compose
```

Bu komut:
1. MongoDB custom image'ını build eder
2. Elasticsearch custom image'ını build eder
3. Integration test image'ını build eder
4. Tüm container'ları başlatır
5. Health check'leri bekler
6. Testleri otomatik çalıştırır

**Avantajları:**
- ✅ Tek komut ile her şey çalışır
- ✅ go-dcp-elasticsearch ile aynı yapı
- ✅ CI/CD için ideal
- ✅ Temiz ve izole ortam

### Yöntem 2: Makefile ile (Adım Adım)

```bash
# Test ortamını başlat
make test-integration-up

# Testleri çalıştır
make test-integration-run

# Logları izle (opsiyonel)
make test-integration-logs

# Ortamı kapat
make test-integration-down
```

**Avantajları:**
- ✅ Daha fazla kontrol
- ✅ Debug için uygun
- ✅ Adım adım çalıştırma

### Yöntem 3: Manuel Docker Compose

```bash
# Servisleri başlat
docker-compose -f test/integration/docker-compose.yml up -d

# Servislerin hazır olmasını bekle
sleep 45

# Testleri çalıştır
cd test/integration
go test -v -timeout 10m

# Servisleri kapat
cd ../..
docker-compose -f test/integration/docker-compose.yml down
```

## Karşılaştırma: İki Yöntem

### go-dcp-elasticsearch Tarzı (Önerilen)

```bash
# Kök dizinde docker-compose.yml var
make compose
```

**Yapı:**
```
go-mongo-cdc-elasticsearch/
├── docker-compose.yml           # Ana compose file
├── test/
│   ├── mongodb/
│   │   └── Dockerfile
│   ├── elasticsearch/
│   │   └── Dockerfile
│   └── integration/
│       ├── Dockerfile           # Test container'ı
│       └── integration_test.go
```

### Klasik Yöntem

```bash
# test/integration/ dizininde docker-compose.yml var
make test-integration
```

**Yapı:**
```
go-mongo-cdc-elasticsearch/
└── test/
    └── integration/
        ├── docker-compose.yml   # Test-specific compose
        └── integration_test.go
```

## Hangi Yöntemi Kullanmalıyım?

### `make compose` (go-dcp-elasticsearch tarzı)
**Kullan eğer:**
- ✅ go-dcp-elasticsearch ile tutarlılık istiyorsanız
- ✅ CI/CD pipeline'ında çalıştıracaksanız
- ✅ Tek komut ile her şeyi yapmak istiyorsanız
- ✅ Temiz ve izole test ortamı istiyorsanız

### `make test-integration` (Klasik)
**Kullan eğer:**
- ✅ Test ortamını uzun süre açık tutacaksanız
- ✅ Debug yapmak istiyorsanız
- ✅ Testleri tekrar tekrar çalıştıracaksanız
- ✅ Servislere manuel erişmek istiyorsanız

## Test Senaryoları

### Tüm Testleri Çalıştır

```bash
# Yöntem 1: Docker Compose
make compose

# Yöntem 2: Makefile
make test-integration

# Yöntem 3: Manuel
cd test/integration
go test -v -timeout 10m
```

### Belirli Bir Testi Çalıştır

```bash
# Önce test ortamını başlat
make test-integration-up

# Belirli testi çalıştır
cd test/integration
go test -v -run TestIntegration_BasicInsertOperation

# Ortamı kapat
cd ../..
make test-integration-down
```

### Mevcut Testler

1. **TestIntegration_BasicInsertOperation**
   - MongoDB'ye insert → Elasticsearch'e senkronizasyon

2. **TestIntegration_MultipleInserts**
   - 10 doküman toplu insert

3. **TestIntegration_UpdateOperation**
   - MongoDB'de update → Elasticsearch'te update

4. **TestIntegration_DeleteOperation**
   - MongoDB'den delete → Elasticsearch'ten delete

5. **TestIntegration_CustomMapper**
   - Custom mapper fonksiyonu testi

## Servis Kontrolü

### MongoDB

```bash
# Container'a bağlan
docker exec -it mongodb-router mongosh

# MongoDB shell'de
show dbs
use testdb
db.testcollection.find()
```

### Elasticsearch

```bash
# Health check
curl http://localhost:9200/_cluster/health?pretty

# Index'leri listele
curl http://localhost:9200/_cat/indices?v

# Dokümanları say
curl http://localhost:9200/test-index/_count

# Dokümanları listele
curl http://localhost:9200/test-index/_search?pretty
```

## Logları İzleme

### Tüm Servislerin Logları

```bash
# docker-compose kullanıyorsanız
docker compose logs -f

# VEYA test/integration/docker-compose.yml kullanıyorsanız
make test-integration-logs
```

### Belirli Bir Servisin Logları

```bash
# MongoDB
docker compose logs -f mongodb-router

# Elasticsearch
docker compose logs -f elasticsearch

# Integration Test
docker compose logs -f integration-test
```

## Sorun Giderme

### Container'lar Başlamıyor

```bash
# Container durumunu kontrol et
docker ps -a

# Logları kontrol et
docker compose logs

# Yeniden başlat
docker compose down -v
docker compose up --build
```

### Port Çakışması

```bash
# Çalışan container'ları kontrol et
docker ps

# Portları kontrol et
lsof -i :27017  # MongoDB
lsof -i :9200   # Elasticsearch

# Eski container'ları temizle
docker compose down -v
```

### Testler Timeout Oluyor

```bash
# Daha uzun timeout
go test -v -timeout 20m

# Veya docker-compose'da wait süresini artır
# healthcheck retries değerini artır
```

### Image Build Hataları

```bash
# Cache'i temizle ve yeniden build et
docker compose build --no-cache

# Veya
docker system prune -a
docker compose up --build
```

## Temizlik

### Tüm Container'ları ve Volume'ları Temizle

```bash
# docker-compose kullanıyorsanız
docker compose down -v

# VEYA
make test-integration-clean
```

### Docker Sistem Temizliği

```bash
# Kullanılmayan image'ları temizle
docker image prune -a

# Tüm sistemi temizle (DİKKAT!)
docker system prune -a --volumes
```

## CI/CD Entegrasyonu

### GitHub Actions Örneği

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Run Integration Tests
        run: make compose
```

### GitLab CI Örneği

```yaml
integration-test:
  stage: test
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker compose up --wait --build
  after_script:
    - docker compose down -v
```

## Performans İpuçları

1. **İlk Build**: 5-10 dakika sürebilir (image'lar build edilir)
2. **Sonraki Build'ler**: Cache sayesinde 1-2 dakika
3. **Test Süresi**: ~5-10 dakika (tüm testler)
4. **RAM Kullanımı**: ~2-3 GB
5. **Disk Kullanımı**: ~2 GB

## Önerilen Workflow

### Geliştirme Sırasında

```bash
# 1. Test ortamını başlat (bir kez)
make test-integration-up

# 2. Kod değişikliği yap

# 3. Testleri çalıştır (tekrar tekrar)
cd test/integration
go test -v -run TestIntegration_YourTest

# 4. Bitince ortamı kapat
cd ../..
make test-integration-down
```

### CI/CD'de

```bash
# Tek komut - her şeyi yapar
make compose
```

## Ek Kaynaklar

- [Integration Test README](test/integration/README.md)
- [MongoDB Test Infrastructure](test/mongodb/README.md)
- [Elasticsearch Test Infrastructure](test/elasticsearch/README.md)
- [Quick Start Guide](test/integration/QUICKSTART.md)

## Yardım

```bash
# Tüm make komutlarını gör
make help

# Veya Makefile'ı oku
cat Makefile
```

