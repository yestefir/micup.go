// Package micup is a standalone Go client for Micup bots, following the
// Fluxer API that Micup runs on.
//
// One import gives you the Gateway, a state cache, the REST client, an
// interaction router and the typed event payloads. Voice lives in the
// separate github.com/yestefir/micup.go/voice package so bots without voice
// do not pull in LiveKit and WebRTC.
//
// micup.go keeps platform transport separate from application behavior. It
// does not implement music search, queues, command syntax, persistence or
// worker coordination; those stay with your application. mitune.net, one of
// the largest music bots on Micup, runs on it.
//
// # Quick start
//
// Create a [Bot] from a token, route events through an [EventMux] and run a
// [Session]:
//
//	b, err := micup.New(ctx, micup.Config{Token: os.Getenv("MICUP_BOT_TOKEN")})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	mux := micup.NewEventMux()
//	mux.On(micup.EventMessageCreate, micup.Listen(func(ctx context.Context, m micup.MessageCreate) error {
//		if m.Author.Bot || m.Content != "!ping" {
//			return nil
//		}
//		_, err := b.REST.SendMessage(ctx, m.ChannelID, "pong", m.ID)
//		return err
//	}))
//
//	s, err := b.NewSession(micup.SessionOptions{Handler: mux.Handle})
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = s.Run(ctx)
//
// # Gateway
//
// A [Session] owns one Gateway connection. It resumes after a disconnect,
// replays unprocessed events, re-identifies on the same socket, can use zstd
// stream compression, accepts ignored events and presence, and paces commands
// below Micup's limits. Gateway commands cover member requests, guild and
// channel member counts, Lazy Request subscriptions and voice state updates.
//
// Bots in many guilds use a shard manager instead of a single session:
//
//	m, err := b.NewShardManager(micup.ShardOptions{ShardCount: 4, Session: options})
//
// Every Micup dispatch event, including threads, calls, emojis, stickers and
// member lists, has a typed payload. [Listen] decodes the payload for your
// handler, so you never unmarshal JSON yourself.
//
// # State
//
// The state cache tracks guilds with their settings, emojis and stickers,
// channels and threads with permission overwrites, roles, member roles and
// voice states. It answers permission and role hierarchy questions, and
// SessionOptions.TrackMembers keeps maintained member inventories.
//
// # REST
//
// The REST client covers messages (send, edit, forward, crosspost, delete,
// bulk delete, list and search), buttons, select menus, modals and commands,
// reactions, pins, typing, DMs, files (multipart and presigned uploads),
// channels, threads and forums, roles, guild settings, members, invites,
// emojis, stickers, audit logs, bans, moderation and webhooks. It follows
// Micup's rate limit rules, returns Micup's error codes as typed errors and
// offers paging iterators for long lists.
//
// # Slash commands and components
//
// Register commands over REST and answer them with an [EventRouter], which
// routes commands, autocomplete, components and modals by name or pattern:
//
//	router := micup.NewEventRouter()
//	router.Command("queue add", addToQueue)
//	router.Component("page:*", changePage)
//
//	mux := micup.NewEventMux()
//	mux.On(micup.EventInteractionCreate, router.Listener())
//
// A router returns [ErrNoRoute] for interactions it has no handler for.
//
// # Helpers
//
// Helpers cover mentions, timestamps, markdown escaping, an embed builder,
// media URLs, snowflake times and permission names, plus optional log/slog
// logging.
//
// # Voice
//
// Session.JoinVoice performs the join handshake with fast refusal and
// permission checks. The voice package turns the resulting grant into a
// LiveKit room, publishes Opus audio, and offers a player with FFmpeg
// sources, pause, skip and volume. FFmpeg built with libopus is required to
// publish audio.
//
// # Requirements and further reading
//
// micup.go needs Go 1.26.8 or newer and a Micup bot token. The README and the
// guides in the docs directory of the repository describe setup, events,
// messages, components, commands, guilds, channels, formatting, voice,
// architecture, security and migration from older versions. Runnable bots are
// in the examples directory: ping, slash-echo, panel, guild-tag and
// voice-player.
package micup
