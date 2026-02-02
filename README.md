# Search Engine Service

Farklı sağlayıcılardan gelen içerikleri tek bir formata dönüştürerek, belirli bir puanlama algoritmasına göre sıralayan ve REST API üzerinden sunabilme yeteneğine sahip servis.

---

## 🎯 Kapsam

- 2 farklı provider (JSON & XML)
- İçerik arama ve filtreleme
- Popülerlik ve alakalılık skoruna göre sıralama
- Sayfalama (pagination)
- Basit ve genişletilebilir puanlama algoritması
- Cache mekanizması
- Temiz ve okunabilir kod yapısı

---

## 🧱 Mimari Yaklaşım

Proje Clean Architecture prensiplerine uygun olarak katmanlı şekilde tasarlanmıştır.

- Presentation Layer  
  HTTP handler’lar ve request/response yönetimi

- Application Layer  
  İş kuralları, servisler ve use-case’ler

- Domain Layer  
  Core modeller ve iş kuralları

- Infrastructure Layer  
  Provider entegrasyonları, veritabanı ve cache

Bu yapı sayesinde yeni provider eklemek veya iş kurallarını değiştirmek basit olmuştur.

---

## 🔌 Provider Entegrasyonu

### Provider 1
- Format: JSON
- İçerik Tipleri: video, text

### Provider 2
- Format: XML
- İçerik Tipleri: video, article (text olarak normalize edilir)

### Provider Mekanizması

Her provider şu şekilde çalışır:

1. **Veri Çekme**: HTTP üzerinden ilgili API'den veri çekilir
2. **Parse & Transform**: Provider-specific format (JSON/XML) ortak domain modeline dönüştürülür
3. **Validasyon**: Gelen veri domain kurallarına göre validate edilir
4. **Normalizasyon**: Farklı provider'lardan gelen benzer veri tipleri standartlaştırılır (örn: article → text)
5. **Puanlama**: Her içerik için skor hesaplanır
6. **Persistence**: Veriler async olarak PostgreSQL'e kaydedilir

Yeni bir provider eklemek için:
- `ContentProvider` interface'ini implemente edin
- Provider factory'ye kaydedin (`provider.Register`)
- Config dosyasına provider bilgilerini ekleyin

---

## 🛡️ Circuit Breaker Mekanizması

Her provider circuit breaker pattern ile korunur. Bu sayede sorunlu provider'lar sistem genelini etkilemez.

### Durumlar

**Closed (Normal)**
- Tüm istekler provider'a iletilir
- Başarısız istek sayısı threshold değerine ulaşırsa → Open

**Open (Devre Açık)**
- Provider'a istek gönderilmez, doğrudan hata döner
- Timeout süresi sonunda → Half-Open

**Half-Open (Test Modu)**
- Bir deneme isteği gönderilir
- Başarılı olursa → Closed
- Başarısız olursa → Open

### Fallback Stratejisi

Provider başarısız olduğunda:

1. Circuit breaker devreye girer
2. Sistem otomatik olarak PostgreSQL'e fallback yapar
3. Veritabanından ilgili provider'ın son verisi servis edilir
4. Kullanıcı kesintisiz hizmet alır

**Örnek Akış:**
```
Provider1 → Circuit Open → Fallback to Database → Serve Cached Data
Provider2 → Circuit Closed → Fetch from API → Serve Fresh Data
```

### Yapılandırma

Circuit breaker parametreleri [config/config.yaml](config/config.yaml) dosyasında ayarlanabilir:

```yaml
providers:
  - name: provider1
    circuit_breaker:
      threshold: 5        # 5 hata sonrası devre açılır
      timeout: 30s        # 30 saniye sonra tekrar denenir
```

---

## 🔒 Rate Limiting

API istekleri IP bazlı rate limiter ile korunur. Bu sayede aşırı kullanım ve kötüye kullanım önlenir.

### Nasıl Çalışır?

- Her IP adresi için belirli bir zaman penceresi içinde maksimum istek sayısı sınırlanır
- Limit aşıldığında `429 Too Many Requests` hatası döner
- Rate limit parametreleri [config/config.yaml](config/config.yaml) dosyasında yapılandırılır

### Yapılandırma

```yaml
provider:
  rate_limit_max: 100        # Maksimum istek sayısı
  rate_limit_window: 1m      # Zaman penceresi (örn: 1m, 60s)
```

### Limit Aşıldığında Dönen Yanıt

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later."
  },
  "meta": {
    "request_id": "..."
  }
}
```

---

## 🚀 API

### İçerik Arama Endpoint'i

**Endpoint:** `POST /api/v1/search`

**Request:**
```json
{
  "query": "docker",
  "tags": ["devops"],
  "types": ["video", "text"],
  "orderBy": "relevant_score",
  "page": 1,
  "perPage": 20
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "external_id": "v1",
        "provider": "provider1",
        "title": "Introduction to Docker",
        "type": "video",
        "published_at": "2024-03-15T00:00:00Z",
        "views": 22000,
        "likes": 1800,
        "tags": ["devops", "containers"],
        "score": 60.82
      }
    ]
  },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

**Özellikler:**
- Anahtar kelimeye göre arama (title içinde)
- Tag bazlı filtreleme
- İçerik türüne göre filtreleme (video, text)
- Skora göre sıralama (relevant_score / published_at)
- Sayfalama desteği

---

## 🗄️ Veri Saklama & Cache

### PostgreSQL
- Tüm içerikler PostgreSQL'e async olarak persist edilir
- Circuit breaker fallback senaryolarında database'den servis yapılır

### Redis Cache
- Arama sonuçları Redis ile cache'lenir
- Cache TTL: 5 dakika
- **Cache Key Stratejisi**: Request parametreleri (query, tags, types, sort, page, perPage) MD5 hash'lenerek unique cache key oluşturulur
  ```
  Format: search:{md5_hash}
  Örnek: search:5d41402abc4b2a76b9719d911017c592
  ```
- Aynı parametrelerle yapılan aramalar cache'den anında servis edilir

---

## 🛠️ Teknolojiler

- Backend: Go (Fiber)
- Database: PostgreSQL
- Cache: Redis
- SQL Layer: SQLC

---

## ▶️ Çalıştırma

### Gereksinimler

- Go 1.21+
- Docker
- Docker Compose
- Make

### Kurulum ve Çalıştırma

```bash
make docker-up
make migrate
make run
```

Uygulama:

```
http://localhost:8080
```

### Servisleri Durdurma

```bash
make docker-down
```

---

## 🧪 Testler

- Puanlama algoritması için unit testler yazılmıştır
