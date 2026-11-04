# micup.go

`micup.go`, Micup botları için bağımsız bir Go istemcisidir. REST ve Gateway adreslerini keşfeder, bot isteklerini doğrular, Gateway olaylarını iletir, embed içeren mesajlar gönderir, etkileşimlere özel yanıt verebilir ve PCM sesini LiveKit odasına yayımlar.

Modül Micup taşımasını uygulama mantığından ayırır. Müzik arama ve kuyruğu, komut sözdizimi, veritabanı ve worker yönetimi kapsam dışındadır.

## Gereksinimler

- Go 1.26.8 veya yenisi
- Micup bot token’ı
- Ses yayını için `libopus` destekli FFmpeg

```sh
go get github.com/yestefir/micup.go@v0.2.0
```

Public depo [github.com/yestefir/micup.go](https://github.com/yestefir/micup.go); `go.mod` modül yolu bu adresle eşleşir.

## Başlangıç

```go
client, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
if err != nil {
	return err
}

err = client.Gateway(ctx, func(ctx context.Context, event micup.Event) error {
	if event.Name != "MESSAGE_CREATE" {
		return nil
	}
	var message micup.MessageCreate
	if err := json.Unmarshal(event.Data, &message); err != nil {
		return err
	}
	if message.Author.Bot || message.Content != "!ping" {
		return nil
	}
	_, err := client.SendMessage(ctx, message.ChannelID, "pong", message.ID)
	return err
})
```

REST, Gateway, LiveKit sesi, güvenlik ve katkı rehberleri için [dokümantasyona](docs/README.md) bakın.

Go kaynaklarında yorum satırı kullanılmaz. API örnekleri ve kullanım açıklamaları Markdown belgelerinde tutulur.

## Lisans

Proje MIT lisansıyla ücretsiz kullanılabilir. Ticari kullanım, değiştirme ve yeniden dağıtım serbesttir; kopyalarda telif ve lisans bildirimi korunmalıdır. Ayrıntılar [LICENSE](LICENSE) dosyasındadır.

Güncel public sürüm `v0.2.0`'dır. Sürüm notları [CHANGELOG.md](CHANGELOG.md) dosyasındadır.
