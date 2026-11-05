# micup.go

`micup.go`, Micup botları için bağımsız bir Go istemcisidir. REST ve Gateway adreslerini keşfeder, bot isteklerini doğrular, Gateway oturumunu bağlı tutup kaldığı yerden sürdürür, sunucu kanallarını ve ses durumlarını önbellekler, ses kanallarına katılır, embed içeren mesajlar gönderir, etkileşimlere özel yanıt verebilir ve PCM sesini LiveKit odasına yayımlar.

Modül Micup taşımasını uygulama mantığından ayırır. Müzik arama ve kuyruğu, komut sözdizimi, veritabanı ve worker yönetimi kapsam dışındadır.

## Gereksinimler

- Go 1.26.8 veya yenisi
- Micup bot token’ı
- Ses yayını için `libopus` destekli FFmpeg

```sh
go get github.com/yestefir/micup.go@v0.3.0
```

Public depo [github.com/yestefir/micup.go](https://github.com/yestefir/micup.go); `go.mod` modül yolu bu adresle eşleşir.

## Başlangıç

```go
client, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	return err
}

session := client.NewSession(micup.SessionOptions{Handler: func(ctx context.Context, event micup.Event) error {
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
	_, err := client.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
}})
err = session.Run(ctx)
```

Çok sayıda sunucuda çalışan botlar `client.NewSession` yerine `client.NewShardManager(micup.ShardManagerOptions{ShardCount: 4, Session: options})` kullanabilir. Bu yönetici tüm shard'ları tek süreçte çalıştırır, IDENTIFY isteklerini sıraya koyar ve guild'leri doğru shard'a yönlendirir. Ayrıntılar [Gateway rehberinde](docs/gateway.md#shard-manager).

Değişiklik göndermeden önce `make check` çalıştırın; CI aynı kontrolleri Linux, macOS ve Windows'ta yapar.

REST, Gateway, LiveKit sesi, güvenlik ve katkı rehberleri için [dokümantasyona](docs/README.md) bakın.

Go kaynaklarında yorum satırı kullanılmaz. API örnekleri ve kullanım açıklamaları Markdown belgelerinde tutulur.

## Lisans

Proje MIT lisansıyla ücretsiz kullanılabilir. Ticari kullanım, değiştirme ve yeniden dağıtım serbesttir; kopyalarda telif ve lisans bildirimi korunmalıdır. Ayrıntılar [LICENSE](LICENSE) dosyasındadır.

Güncel sürüm `v0.3.0`'dır. Sürüm notları [CHANGELOG.md](CHANGELOG.md) dosyasındadır.
