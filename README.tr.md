# micup.go

`micup.go`, Micup botları için bağımsız bir Go istemcisidir. REST ve Gateway adreslerini keşfeder, bot isteklerini doğrular, Gateway oturumlarını bağlı tutup kaldığı yerden sürdürür, birden çok shard'ı tek süreçte çalıştırır, sunucu kanallarını ve ses durumlarını önbellekler, ses kanallarına katılır, embed içeren mesajlar gönderir, etkileşimlere özel yanıt verebilir ve PCM sesini LiveKit odasına yayımlar.

Modül Micup taşımasını uygulama mantığından ayırır. Müzik arama ve kuyruğu, komut sözdizimi, veritabanı ve worker yönetimi kapsam dışındadır.

## Gereksinimler

- Go 1.26.8 veya yenisi
- Micup bot token’ı
- Ses yayını için `libopus` destekli FFmpeg

```sh
go get github.com/yestefir/micup.go@v0.4.0
```

## Başlangıç

```go
import micup "github.com/yestefir/micup.go"

b, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	return err
}

s, err := b.NewSession(micup.SessionOptions{Handler: func(ctx context.Context, event micup.Event) error {
	if event.Name != micup.EventMessageCreate {
		return nil
	}
	var message micup.MessageCreate
	if err := event.Decode(&message); err != nil {
		return err
	}
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
}})
if err != nil {
	return err
}
err = s.Run(ctx)
```

Çok sayıda sunucuda çalışan botlar `b.NewSession` yerine `b.NewShardManager(micup.ShardOptions{ShardCount: 4, Session: options})` kullanır. Yönetici tüm shard'ları tek süreçte çalıştırır, IDENTIFY isteklerini sıraya koyar ve guild'leri doğru shard'a yönlendirir. Ayrıntılar [Gateway rehberinde](docs/gateway.md#shard-manager).

## Public API

Uygulamalar yalnızca iki yolu import eder:

| Import | İçerik |
|---|---|
| `github.com/yestefir/micup.go` | Her şey: bot kurulumu, REST, Gateway olayları, session, shard, önbellek, veri tipleri, hatalar |
| `github.com/yestefir/micup.go/voice` | LiveKit ve PCM→Opus yayını (ses kullanmayan botlar LiveKit/WebRTC derlemesin diye ayrı) |

İkisi de `src/` üzerinde ince bir cephedir. `src/` içindeki dosyalar taşınsa veya bölünse de bu import yolları ve isimler değişmez; cephe kendi mantığını içerirse bir test başarısız olur.

## Klasör yapısı

Uygulama kodu `src/` altında, her domain kendi paketinde:

```text
micup.go       Public cephe: micup paketi
voice/         Public cephe: voice paketi
src/
  bot/         Giriş noktası: keşif, bot kimliği, session ve shard bağlantıları
  discovery/   /.well-known keşfi ve origin doğrulaması
  rest/        Kimlik doğrulamalı REST: kanallar, mesajlar, etkileşimler, kullanıcılar
  gateway/     Opcode'lar, kapanma kodları, olaylar, tek bağlantı döngüsü
  state/       Guild kanal ve ses durumu önbelleği
  session/     Tek shard için yeniden bağlanan oturum ve ses katılımı
  shard/       Guild→shard eşlemesi ve çoklu shard yöneticisi
  model/       Saf Micup veri tipleri
  voice/       LiveKit odası ve PCM→Opus yayını
  internal/    Sürüm, HTTP güvenliği, test fixture'ları, katman ve cephe kuralları
```

Paket bağımlılıkları tek yönlüdür ve bir testle korunur ([Mimari](docs/architecture.md)). v0.3'ten geçiş için [geçiş rehberine](docs/migration.md) bakın.

Değişiklik göndermeden önce `make check` çalıştırın; CI aynı kontrolleri Linux, macOS ve Windows'ta yapar. Go kaynaklarında yorum satırı kullanılmaz; açıklamalar Markdown belgelerindedir.

## Lisans

Proje MIT lisansıyla ücretsiz kullanılabilir. Ticari kullanım, değiştirme ve yeniden dağıtım serbesttir; kopyalarda telif ve lisans bildirimi korunmalıdır. Ayrıntılar [LICENSE](LICENSE) dosyasındadır.

Güncel sürüm `v0.4.0`'dır. Sürüm notları [CHANGELOG.md](CHANGELOG.md) dosyasındadır.
