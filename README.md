
This is a tiny two-way Telegram-bot to XMPP-bridge. I has worked well for me
for talking to some people that prefer Telegram.

You need to

- cp quiteabot.yaml.example quiteabot.yaml
- Fill out the bot's XMPP account (`xmppserver`, `xmppuser`, `xmpppass`)
- Set `xmpptarget` to your own jid, or so
- Get a Telegram account (mobile phone number needed)
- Register a Telegram bot: https://core.telegram.org/bots
- Fill in authorization token (`telegramtoken`)
- Map some Telegram userids to names in the config (actually conversation ids,
  but I don't think they change) (`telegramusers`)

Telegram users need to chat up your bot with `/start`, and then message you.
**quiteabot** will forward the message to the configured `xmpptarget`,
resolving the name from the userid, if present in `telegramusers`.

To send a message to a Telegram user, you send it over XMPP to the bot
(`xmppuser`). The message must be prefixed with a `username:` from the mapping.
Yes, the Telegram-users must initiate the conversation; that's how Telegram
bots work.

TODO: Can only message a user by name, so first need to stuff her into
`telegramusers`.
