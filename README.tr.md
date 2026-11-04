# micup.go

`micup.go`, Micup botları için bağımsız bir Go istemcisidir. REST ve Gateway adreslerini keşfeder, bot isteklerini doğrular, Gateway olaylarını iletir, kanal mesajı gönderir ve PCM sesini LiveKit odasına yayımlar.

Modül Micup taşımasını uygulama mantığından ayırır. Müzik arama ve kuyruğu, komut sözdizimi, veritabanı ve worker yönetimi kapsam dışındadır.

## Gereksinimler

- Go 1.26.8 veya yenisi
- Micup bot token’ı
- Ses yayını için `libopus` destekli FFmpeg

```sh
go get github.com/mitune/micup.go
```

GitHub modül yolu şu an `mitune` sahibi varsayımıyla ayarlandı. Public yayımdan önce gerçek GitHub sahibi kesinleştirilmeli ve `go.mod` ile kurulum örnekleri birlikte güncellenmelidir.

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

İlk sürüm adayı `v0.1.0` olarak belirlenmiştir. Sürüm notları [CHANGELOG.md](CHANGELOG.md) dosyasındadır. Public GitHub adresi kesinleşmeden bu etiket yayımlanmaz.
