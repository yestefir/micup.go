# micup.go

`micup.go`, Micup botları için bağımsız bir Go istemcisidir ve Micup'ın üzerinde çalıştığı Fluxer API'sini izler:

- **Gateway:** kaldığı yerden süren, işlenmemiş olayları yeniden alan ve aynı soketten yeniden IDENTIFY yapan oturumlar; tek süreçte çok shard; isteğe bağlı zstd-stream sıkıştırma; yok sayılan olaylar ve durum; Micup sınırlarının altında komut hızı.
- **Önbellek:** guild'ler, izin geçersiz kılmalı kanallar, roller, üye rolleri, ses durumları ve izin hesabı.
- **REST:** mesajlar (gönderme, düzenleme, silme, toplu silme, listeleme), butonlar ve tıklama yanıtları, tepkiler, sabitleme, yazıyor göstergesi, DM, dosyalar (multipart ve önceden yükleme), üyeler, roller, moderasyon ve webhook'lar; Micup hata kodları ve rate limit kurallarıyla.
- **Ses:** hızlı ret ve izin kontrolüyle katılma, yetki yenileme ve LiveKit'e Opus yayını.

Modül Micup taşımasını uygulama mantığından ayırır. Müzik arama ve kuyruğu, komut sözdizimi, veritabanı ve worker yönetimi kapsam dışındadır.

## Gereksinimler

- Go 1.26.8 veya yenisi
- Micup bot token’ı
- Ses yayını için `libopus` destekli FFmpeg

```sh
go get github.com/yestefir/micup.go@v0.6.0
```

## Başlangıç

```go
import micup "github.com/yestefir/micup.go"

b, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	return err
}

mux := micup.NewEventMux()
mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, message micup.MessageCreate) error {
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := b.REST.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
}))

s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
if err != nil {
	return err
}
err = s.Run(ctx)
```

Her olayın tipli bir veri yapısı vardır; liste ve olay yönlendiricisi için [Olaylar ve tipli veriler](docs/events.md). Slash ve sağ tık komutları REST ile kaydedilir; bkz. [Komutlar](docs/commands.md).

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
events.go      Public cephe: olay adları, veri tipleri, olay yönlendiricisi
voice/         Public cephe: voice paketi
src/
  bot/         Giriş noktası: keşif, bot kimliği, session ve shard bağlantıları
  discovery/   /.well-known keşfi ve origin doğrulaması
  rest/        Kimlik doğrulamalı REST: kanallar, mesajlar, etkileşimler, kullanıcılar
  gateway/     Opcode'lar, kapanma kodları, olay adları, tek bağlantı döngüsü
  events/      Her Gateway olayının veri tipi, kayıt tablosu, tipli çözme, olay yönlendiricisi
  state/       Guild kanal ve ses durumu önbelleği
  session/     Tek shard için yeniden bağlanan oturum ve ses katılımı
  shard/       Guild→shard eşlemesi ve çoklu shard yöneticisi
  model/       REST ve Gateway'in ortak domain tipleri
  voice/       LiveKit odası ve PCM→Opus yayını
  internal/    Sürüm, HTTP güvenliği, test fixture'ları, katman ve cephe kuralları
```

Paket bağımlılıkları tek yönlüdür ve bir testle korunur ([Mimari](docs/architecture.md)). v0.3'ten geçiş için [geçiş rehberine](docs/migration.md) bakın.

Değişiklik göndermeden önce `make check` çalıştırın; CI aynı kontrolleri Linux, macOS ve Windows'ta yapar. Go kaynaklarında yorum satırı kullanılmaz; açıklamalar Markdown belgelerindedir.

## Lisans

Proje MIT lisansıyla ücretsiz kullanılabilir. Ticari kullanım, değiştirme ve yeniden dağıtım serbesttir; kopyalarda telif ve lisans bildirimi korunmalıdır. Ayrıntılar [LICENSE](LICENSE) dosyasındadır.

Güncel sürüm `v0.6.0`'dır. Sürüm notları [CHANGELOG.md](CHANGELOG.md) dosyasındadır.

`SessionOptions.TrackMembers` / `Session.MemberInventories()`: [maintained member inventories](docs/guilds.md#maintained-member-inventories).
